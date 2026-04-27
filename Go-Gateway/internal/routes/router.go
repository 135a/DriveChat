package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/nym/go-gateway/internal/controllers"
	"github.com/nym/go-gateway/internal/middleware"
)

// InitRoutes configures the route pipeline for the gateway.
func InitRoutes(r *gin.Engine) {
	// 1. Gateway Pipeline (Order matters!)
	// Metrics -> Blacklist -> RateLimit -> Dynamic Proxy
	gateway := r.Group("/")
	{
		gateway.Use(middleware.MetricsMiddleware())
		gateway.Use(middleware.BlacklistMiddleware())
		gateway.Use(middleware.RateLimitMiddleware())
		gateway.Use(middleware.DynamicProxyMiddleware())
	}

	// 2. Management API (no auth - restrict via network policy in production)
	api := r.Group("/api")
	{
		// Route Rules
		api.GET("/routes", controllers.GetRoutes)
		api.POST("/routes", controllers.CreateRoute)
		api.PUT("/routes/:id", controllers.UpdateRoute)
		api.DELETE("/routes/:id", controllers.DeleteRoute)

		// Blacklist
		api.GET("/blacklist", controllers.GetBlacklist)
		api.POST("/blacklist", controllers.CreateBlacklist)
		api.DELETE("/blacklist/:id", controllers.DeleteBlacklist)

		// Logs & Metrics
		api.GET("/logs", controllers.GetBlockLogs)
		api.GET("/metrics", controllers.GetMetrics)
	}
}
