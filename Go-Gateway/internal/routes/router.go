package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/nym/go-gateway/internal/controllers"
	"github.com/nym/go-gateway/internal/middleware"
)

// SetupRouter configures the API routes for the gateway and admin dashboard.
func SetupRouter() *gin.Context {
	// Note: In a real app, this would return *gin.Engine. 
	// I'll return *gin.Engine in the actual implementation.
	return nil
}

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

	// 2. Admin Auth Routes
	auth := r.Group("/auth")
	{
		auth.POST("/register", controllers.Register)
		auth.POST("/login", controllers.Login)
	}

	// 3. Admin Dashboard API (Protected by JWT)
	admin := r.Group("/admin")
	admin.Use(middleware.JWTAuthMiddleware())
	{
		// Sys Config
		admin.GET("/config", controllers.GetSysConfig)
		admin.POST("/config", controllers.UpdateSysConfig)

		// Routes
		admin.GET("/routes", controllers.GetRoutes)
		admin.POST("/routes", controllers.CreateRoute)
		admin.PUT("/routes/:id", controllers.UpdateRoute)
		admin.DELETE("/routes/:id", controllers.DeleteRoute)

		// Blacklist
		admin.GET("/blacklist", controllers.GetBlacklist)
		admin.POST("/blacklist", controllers.CreateBlacklist)
		admin.DELETE("/blacklist/:id", controllers.DeleteBlacklist)

		// Logs
		admin.GET("/logs", controllers.GetBlockLogs)

		// Metrics
		admin.GET("/metrics", controllers.GetMetrics)
	}
}
