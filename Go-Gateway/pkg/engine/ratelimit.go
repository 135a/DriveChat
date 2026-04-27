package engine

import (
	"net/http"
	"sync"
	"time"
)

// TokenBucket 实现了经典的“令牌桶”算法。
// 它通过以固定速率向桶中放入令牌，并要求请求在执行前必须先获取令牌，
// 从而实现平滑限流（Traffic Shaping）和应对突发流量（Burstiness）的能力。
type TokenBucket struct {
	rate         float64    // 每秒生成的令牌速率（QPS）
	capacity     float64    // 桶的最大容量，决定了允许的最大突发流量
	tokens       float64    // 当前桶中剩余的令牌数量
	lastRefillAt time.Time  // 上次填充令牌的时间戳
	mu           sync.Mutex // 保护桶状态的并发安全锁
}

// NewTokenBucket 创建一个新的令牌桶实例。
// qps: 每秒允许的请求数，burst: 允许的最大突发请求量。
func NewTokenBucket(qps, burst float64) *TokenBucket {
	return &TokenBucket{
		rate:         qps,
		capacity:     burst,
		tokens:       burst, // 初始时桶是满的
		lastRefillAt: time.Now(),
	}
}

// Allow 尝试从桶中获取一个令牌。
// 如果获取成功（桶中有令牌），返回 true；
// 如果获取失败（桶已空），返回 false。
func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// 1. 根据经过的时间计算应该补充的令牌。
	now := time.Now()
	elapsed := now.Sub(tb.lastRefillAt).Seconds()
	
	// 补充令牌：当前令牌 = min(桶容量, 当前令牌 + 经过秒数 * 生成速率)
	tb.tokens += elapsed * tb.rate
	if tb.tokens > tb.capacity {
		tb.tokens = tb.capacity
	}
	tb.lastRefillAt = now

	// 2. 检查是否有足够的令牌。
	if tb.tokens >= 1.0 {
		tb.tokens -= 1.0
		return true
	}

	return false
}

// RateLimitMiddleware 是网关的限流中间件。
// 它利用令牌桶算法对请求进行拦截。如果超过设定的 QPS，将直接返回 429 Too Many Requests。
func RateLimitMiddleware(qps, burst float64) HandlerFunc {
	// 为该中间件实例创建一个私有的令牌桶。
	// 注意：在多路由场景下，我们可以为不同路由创建不同的限流器。
	tb := NewTokenBucket(qps, burst)

	return func(c *Context) {
		// 尝试获取令牌。
		if !tb.Allow() {
			// 如果令牌不足，记录状态码并中断后续中间件的执行。
			// 429 状态码表示客户端发送请求过快，触发了服务器的速率限制。
			c.JSON(http.StatusTooManyRequests, map[string]string{
				"error": "请求过于频繁，请稍后再试",
			})
			c.Abort()
			return
		}

		// 令牌获取成功，继续执行后续逻辑。
		c.Next()
	}
}
