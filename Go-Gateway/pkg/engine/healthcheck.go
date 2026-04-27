package engine

import (
	"log"
	"net/http"
	"sync"
	"time"
)

// Backend 代表一个后端实例的信息及其当前的健康状态。
type Backend struct {
	URL   string // 实例的访问地址，如 http://10.0.0.1:8080
	Alive bool   // 当前是否存活
	mu    sync.RWMutex
}

// SetAlive 更新后端实例的存活状态。
func (b *Backend) SetAlive(alive bool) {
	b.mu.Lock()
	b.Alive = alive
	b.mu.Unlock()
}

// IsAlive 获取后端实例的存活状态。
func (b *Backend) IsAlive() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.Alive
}

// HealthChecker 负责周期性地探测后端实例的健康状况。
// 它通过定期发送轻量级请求（如 GET /health）来确保网关只将流量导向健康的节点。
type HealthChecker struct {
	backends []*Backend      // 需要监控的后端列表
	interval time.Duration  // 探测间隔时间
	stopChan chan struct{}  // 用于停止监控协程的信号
}

// NewHealthChecker 创建一个健康检查器。
func NewHealthChecker(interval time.Duration) *HealthChecker {
	return &HealthChecker{
		backends: make([]*Backend, 0),
		interval: interval,
		stopChan: make(chan struct{}),
	}
}

// AddBackend 添加一个需要监控的后端实例。
func (hc *HealthChecker) AddBackend(url string) {
	hc.backends = append(hc.backends, &Backend{
		URL:   url,
		Alive: true, // 初始默认为存活，等待第一次检查确认
	})
}

// Start 启动后台健康检查协程。
// 它会按照预设的间隔时间，并发地对所有后端进行探测。
func (hc *HealthChecker) Start() {
	go func() {
		log.Printf("[监控服务] 启动后台健康检查，检查周期: %v", hc.interval)
		ticker := time.NewTicker(hc.interval)
		for {
			select {
			case <-ticker.C:
				hc.checkAll()
			case <-hc.stopChan:
				ticker.Stop()
				return
			}
		}
	}()
}

// Stop 停止健康检查。
func (hc *HealthChecker) Stop() {
	close(hc.stopChan)
}

// checkAll 并发地对所有后端执行健康探测。
func (hc *HealthChecker) checkAll() {
	var wg sync.WaitGroup
	for _, b := range hc.backends {
		wg.Add(1)
		go func(backend *Backend) {
			defer wg.Done()
			
			// 发送 HTTP GET 请求进行探测。
			// 实际场景下可以配置特定的探测路径（如 /health）和更短的超时。
			client := http.Client{
				Timeout: 2 * time.Second,
			}
			
			resp, err := client.Get(backend.URL)
			if err != nil || resp.StatusCode != http.StatusOK {
				if backend.IsAlive() {
					log.Printf("[监控告警] 后端实例异常: %s, 错误: %v", backend.URL, err)
					backend.SetAlive(false)
				}
				return
			}
			resp.Body.Close()

			// 如果之前是异常状态，现在恢复了，则记录日志。
			if !backend.IsAlive() {
				log.Printf("[监控通知] 后端实例恢复正常: %s", backend.URL)
				backend.SetAlive(true)
			}
		}(b)
	}
	wg.Wait()
}

// GetHealthyBackends 返回当前所有健康的后端地址列表。
// 该方法供负载均衡器调用。
func (hc *HealthChecker) GetHealthyBackends() []string {
	healthy := make([]string, 0)
	for _, b := range hc.backends {
		if b.IsAlive() {
			healthy = append(healthy, b.URL)
		}
	}
	return healthy
}
