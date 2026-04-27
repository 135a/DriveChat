package services

import (
	"log"

	"github.com/nym/go-gateway/internal/router"
	"github.com/nym/go-gateway/pkg/redis"
)

// RouteSyncChannel is the Redis Pub/Sub channel for route sync notifications.
const RouteSyncChannel = "gateway:route:sync"

// InitRouteSync loads all routes from Redis into the in-memory Trie and
// starts a background goroutine to listen for route update events.
func InitRouteSync() {
	// 1. Full load from Redis Hash
	loadAllRoutes()

	// 2. Subscribe to route sync channel
	go subscribeRouteSync()
}

// loadAllRoutes loads all routes from Redis Hash into the Trie.
func loadAllRoutes() {
	routes, err := redis.Client.HGetAll(redis.Ctx, "global:gateway:routes").Result()
	if err != nil {
		log.Printf("[RouteSync] Failed to load routes from Redis: %v", err)
		return
	}

	router.Global.Clear()
	for prefix, target := range routes {
		router.Global.Insert(prefix, target)
	}
	log.Printf("[RouteSync] Loaded %d routes into Trie", len(routes))
}

// subscribeRouteSync subscribes to Redis Pub/Sub channel and rebuilds
// the Trie when route update messages are received.
func subscribeRouteSync() {
	pubsub := redis.Client.Subscribe(redis.Ctx, RouteSyncChannel)
	defer pubsub.Close()

	ch := pubsub.Channel()
	log.Printf("[RouteSync] Subscribed to channel: %s", RouteSyncChannel)

	for msg := range ch {
		log.Printf("[RouteSync] Received sync signal: %s, rebuilding Trie...", msg.Payload)
		loadAllRoutes()
	}
}

// PublishRouteSync publishes a route sync notification to trigger Trie rebuild.
func PublishRouteSync() {
	err := redis.Client.Publish(redis.Ctx, RouteSyncChannel, "reload").Err()
	if err != nil {
		log.Printf("[RouteSync] Failed to publish sync signal: %v", err)
	}
}
