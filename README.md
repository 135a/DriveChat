# FlowGate API Gateway

> **FlowGate** 是一个基于 Go 语言实现的轻量级、高性能 API 网关。
>
> **核心组件**：Go 1.25 · Radix Tree · 令牌桶 · 状态机熔断 · Prometheus

---

## ✨ 核心特性

| 功能模块 | 实现方案 | 关键亮点 |
|:---|:---|:---|
| **高性能路由引擎** | 基于 Radix Tree（基数树） | 匹配复杂度 O(K)，支持精确/参数/通配符路由，无锁原子读取 |
| **动态负载均衡** | 原子轮询（Atomic Round Robin） | 零锁竞争，原子递增索引，支持扩展加权轮询 |
| **令牌桶限流** | Token Bucket 算法 | 支持平滑限流和突发流量控制，可按 IP/用户维度配置 |
| **熔断降级** | 状态机（关闭→开启→半开启） | 连续失败自动开启熔断，冷却后自动探测恢复 |
| **健康检查** | 后台 Goroutine 并发探测 | 定时检测后端实例，自动从负载均衡列表中剔除故障节点 |
| **异步日志** | Channel + Worker Pool | 日志异步写入，不阻塞主请求链路，性能损耗 < 1% |
| **可观测性** | Prometheus 指标导出 | 上报 QPS、P99/P95 延迟、错误率等量化指标 |
| **热更新路由** | Copy-on-Write + Atomic Pointer | 路由规则运行时更新，无需重启，读操作完全无锁 |

---

## 🏗️ 架构设计

```
┌─────────────────────────────────────────────────────────────────┐
│                         客户端请求                               │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Engine（拦截器链引擎）                         │
│              洋葱模型（Onion-Layered Middleware）                 │
│                                                                  │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌─────────�├── internal/
│   ├── trie/                       # Trie 前缀树（原有路由实现）
│   │   ├── trie.go
│   │   └── trie_test.go
│   ├── router/                     # 全局路由单例
│   ├── middleware/                 # Gin 中间件层（限流、黑名单、代理）
│   ├── controllers/                # 管理接口控制器
│   ├── models/                     # 数据库模型
│   └── services/                   # 业务服务（路由同步等）
│
├── pkg/
│   ├── engine/                     # ⭐ 核心网关引擎 (公开库)
│   │   ├── context.go              # 请求上下文（传递元数据、中间件链控制）
│   │   ├── engine.go               # 网关主引擎（中间件注册与分发）
│   │   ├── server.go               # HTTP 服务器封装（超时配置）
│   │   ├── router.go               # 基数树路由引擎（无锁读，CoW 写）
│   │   ├── ratelimit.go            # 令牌桶限流中间件
│   │   ├── circuitbreaker.go       # 熔断器（状态机实现）
│   │   ├── healthcheck.go          # 后台健康检查 Worker
│   │   ├── loadbalancer.go         # 负载均衡（原子轮询算法）
│   │   ├── proxy.go                # 反向代理转发中间件
│   │   ├── logger.go               # 异步日志（Channel 缓冲）
│   │   └── metrics.go              # Prometheus 指标收集中间件
│   │
│   ├── redis/                      # Redis 客户端封装
     │  (Proxy)   │    │
│                                              └─────┬──────┘    │
└────────────────────────────────────────────────────┼───────────┘
                                                      │
                  ┌───────────────┬──────────────────┘
                  ▼               ▼
          [后端服务 A]       [后端服务 B]  ←─ 健康检查持续监测
```

---

## 📁 项目结构

