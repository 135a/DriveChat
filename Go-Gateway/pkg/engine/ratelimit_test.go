package engine

import (
	"testing"
	"time"
)

// TestTokenBucket_Allow 测试令牌桶的基本限流逻辑。
func TestTokenBucket_Allow(t *testing.T) {
	// 创建一个 QPS 为 10，突发容量为 10 的限流器。
	tb := NewTokenBucket(10, 10)

	// 1. 初始状态下，桶应该是满的，允许连续获取 10 个令牌。
	for i := 0; i < 10; i++ {
		if !tb.Allow() {
			t.Errorf("第 %d 次尝试被错误拦截", i+1)
		}
	}

	// 2. 第 11 个请求应该被拦截。
	if tb.Allow() {
		t.Error("超额请求未被拦截")
	}

	// 3. 等待 100ms，应该生成 1 个新令牌 (10 QPS = 1 token per 100ms)。
	time.Sleep(110 * time.Millisecond)
	if !tb.Allow() {
		t.Error("等待后未生成新令牌")
	}
}
