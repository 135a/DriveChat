package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nym/go-gateway/pkg/redis"
)

// BlacklistMiddleware checks if the client IP or path is in the blacklist.
func BlacklistMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		path := c.Request.URL.Path

		// Check IP blacklist
		isIPBlocked, _ := redis.Client.SIsMember(redis.Ctx, "global:gateway:blacklist:ips", clientIP).Result()
		if isIPBlocked {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied: IP blacklisted"})
			c.Abort()
			return
		}

		// Check Path blacklist (Exact and Prefix)
		blockedPaths, _ := redis.Client.SMembers(redis.Ctx, "global:gateway:blacklist:paths").Result()
		for _, blockedPath := range blockedPaths {
			if path == blockedPath || strings.HasPrefix(path, blockedPath) {
				c.JSON(http.StatusForbidden, gin.H{"error": "Access denied: Path blacklisted"})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}