```
FlowGate/
├── cmd/
│   └── main.go                     # 程序入口（集成 Gin 路由和网关引擎）
│
├── internal/
│   ├── engine/                     # ⭐ 核心网关引擎（新增）
│   │   ├── context.go              # 请求上下文（传递元数据、中间件链控制）
│   │   ├── engine.go               # 网关主引擎（中间件注册与分发）
│   │   ├── server.go               # HTTP 服务器封装（超时配置）
│   │   ├── router.go               # 基数树路由引擎（无锁读，CoW 写）
│   │   ├── ratelimit.go            # 令牌桶限流中间件
│   │   ├── circuitbreaker.go       # 熔断器（状态机实现）
│   │   ├── healthcheck.go          # 后台健康检查 Worker
│   │   ├── loadbalancer.go         # 负载均衡（原子轮询算法）
│   │   ├── proxy.go                # 反向代理转发中间件
│   │   ├── logger.go               # 异步日志（Channel 缓冲）
│   │   └── metrics.go              # Prometheus 指标收集中间件
│   │
│   ├── trie/                       # Trie 前缀树（原有路由实现）
│   │   ├── trie.go
│   │   └── trie_test.go
│   ├── router/                     # 全局路由单例
│   ├── middleware/                 # Gin 中间件层（限流、黑名单、代理）
│   ├── controllers/                # 管理接口控制器
│   ├── models/                     # 数据库模型
│   └── services/                   # 业务服务（路由同步等）
│
├── pkg/
│   ├── redis/                      # Redis 客户端封装
│   ├── mysql/                      # MySQL 客户端封装
│   └── lua/                        # Redis Lua 脚本（分布式限流）
│
├── benchmark/                      # ⭐ 量化测试套件
│   ├── Gateway_QuantitativeTest.jmx  # JMeter 测试计划（5大场景）
│   ├── QUANTITATIVE_TEST_GUIDE.md    # 完整量化测试手册
│   └── ...
│
├── scripts/
│   └── stress_test.go              # Go 内置压测工具
│
├── conf/                           # 配置文件
├── Dockerfile
└── docker-compose.yml
```

---

## 🚀 快速启动

### 方式 A：Docker Compose（推荐）

```bash
# 克隆项目
git clone <repo-url>
cd FlowGate/FlowGate

# 一键启动（网关 + Redis + MySQL）
docker-compose up -d

# 验证启动成功
curl http://localhost:8080/api/routes
```

### 方式 B：本地开发运行

```bash
# 安装依赖
go mod tidy

# 启动（需要本地 Redis 和 MySQL）
go run cmd/main.go
```

---

## 🛠️ 管理 API

通过 RESTful API 动态管理路由规则和黑名单，无需重启网关。

### 路由规则管理

```bash
# 查看所有路由
curl http://localhost:8080/api/routes

# 创建路由（将 /api/v1/user 代理到 user-service）
curl -X POST http://localhost:8080/api/routes \
     -H "Content-Type: application/json" \
     -d '{"path_prefix": "/api/v1/user", "target_url": "http://user-service:8081"}'

# 更新路由
curl -X PUT http://localhost:8080/api/routes/1 \
     -H "Content-Type: application/json" \
     -d '{"target_url": "http://user-service-v2:8082"}'

# 删除路由
curl -X DELETE http://localhost:8080/api/routes/1
```

### 黑名单管理

```bash
# 封禁 IP
curl -X POST http://localhost:8080/api/blacklist \
     -H "Content-Type: application/json" \
     -d '{"ip": "192.168.1.100"}'

# 查看黑名单
curl http://localhost:8080/api/blacklist

# 解封
curl -X DELETE http://localhost:8080/api/blacklist/1
```

### 监控指标

```bash
# 查看内置指标（QPS、延迟、错误数）
curl http://localhost:8080/api/metrics
```

---

## 📊 量化性能测试

本项目内置了完整的**量化测试套件**，可生成具体的性能数据用于对比和优化。

### 方式 1：JMeter 全功能测试（推荐）

```bash
# 1. 安装 JMeter：https://jmeter.apache.org/download_jmeter.cgi
# 2. 创建报告目录
mkdir benchmark/reports

# 3. 运行完整测试计划（5个场景）
jmeter -n \
  -t benchmark/Gateway_QuantitativeTest.jmx \
  -l benchmark/reports/result.jtl \
  -e -o benchmark/reports/dashboard

# 4. 打开 HTML 报告
open benchmark/reports/dashboard/index.html  # macOS
start benchmark/reports/dashboard/index.html  # Windows
```

**5 大测试场景**：

| 场景 | 目标 | 期望结果 |
|:---|:---|:---|
| **场景1：基准路由** | 100并发 × 20,000请求 | QPS > 5000，P99 < 100ms |
| **场景2：限流触发** | 200并发超阈值冲击 | 超额请求返回 429 |
| **场景3：熔断降级** | 模拟后端宕机 | 502 → 503 状态切换 |
| **场景4：负载均衡** | 多后端分发验证 | A/B 节点命中率偏差 < 1% |
| **场景5：极限压测** | 500并发持续30秒 | Error < 0.1%，P99 < 100ms |

> 📖 详细测试指南：[benchmark/QUANTITATIVE_TEST_GUIDE.md](./Go-Gateway/benchmark/QUANTITATIVE_TEST_GUIDE.md)

### 方式 2：Go 内置 Benchmark（微基准）

