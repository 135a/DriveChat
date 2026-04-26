package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nym/go-gateway/pkg/redis"
)

// MetricsMiddleware tracks QPS and response times.
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Before request: Increment QPS
		redis.Client.Incr(redis.Ctx, "gateway:metrics:qps")

		c.Next()

		// After request: Record latency and status codes if needed
		latency := time.Since(start)
		_ = latency

		status := c.Writer.Status()
		if status >= 400 {
			redis.Client.Incr(redis.Ctx, "gateway:metrics:interceptions")
		}
		
		// Record status code count
		redis.Client.HIncrBy(redis.Ctx, "gateway:metrics:status_codes", strconv.Itoa(status), 1)
	}
}
