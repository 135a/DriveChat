package trie

import (
	"strings"
	"sync"
)

// RouteInfo stores the backend target configuration for a matched route.
type RouteInfo struct {
	TargetURL  string
	PathPrefix string
}

// Node represents a single node in the Trie tree.
type Node struct {
	children map[string]*Node // static segment children
	wildcard *Node            // wildcard child (e.g., :param)
	paramKey string           // parameter name for wildcard nodes
	route    *RouteInfo       // non-nil if this node is a terminal route
}

// Trie is a thread-safe prefix tree for URL path matching.
type Trie struct {
	root *Node
	mu   sync.RWMutex
}

// NewTrie creates an empty Trie.
func NewTrie() *Trie {
	return &Trie{
		root: &Node{
			children: make(map[string]*Node),
		},
	}
}

// Insert adds a route to the Trie. Path segments starting with ':' are treated as wildcards.
// Example: /api/:version/users
func (t *Trie) Insert(pathPrefix, targetURL string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	segments := splitPath(pathPrefix)
	current := t.root

	for _, seg := range segments {
		if strings.HasPrefix(seg, ":") {
			// Wildcard segment
			if current.wildcard == nil {
				current.wildcard = &Node{
					children: make(map[string]*Node),
					paramKey: seg[1:],
				}
			}
			current = current.wildcard
		} else {
			// Static segment
			child, exists := current.children[seg]
			if !exists {
				child = &Node{
					children: make(map[string]*Node),
				}
				current.children[seg] = child
			}
			current = child
		}
	}

	current.route = &RouteInfo{
		TargetURL:  targetURL,
		PathPrefix: pathPrefix,
	}
}

// Search finds the best matching route for the given path.
// It prioritizes exact matches over wildcard matches (longest prefix match).
// Returns the RouteInfo and extracted path parameters.
func (t *Trie) Search(path string) (*RouteInfo, map[string]string) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	segments := splitPath(path)
	params := make(map[string]string)

	route := t.search(t.root, segments, 0, params)
	if route != nil {
		return route, params
	}
	return nil, nil
}

// search recursively finds the best match, prioritizing static segments over wildcards.
func (t *Trie) search(node *Node, segments []string, index int, params map[string]string) *RouteInfo {
	// Base case: consumed all segments
	if index == len(segments) {
		return node.route
	}

	seg := segments[index]

	// 1. Try exact (static) match first
	if child, exists := node.children[seg]; exists {
		if result := t.search(child, segments, index+1, params); result != nil {
			return result
		}
	}

	// 2. Try wildcard match
	if node.wildcard != nil {
		params[node.wildcard.paramKey] = seg
		if result := t.search(node.wildcard, segments, index+1, params); result != nil {
			return result
		}
		delete(params, node.wildcard.paramKey) // backtrack
	}

	// 3. Check if current node has a route (prefix match)
	if node.route != nil {
		return node.route
	}

	return nil
}

// Delete removes a route from the Trie.
func (t *Trie) Delete(pathPrefix string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	segments := splitPath(pathPrefix)
	return t.delete(t.root, segments, 0)
}

func (t *Trie) delete(node *Node, segments []string, index int) bool {
	if index == len(segments) {
		if node.route != nil {
			node.route = nil
			return true
		}
		return false
	}

	seg := segments[index]

	if strings.HasPrefix(seg, ":") {
		if node.wildcard != nil {
			return t.delete(node.wildcard, segments, index+1)
		}
	} else {
		if child, exists := node.children[seg]; exists {
			return t.delete(child, segments, index+1)
		}
	}

	return false
}

// Clear removes all routes from the Trie (used during hot-reload).
func (t *Trie) Clear() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.root = &Node{
		children: make(map[string]*Node),
	}
}

// splitPath splits a URL path into segments, filtering empty strings.
func splitPath(path string) []string {
	raw := strings.Split(strings.Trim(path, "/"), "/")
	segments := make([]string, 0, len(raw))
	for _, s := range raw {
		if s != "" {
			segments = append(segments, s)
		}
	}
	return segments
}
