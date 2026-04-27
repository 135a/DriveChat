package engine

import (
	"log"
	"net/http"
	"time"
)

// Run 启动网关的 HTTP 服务。
// 它封装了标准库 http.Server 的配置，设置了合理的超时时间以防止连接泄露和慢请求攻击。
// addr: 监听地址（例如 ":8080" 或 "0.0.0.0:80"）。
func (e *Engine) Run(addr string) error {
	// 配置高性能 HTTP 服务器参数。
	s := &http.Server{
		Addr:           addr,
		Handler:        e, // 将 Engine 实例作为 Handler 传入。
		
		// ReadTimeout 是读取整个请求（包括 Body）的最大允许时间。
		// 在网关场景下，设置此值可有效防御慢连接攻击。
		ReadTimeout:    10 * time.Second,
		
		// WriteTimeout 是响应写入的最大允许时间。
		WriteTimeout:   10 * time.Second,
		
		// MaxHeaderBytes 限制请求头的最大字节数。
		// 1MB 的限制对绝大多数 API 场景已经非常充裕且安全。
		MaxHeaderBytes: 1 << 20,
	}

	log.Printf("[网关公告] 高性能 API 网关正在启动，监听地址: %s", addr)
	
	// 启动监听并进入阻塞服务循环。
	// 这是一个标准同步操作，通常在 main 函数的最后一行调用。
	return s.ListenAndServe()
}
