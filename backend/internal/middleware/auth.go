package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/taskflow/backend/internal/service"
)

const (
	ContextKeyUserID = "userID"
	ContextKeyEmail  = "userEmail"
)

func RequireAuth(authService service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   true,
				"message": "Authorization header is required",
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   true,
				"message": "Authorization header format must be 'Bearer <token>'",
			})
			return
		}

		claims, err := authService.ValidateToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   true,
				"message": "Invalid or expired token",
			})
			return
		}

		c.Set(ContextKeyUserID, claims.UserID)
		c.Set(ContextKeyEmail, claims.Email)
		c.Next()
	}
}

func OptionalAuth(authService service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
				claims, err := authService.ValidateToken(parts[1])
				if err == nil {
					c.Set(ContextKeyUserID, claims.UserID)
					c.Set(ContextKeyEmail, claims.Email)
				}
			}
		}
		c.Next()
	}
}

func GetUserID(c *gin.Context) (uuid.UUID, bool) {
	val, exists := c.Get(ContextKeyUserID)
	if !exists {
		return uuid.Nil, false
	}
	id, ok := val.(uuid.UUID)
	return id, ok
}
