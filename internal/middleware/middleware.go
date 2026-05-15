package middleware

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/dolsom/user-service/internal/model"
	"github.com/dolsom/user-service/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// RequireUserID extracts X-User-ID set by the gateway and aborts if missing.
func RequireUserID() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetHeader("X-User-ID")
		if userID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "missing user identity",
			})
			return
		}
		c.Set("user_id", userID)
		c.Next()
	}
}

// RequestLogger logs every request as structured JSON to stdout.
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		log.Info().
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Int("status", c.Writer.Status()).
			Dur("duration_ms", time.Since(start)).
			Str("user_id", c.GetString("user_id")).
			Msg("request")
	}
}

// ActivityLogger logs API requests to the activity_logs table.
func ActivityLogger(activityRepo *repository.ActivityRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		userID := c.GetString("user_id")
		if userID == "" {
			return
		}
		meta, _ := json.Marshal(map[string]any{
			"method": c.Request.Method,
			"path":   c.Request.URL.Path,
			"status": c.Writer.Status(),
		})
		_ = activityRepo.Log(userID, model.ActivityAPIRequest, string(meta))
	}
}
