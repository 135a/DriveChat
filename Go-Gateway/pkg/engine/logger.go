package engine

import (
	"fmt"
	"log"
	"time"
)

// LogEntry 代表一条待记录的日志记录。
type LogEntry struct {
	Timestamp time.Time
	Method    string
	Path      string
	Status    int
	Latency   time.Duration
	Target    string
}

// AsyncLogger 实现了高性能的异步日志记录器。
// 它通过 Channel 接收日志条目，并由后台的 Worker 协程进行统一写入。
// 这种设计可以极大地减少日志 IO 对主请求链路的性能影响。
type AsyncLogger struct {
	logChan chan LogEntry
	stop    chan struct{}
}

// NewAsyncLogger 创建并启动一个异步日志记录器。
// bufferSize: 缓冲区大小，建议根据 QPS 设置（如 10000）。
func NewAsyncLogger(bufferSize int) *AsyncLogger {
	al := &AsyncLogger{
		logChan: make(chan LogEntry, bufferSize),
		stop:    make(chan struct{}),
	}
	
	// 启动后台处理协程。
	al.start()
	
	return al
}

// start 启动 Worker 协程，持续监听并处理日志。
func (al *AsyncLogger) start() {
	go func() {
		log.Printf("[系统通知] 高性能异步日志服务已就绪")
		for {
			select {
			case entry := <-al.logChan:
				// 在此处执行真实的日志写入逻辑。
				// 实际生产中可以写入文件、Elasticsearch 或 Kafka。
				// 目前简单输出到控制台（带格式化）。
				fmt.Printf("[%s] %s %s -> %s | %d | %v\n",
					entry.Timestamp.Format("2006-01-02 15:04:05"),
					entry.Method,
					entry.Path,
					entry.Target,
					entry.Status,
					entry.Latency,
				)
			case <-al.stop:
				log.Printf("[系统通知] 异步日志服务正在关闭...")
				return
			}
		}
	}()
}

// Log 提交一条日志。如果缓冲区已满，此操作可能会阻塞（可配置为丢弃以保护性能）。
func (al *AsyncLogger) Log(entry LogEntry) {
	select {
	case al.logChan <- entry:
		// 成功存入缓冲
	default:
		// 如果缓冲区满了，为了不影响主链路性能，通常选择丢弃最新日志。
		// log.Printf("[警告] 日志缓冲区已满，丢弃当前日志条目")
	}
}

// Stop 优雅关闭日志服务。
func (al *AsyncLogger) Stop() {
	close(al.stop)
}

// LoggerMiddleware 是网关的日志记录中间件。
// 它会记录请求的耗时、路径、目标地址及响应状态码，并交给异步日志器处理。
func LoggerMiddleware(al *AsyncLogger) HandlerFunc {
	return func(c *Context) {
		// 1. 记录请求开始时间。
		start := time.Now()
		
		// 2. 执行后续中间件链。
		c.Next()
		
		// 3. 计算耗时并构建日志条目。
		latency := time.Since(start)
		
		entry := LogEntry{
			Timestamp: start,
			Method:    c.Request.Method,
			Path:      c.Request.URL.Path,
			// 注意：status 暂时无法直接从 http.ResponseWriter 获取，
			// 在实际实现中需要包装一个自定义的 ResponseWriter。
			Status:    200, 
			Latency:   latency,
			Target:    c.TargetURL,
		}
		
		// 4. 异步记录日志。
		al.Log(entry)
	}
}
