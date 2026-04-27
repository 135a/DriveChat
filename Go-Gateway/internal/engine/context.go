package engine

import (
	"math"
	"net/http"
	"sync"
)

// abortIndex 定义了一个足够大的索引值，用于标记中间件链已经停止执行。
// 当 index 被设置为这个值时，Next() 方法中的循环将立即终止。
const abortIndex int8 = math.MaxInt8 / 2

// HandlerFunc 定义了网关中间件或拦截器的函数原型。
// 所有的处理逻辑（限流、路由、日志等）都通过实现此函数来接入网关流程。
type HandlerFunc func(*Context)

// Context 代表了单个请求在网关中的完整上下文。
// 它封装了原始的 HTTP 请求和响应对象，同时管理中间件链的执行顺序、
// 存储请求生命周期内的元数据，并提供辅助方法。
type Context struct {
	// Writer 是标准的 HTTP 响应写入接口。
	Writer  http.ResponseWriter
	// Request 是来自客户端的原始 HTTP 请求对象。
	Request *http.Request

	// Keys 用于存储仅在当前请求生命周期内有效的键值对数据。
	// 这在不同中间件之间传递数据（例如：用户信息、限流状态等）非常有用。
	Keys map[string]interface{}
	// mu 保护 Keys 映射的并发访问安全。
	mu   sync.RWMutex

	// handlers 存储了当前请求需要经过的所有中间件函数切片。
	handlers []HandlerFunc
	// index 记录当前正在执行的中间件索引。
	index    int8

	// TargetURL 存储经过路由匹配后确定的后端目标地址（例如：http://127.0.0.1:8081）。
	TargetURL string
}

// NewContext 创建并初始化一个新的请求上下文。
// w: 响应写入器，r: 请求对象，handlers: 初始化时加载的中间件链。
func NewContext(w http.ResponseWriter, r *http.Request, handlers []HandlerFunc) *Context {
	return &Context{
		Writer:   w,
		Request:  r,
		handlers: handlers,
		index:    -1, // 初始索引为 -1，第一次调用 Next() 时会变为 0。
	}
}

// Next 用于控制权移交给中间件链中的下一个处理器。
// 在中间件函数内部调用 c.Next() 可以实现“洋葱模型”式的逻辑（先执行后续逻辑，再回溯执行本中间件剩余逻辑）。
func (c *Context) Next() {
	c.index++
	// 遍历并依次执行后续的所有处理器。
	for c.index < int8(len(c.handlers)) {
		c.handlers[c.index](c)
		c.index++
	}
}

// Abort 立即终止中间件链的后续执行。
// 调用此方法后，当前中间件执行完毕后，后续的中间件将不再被触发。
func (c *Context) Abort() {
	c.index = abortIndex
}

// Set 在上下文中存储一个键值对。
// 此操作是并发安全的，底层使用了读写锁。
func (c *Context) Set(key string, value interface{}) {
	c.mu.Lock()
	if c.Keys == nil {
		c.Keys = make(map[string]interface{})
	}
	c.Keys[key] = value
	c.mu.Unlock()
}

// Get 从上下文中获取指定键的值。
// 返回获取到的值以及一个布尔值，用于表示该键是否存在。
func (c *Context) Get(key string) (value interface{}, exists bool) {
	c.mu.RLock()
	value, exists = c.Keys[key]
	c.mu.RUnlock()
	return
}

// JSON 是一个辅助方法，用于向客户端返回 JSON 格式的响应。
// code: HTTP 状态码，obj: 需要序列化的对象。
func (c *Context) JSON(code int, obj interface{}) {
	// 设置 Content-Type 为 JSON 格式。
	c.Writer.Header().Set("Content-Type", "application/json")
	// 写入 HTTP 状态码。
	c.Writer.WriteHeader(code)
	// 在实际高性能场景下，这里应集成类似 sonic 或 json-iterator 的快速序列化库。
	// 目前仅预留接口位置。
}
