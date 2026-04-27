package middleware

import (
	_ "embed"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nym/go-gateway/internal/models"
	"github.com/nym/go-gateway/pkg/mysql"
	"github.com/nym/go-gateway/pkg/redis"
)

//go:embed lua/sliding_window.lua
var slidingWindowScript string

// RateLimitMiddleware implements dual-threshold rate limiting using Lua-based sliding window.
func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		key := fmt.Sprintf("gateway:ratelimit:%s", clientIP)

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

		now := time.Now().UnixMilli()
		member := fmt.Sprintf("%d:%d", now, rand.Int63())
		windowMs := int64(60000) // 1 minute sliding window

		// Execute Lua script atomically: check against "slightly frequent" threshold
		result, err := redis.Client.Eval(redis.Ctx, slidingWindowScript, []string{key}, windowMs, slightlyThreshold, now, member).Int()
		if err != nil {
			// On Redis error, fail open (allow request)
			c.Next()
			return
		}

		if result == 1 {
			// Check current count to determine if it's "too frequent" (auto-blacklist)
			count, _ := redis.Client.ZCard(redis.Ctx, key).Result()
			if int(count) > tooThreshold {
				autoBlacklist(clientIP, c.Request.URL.Path)
				c.JSON(http.StatusForbidden, gin.H{"error": "Access denied: Auto-blacklisted due to excessive requests"})
				c.Abort()
				return
			}

			// Slightly Frequent -> 429
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
