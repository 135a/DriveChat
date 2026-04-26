package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nym/go-gateway/pkg/redis"
)

// Metrics keys
const (
	KeyMetricsQPS            = "gateway:metrics:qps"
	KeyMetricsInterceptions = "gateway:metrics:interceptions"
)

// GetMetrics returns real-time gateway performance data.
func GetMetrics(c *gin.Context) {
	qps, _ := redis.Client.Get(redis.Ctx, KeyMetricsQPS).Result()
	interceptions, _ := redis.Client.Get(redis.Ctx, KeyMetricsInterceptions).Result()

	c.JSON(http.StatusOK, gin.H{
		"qps":            qps,
		"interceptions": interceptions,
	})
}
