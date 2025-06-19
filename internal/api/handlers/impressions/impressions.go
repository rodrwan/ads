package impressions

import (
	"ads-system/internal/database"
	"ads-system/internal/interfaces"
	"ads-system/internal/logger"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"
)

type ImpressionHandler struct {
	Querier         database.Querier
	RedisClient     *redis.Client
	MetricsNotifier interfaces.MetricsNotifier
}

type ImpressionOption func(*ImpressionHandler)

func WithQuerier(querier database.Querier) ImpressionOption {
	return func(ih *ImpressionHandler) {
		ih.Querier = querier
	}
}

func WithRedisClient(redisClient *redis.Client) ImpressionOption {
	return func(ih *ImpressionHandler) {
		ih.RedisClient = redisClient
	}
}

func WithMetricsNotifier(notifier interfaces.MetricsNotifier) ImpressionOption {
	return func(ih *ImpressionHandler) {
		ih.MetricsNotifier = notifier
	}
}

func NewImpressionHandler(options ...ImpressionOption) *ImpressionHandler {
	ih := &ImpressionHandler{}

	for _, option := range options {
		option(ih)
	}

	return ih
}

type ImpressionRequest struct {
	AdID        string                 `json:"ad_id"`
	PlacementID string                 `json:"placement_id"`
	AuctionID   string                 `json:"auction_id"`
	UserContext map[string]interface{} `json:"user_context"`
}

func (ih *ImpressionHandler) CreateImpression(c *gin.Context) (interface{}, int, error) {
	var req ImpressionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, http.StatusBadRequest, err
	}

	userCtxJson, err := json.Marshal(req.UserContext)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	var id pgtype.UUID
	_ = id.Scan(uuid.New().String())

	_, err = ih.Querier.CreateImpression(c, database.CreateImpressionParams{
		ID:          id,
		AdID:        pgtype.UUID{Bytes: [16]byte(uuid.MustParse(req.AdID))},
		PlacementID: pgtype.UUID{Bytes: [16]byte(uuid.MustParse(req.PlacementID))},
		AuctionID:   pgtype.UUID{Bytes: [16]byte(uuid.MustParse(req.AuctionID))},
		UserContext: userCtxJson,
	})

	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	// Notificar actualización de métricas en tiempo real
	if ih.MetricsNotifier != nil {
		go ih.MetricsNotifier.NotifyMetricsUpdate()
	}

	go func() {
		_, err := ih.RedisClient.XAdd(c.Request.Context(), &redis.XAddArgs{
			Stream: "stream:impressions",
			Values: map[string]interface{}{
				"ad_id":        req.AdID,
				"placement_id": req.PlacementID,
				"auction_id":   req.AuctionID,
				"user_context": string(userCtxJson),
			},
		}).Result()
		if err != nil {
			logger.Error("Error adding impression to Redis", err)
		}
	}()

	return nil, http.StatusOK, nil
}
