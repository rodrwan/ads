package api

import (
	"ads-system/internal/api/handlers/auction"
	"ads-system/internal/api/handlers/clicks"
	"ads-system/internal/api/handlers/impressions"
	"ads-system/internal/database"
	"ads-system/internal/logger"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type Server struct {
	Querier      database.Querier
	RedisClient  *redis.Client
	WebSocketMgr *WebSocketManager
}

func NewServer(querier database.Querier, redisClient *redis.Client) *Server {
	wsManager := NewWebSocketManager(querier)

	// Iniciar el manager de WebSockets en background
	go wsManager.Start()

	// Iniciar el broadcast automático de métricas
	wsManager.StartMetricsBroadcaster()

	return &Server{
		Querier:      querier,
		RedisClient:  redisClient,
		WebSocketMgr: wsManager,
	}
}

func (server *Server) Routes(r *gin.Engine) *gin.Engine {
	auctionHandler := auction.NewAuctionHandler(server.Querier)
	auctionGroup := r.Group("/auction")
	{
		auctionGroup.POST("", server.handle(auctionHandler.CreateAuction))
	}

	impressionHandler := impressions.NewImpressionHandler(
		impressions.WithQuerier(server.Querier),
		impressions.WithRedisClient(server.RedisClient),
		impressions.WithMetricsNotifier(server.WebSocketMgr),
	)
	impressionGroup := r.Group("/impression")
	{
		impressionGroup.POST("", server.handle(impressionHandler.CreateImpression))
	}

	clickHandler := clicks.NewClickHandler(
		clicks.WithQuerier(server.Querier),
		clicks.WithRedisClient(server.RedisClient),
		clicks.WithMetricsNotifier(server.WebSocketMgr),
	)
	clickGroup := r.Group("/click")
	{
		clickGroup.POST("", server.handle(clickHandler.CreateClick))
	}

	// WebSocket endpoint para actualizaciones en tiempo real
	r.GET("/ws", server.WebSocketMgr.WebSocketHandler())

	return r
}

type handlerFunc func(c *gin.Context) (interface{}, int, error)

func (server *Server) handle(handler handlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		data, statusCode, err := handler(c)
		if err != nil {
			logger.WithFields(logger.Fields{
				"status_code": statusCode,
				"error":       err,
			}).WithError(err).Error("Error in handler")

			c.JSON(statusCode, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(statusCode, data)
	}
}
