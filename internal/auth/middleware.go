package auth

import (
	"net/http"
	"strings"

	"datacenter-calc/internal/model"
	"datacenter-calc/internal/redis"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(redisClient *redis.RedisClient, allowedRoles ...model.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing Authorization header"})
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenStr == authHeader {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid Authorization format"})
			return
		}

		if blacklisted, _ := redisClient.IsBlacklisted(c.Request.Context(), tokenStr); blacklisted {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token revoked"})
			return
		}

		claims, err := ParseToken(tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		hasAccess := false
		for _, role := range allowedRoles {
			if claims.Role == role {
				hasAccess = true
				break
			}
		}
		if !hasAccess {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("userRole", claims.Role)
		c.Next()
	}
}

func GuestOrAuthMiddleware(redisClient *redis.RedisClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Set("userID", 0)
			c.Set("userRole", "")
			c.Next()
			return
		}
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := ParseToken(tokenStr)
		if err != nil {
			c.Set("userID", 0)
			c.Set("userRole", "")
			c.Next()
			return
		}
		c.Set("userID", claims.UserID)
		c.Set("userRole", claims.Role)
		c.Next()
	}
}
