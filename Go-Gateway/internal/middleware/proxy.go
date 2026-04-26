package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nym/go-gateway/internal/services"
	"github.com/nym/go-gateway/pkg/redis"
)

// DynamicProxyMiddleware handles matching the request path to a target URL in Redis.
func DynamicProxyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only proxy if not an admin/config route (handled by other groups)
		if strings.HasPrefix(c.Request.URL.Path, "/admin") || strings.HasPrefix(c.Request.URL.Path, "/auth") {
			c.Next()
			return
		}

		// Get routes from Redis hash
		routes, err := redis.Client.HGetAll(redis.Ctx, "global:gateway:routes").Result()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load routes"})
			c.Abort()
			return
		}

		var targetURL string
		path := c.Request.URL.Path

		// Match path prefix (Longest prefix match would be better, but let's start simple)
		for prefix, target := range routes {
			if strings.HasPrefix(path, prefix) {
				targetURL = target
				break
			}
		}

		if targetURL == "" {
			c.JSON(http.StatusNotFound, gin.H{"error": "No route matched"})
			c.Abort()
			return
		}

		// Perform proxying
		services.ProxyRequest(c, targetURL)
		c.Abort() // Proxy handles the response
	}
}
