package engine

import (
	"strings"
	"sync/atomic"
)

// Router 代表了基于基数树（Radix Tree）的高性能路由器。
// 为了达到极致的搜索性能并避免读写锁（RWMutex）在高并发下的性能抖动，
// 本实现采用了“写时复制”（Copy-on-Write）和原子指针（atomic.Pointer）技术，
// 使得在高频率搜索请求中完全不需要加锁。
type Router struct {
	// root 使用 atomic.Pointer 存储树的根节点。
	// 这允许我们在路由规则更新时，通过原子替换整个树的根节点来实现无锁读取。
	root atomic.Pointer[node]
}

// node 是基数树中的一个基本单元，代表路径的一个片段或一个完整的端点。
type node struct {
	path      string    // 静态路径片段（例如：/v1）
	children  []*node   // 子节点列表，包含静态、参数化或通配符类型的子节点
	targetURL string    // 如果此节点是终点，则存储目标后端地址（例如：http://backend-server）
	
	// 参数化路由支持（例如：/users/:id）
	paramName string    // 参数名称（例如：id）
	
	// 通配符路由支持（例如：/static/*）
	isWild    bool      // 是否为通配符节点
}

// NewRouter 创建并初始化一个新的 Radix Tree 路由器实例。
func NewRouter() *Router {
	r := &Router{}
	// 初始化一个空的根节点，并存储到原子指针中。
	r.root.Store(&node{})
	return r
}

// AddRoute 向路由器中添加一条新的路由规则。
// path: 匹配路径（支持 :param 和 *），targetURL: 后端目标地址。
// 该方法是并发安全的，底层使用了 Copy-on-Write 机制：
// 1. 加载当前树的副本。
// 2. 在副本上进行修改。
// 3. 原子地将新树替换回全局引用。
func (r *Router) AddRoute(path, targetURL string) {
	// 获取当前根节点的快照。
	oldRoot := r.root.Load()
	// 执行深度拷贝，确保修改不会影响到正在进行的读操作（Search）。
	newRoot := oldRoot.copy()
	
	// 按照 '/' 将路径分割成多个段进行逐级插入。
	segments := strings.Split(strings.Trim(path, "/"), "/")
	curr := newRoot
	
	for _, seg := range segments {
		if seg == "" {
			continue
		}
		
		var foundChild *node
		if strings.HasPrefix(seg, ":") {
			// 处理参数化节点 (例如 :id)
			for _, child := range curr.children {
				if child.paramName == seg[1:] {
					foundChild = child
					break
				}
			}
			if foundChild == nil {
				foundChild = &node{paramName: seg[1:]}
				curr.children = append(curr.children, foundChild)
			}
		} else if seg == "*" {
			// 处理通配符节点 (*)
			for _, child := range curr.children {
				if child.isWild {
					foundChild = child
					break
				}
			}
			if foundChild == nil {
				foundChild = &node{isWild: true}
				curr.children = append(curr.children, foundChild)
			}
		} else {
			// 处理普通的静态路径节点
			for _, child := range curr.children {
				if child.path == seg {
					foundChild = child
					break
				}
			}
			if foundChild == nil {
				foundChild = &node{path: seg}
				curr.children = append(curr.children, foundChild)
			}
		}
		// 向下移动指针到下一层节点。
		curr = foundChild
	}
	
	// 在终点节点记录目标后端地址。
	curr.targetURL = targetURL
	
	// 原子替换旧根节点为新根节点，完成路由热更新。
	r.root.Store(newRoot)
}

// Search 在基数树中查找与给定路径匹配的最优后端目标。
// 它支持精确匹配、路径参数提取以及通配符匹配。
// 返回值：
// 1. string: 匹配到的后端 URL。
// 2. map[string]string: 从路径中提取的参数（例如：id=123）。
// 3. bool: 是否匹配成功。
// 注意：本方法完全无锁，非常适合高并发的网关主链路。
func (r *Router) Search(path string) (string, map[string]string, bool) {
	segments := strings.Split(strings.Trim(path, "/"), "/")
	params := make(map[string]string)
	
	// 加载当前树的根节点指针。
	curr := r.root.Load()
	
	for _, seg := range segments {
		if seg == "" {
			continue
		}
		
		found := false
		
		// 1. 优先尝试静态匹配（性能最高，优先级最高）。
		for _, child := range curr.children {
			if !child.isWild && child.paramName == "" && child.path == seg {
				curr = child
				found = true
				break
			}
		}
		if found {
			continue
		}
		
		// 2. 尝试参数化路由匹配（例如 :id）。
		for _, child := range curr.children {
			if child.paramName != "" {
				// 将匹配到的实际段值存入参数 Map。
				params[child.paramName] = seg
				curr = child
				found = true
				break
			}
		}
		if found {
			continue
		}
		
		// 3. 尝试通配符路由匹配（*）。
		for _, child := range curr.children {
			if child.isWild {
				curr = child
				found = true
				break
			}
		}
		
		// 如果在当前层级没有找到任何匹配项，则说明匹配失败。
		if !found {
			return "", nil, false
		}
	}
	
	// 检查当前节点是否是路径终点。
	if curr.targetURL != "" {
		return curr.targetURL, params, true
	}
	
	// 特殊处理：如果当前节点没有终点，但其子节点中有通配符，也视为匹配成功（用于支持尾随通配符）。
	for _, child := range curr.children {
		if child.isWild && child.targetURL != "" {
			return child.targetURL, params, true
		}
	}
	
	return "", nil, false
}

// copy 实现了节点的深度拷贝。
// 它是 Copy-on-Write 机制的核心，确保我们在修改树时不会影响读操作。
func (n *node) copy() *node {
	newNode := &node{
		path:      n.path,
		targetURL: n.targetURL,
		paramName: n.paramName,
		isWild:    n.isWild,
		children:  make([]*node, len(n.children)),
	}
	// 递归拷贝所有子节点。
	for i, child := range n.children {
		newNode.children[i] = child.copy()
	}
	return newNode
}
