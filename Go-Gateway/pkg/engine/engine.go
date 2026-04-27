package engine

import (
	"net/http"
)

// Engine 是网关的核心控制结构。
// 它负责持有全局配置、中间件链，并实现 http.Handler 接口以分发所有入站请求。
type Engine struct {
	// middlewares 存储全局生效的中间件，所有请求都会按顺序经过这些处理器。
	middlewares []HandlerFunc
}

// New 创建并返回一个新的网关引擎实例。
// 初始化时会分配中间件切片的内存。
func New() *Engine {
	return &Engine{
		middlewares: make([]HandlerFunc, 0),
	}
}

// Use 将一个或多个中间件函数添加到引擎的全局中间件链中。
// 这些中间件将按照被添加的顺序执行。通常用于限流、日志、权限校验等通用功能。
func (e *Engine) Use(middlewares ...HandlerFunc) {
	e.middlewares = append(e.middlewares, middlewares...)
}

// ServeHTTP 实现了标准库的 http.Handler 接口。
// 它是网关处理每一个 HTTP 请求的入口点。
// 每次请求到来时，都会创建一个全新的 Context 对象来承载该请求的生命周期。
func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 实例化上下文，将响应写入器、请求对象以及全局中间件链关联起来。
	// 这种设计确保了请求处理的隔离性，同时也方便了后续的单元测试和性能分析。
	c := NewContext(w, r, e.middlewares)
	
	// 启动中间件链的第一个执行步骤。
	c.Next()
}

// 后续可以扩展 Group 功能，用于为特定的 API 路径前缀（如 /api/v1）配置专属的中间件。