```bash
cd internal/engine

# 运行所有基准测试
go test -bench=. -benchmem -count=3

# 典型输出（量化路由匹配耗时）
# BenchmarkRouter_Search-8   68543210   15.3 ns/op   0 B/op   0 allocs/op
#                                        ↑ 单次路由匹配 15 纳秒，零内存分配
```

### 方式 3：单元测试验证

```bash
# 运行所有单元测试（覆盖限流、熔断、路由等核心逻辑）
go test ./internal/engine/... -v

# 预期输出
# --- PASS: TestTokenBucket_Allow (0.11s)        ✅ 限流逻辑正确
# --- PASS: TestCircuitBreaker_StateTransition (0.11s) ✅ 熔断状态机正确
# --- PASS: TestRouter_Search (0.00s)            ✅ 路由匹配正确
```

---

## ⚙️ 核心技术实现

### 1. 无锁路由引擎（Copy-on-Write）

```go
// 读操作：完全无锁，直接原子加载指针
func (r *Router) Search(path string) (string, map[string]string, bool) {
    curr := r.root.Load()  // 原子读，无锁！
    // ... 遍历匹配
}

// 写操作：深度拷贝后原子替换（Copy-on-Write）
func (r *Router) AddRoute(path, targetURL string) {
    oldRoot := r.root.Load()
    newRoot := oldRoot.copy()   // 拷贝，不影响正在读取的 Goroutine
    insert(newRoot, path, targetURL)
    r.root.Store(newRoot)       // 原子替换
}
```

### 2. 令牌桶限流（Token Bucket）

```go
// 平滑限流：根据时间流逝自动补充令牌
func (tb *TokenBucket) Allow() bool {
    elapsed := time.Since(tb.lastRefillAt).Seconds()
    tb.tokens = min(tb.capacity, tb.tokens + elapsed * tb.rate)
    if tb.tokens >= 1.0 {
        tb.tokens -= 1.0
        return true  // 放行
    }
    return false  // 限流
}
```

### 3. 异步日志（Channel 缓冲）

```go
// 主链路：非阻塞投递日志（select + default 丢弃策略）
func (al *AsyncLogger) Log(entry LogEntry) {
    select {
    case al.logChan <- entry:  // 投入缓冲
    default:                   // 缓冲满则丢弃，保护主链路
    }
}
// Worker 协程：独立消费，不占用请求处理线程
go func() {
    for entry := range al.logChan {
        writeToOutput(entry)
    }
}()
```

---

## 🔑 技术选型说明

| 技术决策 | 方案选择 | 原因 |
|:---|:---|:---|
| 路由数据结构 | Radix Tree（基数树） | O(K) 匹配，优于线性扫描和 HashMap |
| 路由更新并发安全 | Copy-on-Write + atomic.Pointer | 读无锁，高并发下性能远超 RWMutex |
| 限流算法 | 令牌桶（Token Bucket） | 支持突发流量，比漏桶更灵活 |
| 日志写入 | Channel 异步 | 不阻塞主请求链路，GC 友好 |
| 负载均衡计数 | atomic.AddUint64 | 无锁递增，避免 Mutex 竞争 |

---

## 📈 性能量化（参考数据）

> 以下数据基于本机测试（8核 CPU），实际结果请运行 benchmark 套件获取。

```
路由匹配（Radix Tree）：~15 ns/op，0 B/op（零内存分配）
网关吞吐量（100并发）：> 5,000 QPS
P99 延迟（100并发）：< 100 ms
限流器准确率：误差 < 2%（令牌桶算法）
```

---

## 🗺️ 开发路线图

- [x] 核心引擎与中间件链（洋葱模型）
- [x] Radix Tree 高性能路由引擎（无锁读）
- [x] 令牌桶限流中间件
- [x] 熔断降级（状态机）
- [x] 主动健康检查（后台 Worker）
- [x] 负载均衡（轮询）
- [x] 异步日志系统
- [x] Prometheus 指标接口
- [x] JMeter 量化测试套件（5大场景）
- [ ] gRPC/HTTP2 后端支持
- [ ] 动态配置热加载（Etcd/Consul）
- [ ] 基于 JWT 的 API Key 鉴权
- [ ] 响应缓存（Redis 热点接口缓存）
- [ ] Grafana Dashboard 模板

---

## 📚 相关文档

- [量化测试手册](./Go-Gateway/benchmark/QUANTITATIVE_TEST_GUIDE.md)
- [JMeter 测试计划](./Go-Gateway/benchmark/Gateway_QuantitativeTest.jmx)
