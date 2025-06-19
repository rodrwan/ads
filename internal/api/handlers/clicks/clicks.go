package clicks

import (
	"ads-system/internal/database"
	"ads-system/internal/interfaces"
	"ads-system/internal/logger"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"
)

type ClickHandler struct {
	Querier         database.Querier
	RedisClient     *redis.Client
	MetricsNotifier interfaces.MetricsNotifier
}

type ClickOption func(*ClickHandler)

func WithQuerier(querier database.Querier) ClickOption {
	return func(ih *ClickHandler) {
		ih.Querier = querier
	}
}

func WithRedisClient(redisClient *redis.Client) ClickOption {
	return func(ih *ClickHandler) {
		ih.RedisClient = redisClient
	}
}

func WithMetricsNotifier(notifier interfaces.MetricsNotifier) ClickOption {
	return func(ih *ClickHandler) {
		ih.MetricsNotifier = notifier
	}
}

func NewClickHandler(options ...ClickOption) *ClickHandler {
	ih := &ClickHandler{}

	for _, option := range options {
		option(ih)
	}

	return ih
}

type ClickRequest struct {
	ImpressionID string `json:"impression_id"`
}

func (ch *ClickHandler) CreateClick(c *gin.Context) (interface{}, int, error) {
	var req ClickRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, http.StatusBadRequest, err
	}

	var id pgtype.UUID
	_ = id.Scan(uuid.New().String())

	_, err := ch.Querier.CreateClick(c, database.CreateClickParams{
		ID:           id,
		ImpressionID: pgtype.UUID{Bytes: [16]byte(uuid.MustParse(req.ImpressionID))},
	})

	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	// Notificar actualización de métricas en tiempo real
	if ch.MetricsNotifier != nil {
		go ch.MetricsNotifier.NotifyMetricsUpdate()
	}

	go func() {
		_, err := ch.RedisClient.XAdd(c.Request.Context(), &redis.XAddArgs{
			Stream: "stream:clicks",
			Values: map[string]interface{}{
				"impression_id": req.ImpressionID,
			},
		}).Result()
		if err != nil {
			logger.Error("Error adding click to Redis", err)
		}
	}()

	return nil, http.StatusOK, nil
}
