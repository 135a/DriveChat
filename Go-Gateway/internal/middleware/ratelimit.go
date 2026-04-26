package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nym/go-gateway/internal/models"
	"github.com/nym/go-gateway/pkg/mysql"
	"github.com/nym/go-gateway/pkg/redis"
)

// RateLimitMiddleware implements dual-threshold rate limiting.
func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		key := fmt.Sprintf("gateway:ratelimit:%s:%d", clientIP, time.Now().Unix()/60) // Per minute

		// Get thresholds from Redis (or use defaults)
		slightlyThreshold := 50
		tooThreshold := 100
		
		val, _ := redis.Client.Get(redis.Ctx, "global:sys_config:slightly_freq_threshold").Int()
		if val > 0 {
			slightlyThreshold = val
		}
		val, _ = redis.Client.Get(redis.Ctx, "global:sys_config:too_freq_threshold").Int()
		if val > 0 {
			tooThreshold = val
		}

		count, err := redis.Client.Incr(redis.Ctx, key).Result()
		if err == nil && count == 1 {
			redis.Client.Expire(redis.Ctx, key, time.Minute)
		}

		// Too Frequent -> Auto Blacklist
		if int(count) > tooThreshold {
			autoBlacklist(clientIP, c.Request.URL.Path)
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied: Auto-blacklisted due to excessive requests"})
			c.Abort()
			return
		}

		// Slightly Frequent -> 429
		if int(count) > slightlyThreshold {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func autoBlacklist(ip, path string) {
	// Add to Redis
	redis.Client.SAdd(redis.Ctx, "global:gateway:blacklist:ips", ip)

	// Persist to MySQL
	rule := models.BlacklistRule{
		Type:   "IP",
		Value:  ip,
		IsAuto: true,
		Status: 1,
	}
	mysql.DB.Create(&rule)

	// Log the block
	log := models.BlockLog{
		ClientIP:  ip,
		ReqPath:   path,
		RuleType:  "RATE_LIMIT_AUTO_BLOCK",
		RuleValue: ip,
	}
	mysql.DB.Create(&log)
}
