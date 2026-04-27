package engine

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

// ProxyMiddleware 是网关最核心的“反向代理”中间件。
// 它负责将经过所有拦截器校验后的请求，最终转发到后端的真实服务器上。
func ProxyMiddleware() HandlerFunc {
	return func(c *Context) {
		// 1. 获取目标后端地址。
		// 该地址通常由之前的“路由”和“负载均衡”中间件确定并存入 Context 中。
		targetStr := c.TargetURL
		if targetStr == "" {
			// 如果走到转发这一步还没有确定目标，说明路由配置或逻辑有问题。
			c.JSON(http.StatusBadGateway, map[string]string{
				"error": "未找到有效的后端服务地址",
			})
			c.Abort()
			return
		}

		// 2. 解析目标 URL。
		target, err := url.Parse(targetStr)
		if err != nil {
			c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "后端地址解析失败",
			})
			c.Abort()
			return
		}

		// 3. 使用标准库提供的 SingleHostReverseProxy。
		// httputil.NewSingleHostReverseProxy 是生产环境广泛使用的成熟方案，
		// 它自动处理了分块传输（Chunked Encoding）、连接重用等复杂逻辑。
		proxy := httputil.NewSingleHostReverseProxy(target)

		// 4. 执行转发。
		// 注意：这里的 c.Writer 和 c.Request 是原始的 HTTP 对象。
		// 转发过程会阻塞直到后端响应或超时。
		proxy.ServeHTTP(c.Writer, c.Request)
		
		// 5. 转发完成后，链式调用结束。
		c.Next()
	}
}
