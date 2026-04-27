package engine

import (
	"testing"
)

func BenchmarkRouter_Search(b *testing.B) {
	router := NewRouter()
	router.AddRoute("/api/v1/user", "http://user-service")
	router.AddRoute("/api/v1/order", "http://order-service")
	router.AddRoute("/api/v2/payment", "http://payment-service")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			router.Search("/api/v1/user")
		}
	})
}
