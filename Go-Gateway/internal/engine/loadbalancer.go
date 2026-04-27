package engine

import (
	"sync/atomic"
)

// LoadBalancer 定义了负载均衡器的统一接口。
// 无论底层使用哪种算法（轮询、随机、最小连接数等），最终都通过 Next 返回一个后端地址。
type LoadBalancer interface {
	Next() (string, bool)
}

// RoundRobinBalancer 实现了经典的“加权轮询”负载均衡算法。
// 它通过依次选择列表中的后端实例来确保请求被均匀分配。
type RoundRobinBalancer struct {
	backends []string
	index    uint64 // 使用原子操作递增的索引，确保并发安全
}

// NewRoundRobinBalancer 创建一个轮询负载均衡器。
func NewRoundRobinBalancer(backends []string) *RoundRobinBalancer {
	return &RoundRobinBalancer{
		backends: backends,
		index:    0,
	}
}

// Next 返回下一个可用的后端地址。
// 使用原子加法（Atomic Add）来实现无锁的索引递增，能够承受极高的并发请求。
func (rr *RoundRobinBalancer) Next() (string, bool) {
	if len(rr.backends) == 0 {
		return "", false
	}

	// 1. 原子递增索引
	// 2. 取模运算确保索引在合法范围内
	idx := atomic.AddUint64(&rr.index, 1)
	target := rr.backends[(idx-1)%uint64(len(rr.backends))]
	
	return target, true
}

// WeightedRoundRobin (预留扩展)
// 在实际工业场景中，可以根据后端节点的权重（Weight）来分配流量。
// 权重高的节点将获得更多的请求机会。
