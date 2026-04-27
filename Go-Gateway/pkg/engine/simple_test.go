package engine

import "testing"

func TestRouter_Search(t *testing.T) {
	router := NewRouter()
	router.AddRoute("/api/v1/user", "http://user-service")
	target, params, found := router.Search("/api/v1/user")
	if !found || target != "http://user-service" {
		t.Errorf("expected http://user-service, got %s", target)
	}
	_ = params
}
