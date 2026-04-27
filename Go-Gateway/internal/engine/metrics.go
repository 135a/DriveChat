package engine

import (
	"time"
	// "github.com/prometheus/client_golang/prometheus"
	// "github.com/prometheus/client_golang/prometheus/promauto"
)

// 注意：由于环境权限限制无法安装 Prometheus SDK，以下代码已暂时注释。
// 在实际生产环境中，请取消注释并确保已运行 `go get github.com/prometheus/client_golang/prometheus`。

/*
var (
	httpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gateway_http_requests_total",
		Help: "网关处理的 HTTP 请求总数",
	}, []string{"method", "path", "status"})

	httpRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "gateway_http_request_duration_seconds",
		Help:    "网关处理 HTTP 请求的耗时分布（秒）",
		Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
	}, []string{"method", "path"})
)
*/

// MetricsMiddleware 是网关的指标收集中间件。
func MetricsMiddleware() HandlerFunc {
	return func(c *Context) {
		start := time.Now()
		c.Next()
		_ = time.Since(start).Seconds()

		// httpRequestsTotal.WithLabelValues(c.Request.Method, c.Request.URL.Path, "200").Inc()
		// httpRequestDuration.WithLabelValues(c.Request.Method, c.Request.URL.Path).Observe(duration)
	}
}
