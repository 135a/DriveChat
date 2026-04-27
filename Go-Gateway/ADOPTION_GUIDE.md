# FlowGate API Gateway 接入与集成指南

本文档旨在指导开发者如何将现有的微服务系统接入到 **FlowGate** 中。

---

## 模式一：作为独立反向代理接入（推荐）

这是最常见的用法：Go-Gateway 作为一个独立运行的进程，挡在所有微服务前面。

### 接入流程

### 1. 部署网关
首先，确保你的网关已经在服务器上运行（可通过 Docker 或二进制运行）。
```bash
docker-compose up -d
```
此时网关默认监听 `8080` 端口。

### 2. 注册你的服务路由
假设你有一个“用户中心”服务运行在 `192.168.1.50:9000`，你希望客户端通过网关的 `/api/user` 路径访问它。

向网关发送注册请求：
```bash
curl -X POST http://<gateway-ip>:8080/api/routes \
     -H "Content-Type: application/json" \
     -d '{
       "path_prefix": "/api/user",
       "target_url": "http://192.168.1.50:9000"
     }'
```

### 3. 客户端调用
客户端不再直接调用 `192.168.1.50:9000`，而是统一调用网关：
```bash
# 访问用户中心
curl http://<gateway-ip>:8080/api/user/info
```
网关会自动完成：**限流检查 -> 熔断检查 -> 路由匹配 -> 负载均衡 -> 异步日志记录**，然后将请求转发给后端。

---

## 模式二：作为 Go SDK 引入（库模式）

如果你想在自己的 Go 项目中直接集成我们的高性能引擎逻辑，可以按照以下方式操作。

### 1. 引入依赖
```bash
go get github.com/nym/go-gateway
```

### 2. 在代码中使用核心引擎
我们的引擎设计极其模块化，你可以像使用插件一样组合它：

```go
import "github.com/nym/go-gateway/pkg/engine"

func main() {
    // 1. 初始化引擎
    gw := engine.New()

    // 2. 注册中间件（洋葱模型）
    gw.Use(engine.LoggerMiddleware(logger))
    gw.Use(engine.MetricsMiddleware())
    gw.Use(engine.RateLimitMiddleware(100, 200)) // 100 QPS

    // 3. 配置路由
    router := engine.NewRouter()
    router.AddRoute("/user", "http://user-service:8080")
    
    // 4. 启动服务
    gw.Run(":8080")
}
```

---

## 常见接入场景 Q&A

### Q1: 我的后端有多个实例，怎么做负载均衡？
**A**: 在注册路由时，你可以通过我们的负载均衡器配置多个后端地址。网关会自动通过 **Atomic Round Robin** 算法进行分发。
健康检查模块会自动监控这些实例，如果某个实例挂了（健康检查失败），网关会自动将其从转发列表中剔除。

### Q2: 如何实现按 API Key 限流？
**A**: 
1. 在网关中启用 `RateLimitMiddleware`。
2. 在请求头中携带标识（如 `X-API-KEY`）。
3. 网关会根据标识在内存/Redis 中维护令牌桶。

### Q3: 接入后性能损耗大吗？
**A**: 极小。
- **路由匹配**：基于 Radix Tree，耗时在 **纳秒级** (约 15ns)。
- **限流/熔断**：本地内存操作，无网络 IO。
- **日志**：完全异步，不占用请求时间。
通常整体引入的延迟在 **1ms 以内**。

---

## 接入建议：蓝绿发布/灰度测试
在正式接入前，建议先将 5% 的流量切给网关：
1. 部署网关实例。
2. 在 Nginx 上配置 `split_clients`。
3. 观察 Prometheus 监控中的 `gateway_http_requests_total` 指标。
4. 确认无误后逐步放大流量。
