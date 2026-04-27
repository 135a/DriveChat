package trie

import (
	"testing"
)

func TestTrieInsertAndExactMatch(t *testing.T) {
	tr := NewTrie()
	tr.Insert("/api/users", "http://user-service:8080")
	tr.Insert("/api/orders", "http://order-service:8080")

	route, _ := tr.Search("/api/users")
	if route == nil {
		t.Fatal("Expected to find route for /api/users")
	}
	if route.TargetURL != "http://user-service:8080" {
		t.Errorf("Expected target http://user-service:8080, got %s", route.TargetURL)
	}

	route, _ = tr.Search("/api/orders")
	if route == nil {
		t.Fatal("Expected to find route for /api/orders")
	}
	if route.TargetURL != "http://order-service:8080" {
		t.Errorf("Expected target http://order-service:8080, got %s", route.TargetURL)
	}
}

func TestTrieWildcardMatch(t *testing.T) {
	tr := NewTrie()
	tr.Insert("/api/:version/users", "http://user-service:8080")

	route, params := tr.Search("/api/v1/users")
	if route == nil {
		t.Fatal("Expected to find route for /api/v1/users")
	}
	if route.TargetURL != "http://user-service:8080" {
		t.Errorf("Expected target http://user-service:8080, got %s", route.TargetURL)
	}
	if params["version"] != "v1" {
		t.Errorf("Expected param version=v1, got %s", params["version"])
	}

	// Different version should also match
	route, params = tr.Search("/api/v2/users")
	if route == nil {
		t.Fatal("Expected to find route for /api/v2/users")
	}
	if params["version"] != "v2" {
		t.Errorf("Expected param version=v2, got %s", params["version"])
	}
}

func TestTrieNoMatch(t *testing.T) {
	tr := NewTrie()
	tr.Insert("/api/users", "http://user-service:8080")

	route, _ := tr.Search("/api/products")
	if route != nil {
		t.Error("Expected no match for /api/products")
	}

	route, _ = tr.Search("/completely/different")
	if route != nil {
		t.Error("Expected no match for /completely/different")
	}
}

func TestTriePrefixMatch(t *testing.T) {
	tr := NewTrie()
	tr.Insert("/api", "http://api-service:8080")

	// /api itself should match
	route, _ := tr.Search("/api")
	if route == nil {
		t.Fatal("Expected to find route for /api")
	}
	if route.TargetURL != "http://api-service:8080" {
		t.Errorf("Expected target http://api-service:8080, got %s", route.TargetURL)
	}
}

func TestTrieStaticOverWildcard(t *testing.T) {
	tr := NewTrie()
	tr.Insert("/api/:id", "http://generic-service:8080")
	tr.Insert("/api/health", "http://health-service:8080")

	// Static match should take priority
	route, _ := tr.Search("/api/health")
	if route == nil {
		t.Fatal("Expected to find route for /api/health")
	}
	if route.TargetURL != "http://health-service:8080" {
		t.Errorf("Expected static match http://health-service:8080, got %s", route.TargetURL)
	}

	// Non-static should fall through to wildcard
	route, params := tr.Search("/api/123")
	if route == nil {
		t.Fatal("Expected to find route for /api/123")
	}
	if route.TargetURL != "http://generic-service:8080" {
		t.Errorf("Expected wildcard match http://generic-service:8080, got %s", route.TargetURL)
	}
	if params["id"] != "123" {
		t.Errorf("Expected param id=123, got %s", params["id"])
	}
}

func TestTrieDelete(t *testing.T) {
	tr := NewTrie()
	tr.Insert("/api/users", "http://user-service:8080")

	route, _ := tr.Search("/api/users")
	if route == nil {
		t.Fatal("Expected to find route before deletion")
	}

	deleted := tr.Delete("/api/users")
	if !deleted {
		t.Fatal("Expected Delete to return true")
	}

	route, _ = tr.Search("/api/users")
	if route != nil {
		t.Error("Expected no match after deletion")
	}
}

func TestTrieClear(t *testing.T) {
	tr := NewTrie()
	tr.Insert("/api/users", "http://user-service:8080")
	tr.Insert("/api/orders", "http://order-service:8080")

	tr.Clear()

	route, _ := tr.Search("/api/users")
	if route != nil {
		t.Error("Expected no match after Clear()")
	}
	route, _ = tr.Search("/api/orders")
	if route != nil {
		t.Error("Expected no match after Clear()")
	}
}

func TestTrieConcurrentAccess(t *testing.T) {
	tr := NewTrie()
	done := make(chan bool, 10)

	// Concurrent writes
	for i := 0; i < 5; i++ {
		go func(n int) {
			tr.Insert("/api/service"+string(rune('A'+n)), "http://service:8080")
			done <- true
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 5; i++ {
		go func() {
			tr.Search("/api/serviceA")
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}
