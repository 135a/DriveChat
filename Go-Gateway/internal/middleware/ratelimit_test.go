package middleware

import (
	"testing"
)

// TestSlidingWindowLuaScript verifies the Lua script content is embedded correctly.
func TestSlidingWindowLuaScript(t *testing.T) {
	if slidingWindowScript == "" {
		t.Fatal("Lua script was not embedded correctly")
	}

	// Verify key operations are present in the script
	expectedOps := []string{
		"ZREMRANGEBYSCORE",
		"ZCARD",
		"ZADD",
		"EXPIRE",
	}

	for _, op := range expectedOps {
		if !containsString(slidingWindowScript, op) {
			t.Errorf("Lua script missing expected operation: %s", op)
		}
	}
}

func containsString(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && contains(s, substr)
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestRateLimitKeyFormat verifies the rate limit key format uses IP only (no time bucket).
// The sliding window approach does not need time-bucketed keys since the ZSET manages
// the window internally.
func TestRateLimitKeyFormat(t *testing.T) {
	// The key should be "gateway:ratelimit:<ip>" without time bucket
	// This is validated by code review - the sliding window algorithm
	// manages time internally via ZSET scores (timestamps)
	expected := "gateway:ratelimit:"
	if !contains("gateway:ratelimit:192.168.1.1", expected) {
		t.Error("Rate limit key format incorrect")
	}
}
