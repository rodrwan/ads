package session

import (
	"ads-system/internal/database"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SessionHandler struct {
	Querier database.Querier
}

func NewSessionHandler(querier database.Querier) *SessionHandler {
	return &SessionHandler{Querier: querier}
}

func (h *SessionHandler) CreateSession(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
}
