package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nym/go-gateway/internal/router"
	"github.com/nym/go-gateway/internal/services"
)

// DynamicProxyMiddleware handles matching the request path to a target URL using the Trie router.
func DynamicProxyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only proxy if not a management API route
		if strings.HasPrefix(c.Request.URL.Path, "/api") {
			c.Next()
			return
		}

		// Match route using in-memory Trie (O(K) complexity)
		route, _ := router.Global.Search(c.Request.URL.Path)

		if route == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "No route matched"})
			c.Abort()
			return
		}

		// Perform proxying
		services.ProxyRequest(c, route.TargetURL)
		c.Abort() // Proxy handles the response
	}
}
