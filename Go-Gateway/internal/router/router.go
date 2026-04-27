// Package router holds the shared global Trie router instance,
// used by both the proxy middleware and the route sync service.
package router

import "github.com/nym/go-gateway/internal/trie"

// Global is the in-memory Trie router instance.
// It is initialized at startup and updated via Redis Pub/Sub.
var Global *trie.Trie

func init() {
	Global = trie.NewTrie()
}
