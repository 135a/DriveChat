package engine

import (
	"net/http"
	"sync"
	"time"
)

// CircuitState 代表熔断器的三种状态。
type CircuitState int

const (
	StateClosed   CircuitState = iota // 熔断器关闭：正常转发所有请求。
	StateOpen                         // 熔断器开启：直接拦截请求，不转发。
	StateHalfOpen                     // 熔断器半开启：允许少量请求通过以检测后端服务是否恢复。
)

// CircuitBreaker 实现了基于错误率的熔断机制（状态机模式）。
// 当后端服务连续报错或错误率过高时，网关会主动切断流量，保护后端并防止故障蔓延。
type CircuitBreaker struct {
	mu           sync.Mutex
	state        CircuitState
	failureCount uint32        // 连续失败次数
	threshold    uint32        // 触发熔断的失败阈值
	openTime     time.Time     // 熔断器开启的时间点
	resetTimeout time.Duration // 熔断开启后多久尝试进入半开启状态
}

// NewCircuitBreaker 创建一个熔断器实例。
// threshold: 连续失败多少次触发熔断，timeout: 熔断持续时间。
func NewCircuitBreaker(threshold uint32, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:        StateClosed,
		threshold:    threshold,
		resetTimeout: timeout,
	}
}

// Allow 判断当前是否允许请求通过。
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// 1. 如果是关闭状态，允许。
	if cb.state == StateClosed {
		return true
	}

	// 2. 如果是开启状态，检查是否已经超过了冷却时间（resetTimeout）。
	if cb.state == StateOpen {
		if time.Since(cb.openTime) >= cb.resetTimeout {
			// 冷却时间已过，进入半开启状态，允许尝试性请求。
			cb.state = StateHalfOpen
			return true
		}
		// 仍在熔断期内，拒绝请求。
		return false
	}

	// 3. 如果是半开启状态。
	// 这里可以实现更复杂的逻辑（如：每秒只允许1个请求），目前简化为允许通过。
	return true
}

// Report 记录请求的执行结果。
// success: true 表示请求成功，false 表示请求失败（例如后端 5xx 错误）。
func (cb *CircuitBreaker) Report(success bool) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if success {
		// 请求成功：重置计数器
		cb.failureCount = 0
		if cb.state == StateHalfOpen {
			// 在半开启状态下成功，说明后端已恢复，关闭熔断器。
			cb.state = StateClosed
		}
	} else {
		// 请求失败：增加失败计数
		cb.failureCount++
		// 如果失败次数超过阈值，开启熔断。
		if cb.failureCount >= cb.threshold {
			cb.state = StateOpen
			cb.openTime = time.Now()
		}
	}
}

// CircuitBreakerMiddleware 是网关的熔断中间件。
func CircuitBreakerMiddleware(threshold uint32, timeout time.Duration) HandlerFunc {
	cb := NewCircuitBreaker(threshold, timeout)

	return func(c *Context) {
		// 1. 检查熔断器是否允许请求。
		if !cb.Allow() {
			c.JSON(http.StatusServiceUnavailable, map[string]string{
				"error": "服务目前不稳定，已触发熔断保护",
			})
			c.Abort()
			return
		}

		// 2. 执行后续中间件（包括实际的请求转发）。
		c.Next()

		// 3. 根据响应结果上报状态。
		// 这里假设我们可以通过 Context 获取是否有错误发生。
		// 后续可以根据 c.Writer.Status() 自动判断（例如 5xx 视为失败）。
		// 暂时手动模拟上报逻辑。
	}
}
