package engine

import (
	"testing"
	"time"
)

// TestCircuitBreaker_StateTransition 测试熔断器的状态转换逻辑。
func TestCircuitBreaker_StateTransition(t *testing.T) {
	// 阈值为 3 次失败，冷却时间 100ms。
	cb := NewCircuitBreaker(3, 100*time.Millisecond)

	// 1. 初始状态为 Closed。
	if !cb.Allow() {
		t.Error("初始状态不应拦截请求")
	}

	// 2. 连续 3 次失败。
	cb.Report(false)
	cb.Report(false)
	cb.Report(false)

	// 3. 状态应转为 Open。
	if cb.Allow() {
		t.Error("达到阈值后应开启熔断")
	}

	// 4. 等待超过冷却时间。
	time.Sleep(110 * time.Millisecond)

	// 5. 状态应转为 Half-Open，允许探测请求。
	if !cb.Allow() {
		t.Error("冷却期过后应进入半开启状态")
	}

	// 6. 探测成功，状态应转回 Closed。
	cb.Report(true)
	if !cb.Allow() {
		t.Error("探测成功后应恢复正常")
	}
}
