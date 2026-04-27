# FlowGate API Gateway 量化测试手册
> 版本：v2.0 | 适配完整高性能网关架构（路由 + 限流 + 熔断 + 负载均衡）

---

## 目录

1. [测试环境准备](#1-测试环境准备)
2. [JMeter 安装与配置](#2-jmeter-安装与配置)
3. [测试场景详解](#3-测试场景详解)
   - [场景 1：基准路由性能测试](#场景-1基准路由性能测试)
   - [场景 2：令牌桶限流触发验证](#场景-2令牌桶限流触发验证)
   - [场景 3：熔断降级效果验证](#场景-3熔断降级效果验证)
   - [场景 4：负载均衡分布验证](#场景-4负载均衡分布验证)
   - [场景 5：极限压力测试](#场景-5极限压力测试)
4. [如何读懂量化报告](#4-如何读懂量化报告)
5. [Go 内置 Benchmark 使用](#5-go-内置-benchmark-使用)
6. [结合 Prometheus 实时监控](#6-结合-prometheus-实时监控)
7. [量化数据填写模板](#7-量化数据填写模板)

---

## 1. 测试环境准备

### 1.1 启动网关

在 `d:\Go-Gateway\Go-Gateway` 目录下运行：

```bash
# 方式 A：直接运行（开发环境）
go run cmd/main.go

# 方式 B：Docker Compose（推荐，包含 Redis/MySQL）
docker-compose up -d
```

### 1.2 确认网关正常响应

```bash
curl http://localhost:8080/api/v1/user
```

**预期**：收到后端服务的响应（或 502 如果后端未配置）。

### 1.3 准备后端模拟服务（可选）

如果需要测试真实转发，用最简单的方式启动两个"假后端"以模拟负载均衡节点：

```bash
# 后端 A（端口 8081）
python -m http.server 8081

# 后端 B（端口 8082）
python -m http.server 8082
```

然后在网关路由配置中，将 `/api/v1/user` 映射到这两个地址。

---

## 2. JMeter 安装与配置

### 2.1 下载 JMeter

1. 访问 [Apache JMeter 官网](https://jmeter.apache.org/download_jmeter.cgi)。
2. 下载 **Binary** 版（如 `apache-jmeter-5.6.3.zip`）。
3. 解压到本地，例如 `D:\tools\apache-jmeter-5.6.3`。

### 2.2 打开测试计划文件

```bash
# GUI 模式（用于调试配置）
D:\tools\apache-jmeter-5.6.3\bin\jmeter.bat

# 然后菜单：File → Open → 选择 benchmark/Gateway_QuantitativeTest.jmx
```

### 2.3 创建报告输出目录

```bash
mkdir d:\Go-Gateway\Go-Gateway\benchmark\reports
```

---

## 3. 测试场景详解

### 场景 1：基准路由性能测试

**目的**：在正常并发下测量网关的 P99 延迟和吞吐量（QPS），作为性能基准数据。

**参数配置**：
| 参数 | 值 |
|:---|:---|
| 并发用户数 | 100 |
| Ramp-up 时间 | 10 秒 |
| 每用户请求数 | 200 |
| **总请求数** | **20,000** |

**测试路径**：
- `GET /api/v1/user` — 精确匹配
- `GET /api/v1/order/:id` — 参数化路由
- `GET /no/such/route` — 404 场景

**运行命令（非 GUI 模式，推荐）**：

```bash
cd d:\Go-Gateway\Go-Gateway

jmeter -n ^
  -t benchmark\Gateway_QuantitativeTest.jmx ^
  -l benchmark\reports\result_s1.jtl ^
  -e -o benchmark\reports\dashboard_s1 ^
  -Jnum_threads=100 ^
  -Jloops=200
```

**期望结果**：

| 指标 | 期望值 |
|:---|:---|
| 平均延迟（Average） | < 20ms |
| P99 延迟（99% Line） | < 100ms |
| 吞吐量（Throughput） | > 5000 QPS |
| 错误率（Error %） | < 1% |

---

### 场景 2：令牌桶限流触发验证

**目的**：确认当单 IP 的 QPS 超过设定阈值时，网关正确返回 `429 Too Many Requests`，且正常请求不受影响。

**操作步骤**：

1. 在 JMeter 中启用 **【场景2】令牌桶限流触发验证** 线程组（其他场景禁用）。
2. `Header Manager` 已配置固定 IP `10.0.0.1`，**200 并发**会远超正常 QPS 阈值。
3. 运行测试，观察汇总报告中的 **Error %**。

**关键量化点**：

```
Error % ≈ (200并发 - 限流QPS阈值) / 200并发 × 100%

例如：限流阈值 = 100 QPS，200 并发：
期望 Error % ≈ 50%（超出阈值的 100 个请求被拦截）
```

**非 GUI 运行**：

```bash
jmeter -n ^
  -t benchmark\Gateway_QuantitativeTest.jmx ^
  -l benchmark\reports\result_s2.jtl ^
  -e -o benchmark\reports\dashboard_s2
```

**验证指标**：

| 指标 | 含义 |
|:---|:---|
| Error % | 被限流拦截的请求比例（面试展示此数据） |
| 平均延迟（被限流请求） | 应 < 5ms（限流很快返回，延迟极低） |
| 状态码分布 | `200:100个`，`429:100个` 左右 |

---

### 场景 3：熔断降级效果验证

**目的**：验证当后端持续返回错误时，熔断器是否在达到阈值后自动开启，并在冷却期后尝试恢复。

**操作步骤**：

1. **手动关闭后端服务**（或让后端返回 `500` 错误）。
2. 运行场景 3：10 个线程，每个请求 20 次。
3. 观察"结果树"中的响应码变化：

```
初始阶段：502（后端连接失败，网关透传错误）
连续失败 N 次后：503（熔断器开启，网关主动拦截）
等待冷却期后：502 或 200（熔断器进入半开启，尝试探测）
后端恢复后：200（熔断器关闭，恢复正常）
```

**关键量化点（面试话术）**：

> "经 JMeter 测试，当后端连续失败 **3 次**后，熔断器在第 **X ms** 内自动开启；
> 冷却期 **1 秒**后进入半开启状态，发送探测请求；
> 后端恢复后，**X 个请求**内完成自愈。"

---

### 场景 4：负载均衡分布验证

**目的**：验证轮询算法是否将请求均匀分配到多个后端实例。

**前置准备**：
1. 启动两个或三个"假后端"，每个返回不同内容（如 `{"server": "A"}`）。
2. 配置网关路由规则，指向多个后端。

**量化方法**：

在 JMeter 的"察看结果树"中观察每个响应的 `Body`，统计 `server: A` 和 `server: B` 出现的次数：

```
总请求：3000
server: A 命中：1502（50.07%）
server: B 命中：1498（49.93%）

→ 轮询分布偏差 < 1%，验证成功。
```

---

### 场景 5：极限压力测试

**目的**：模拟 500 并发持续 30 秒的流量冲击，验证网关稳定性并获取极限性能数据。

> ⚠️ **注意**：请在独立机器或 Docker 环境中运行，避免影响本地开发环境。

**运行命令**：

```bash
jmeter -n ^
  -t benchmark\Gateway_QuantitativeTest.jmx ^
  -l benchmark\reports\result_stress.jtl ^
  -e -o benchmark\reports\dashboard_stress ^
  -Jnum_threads=500 ^
  -Jduration=30
```

**关注指标**：

| 指标 | 预期 | 含义 |
|:---|:---|:---|
| 吞吐量 | > 10,000 QPS | 网关极限处理能力 |
| P99 延迟 | < 100ms | 长尾请求控制 |
| P95 延迟 | < 50ms | 大多数请求性能 |
| 错误率 | < 0.1% | 稳定性指标 |
| CPU 占用 | < 80% | 资源利用合理 |

---

## 4. 如何读懂量化报告

### 4.1 生成 HTML 报告

非 GUI 模式运行时，使用 `-e -o` 参数自动生成 HTML 图表报告：

```bash
jmeter -n -t benchmark\Gateway_QuantitativeTest.jmx ^
       -l benchmark\reports\result.jtl ^
       -e -o benchmark\reports\dashboard
```

用浏览器打开 `benchmark\reports\dashboard\index.html` 查看图表。

### 4.2 核心指标说明

```
Throughput (QPS/TPS)
  └─ 每秒处理的请求数，网关的"核心战斗力"指标
     ★ 面试量化：本网关在 100 并发下 QPS 达到 X，优于对比方案 Y

Average（平均延迟）
  └─ 所有请求的平均响应时间（ms）
     ⚠ 注意：平均值可能被少数慢请求拉高，不代表实际体验

P99 延迟（99% Line）
  └─ 99% 的请求都在此时间内完成，代表"最坏正常情况"
     ★ 面试量化：99% 的请求延迟 < X ms，满足 SLA 要求

P95 延迟（95% Line）
  └─ 95% 请求完成时间，通常比 P99 更有参考价值

Error %（错误率）
  └─ 请求失败的比例
     ★ 在限流场景下，Error% 代表限流触发率（有意义！）
     ★ 在正常场景下，Error% 应 < 0.1%

Min / Max
  └─ 最快/最慢请求时间，用于诊断偶发异常
```

### 4.3 从 .jtl 文件提取 P99 数据

```bash
# 使用 JMeter 命令行从 .jtl 生成报告
jmeter -g benchmark\reports\result.jtl -o benchmark\reports\dashboard_from_jtl
```

---

## 5. Go 内置 Benchmark 使用

### 5.1 运行路由引擎基准测试

```bash
cd d:\Go-Gateway\Go-Gateway\internal\engine

# 运行所有 Benchmark，输出纳秒级延迟
go test -bench=. -benchmem -count=3

# 运行指定 Benchmark
go test -bench=BenchmarkRouter_Search -benchmem -count=3 -benchtime=5s
```

**预期输出**：

```
goos: windows
goarch: amd64
BenchmarkRouter_Search-8    68543210    15.3 ns/op    0 B/op    0 allocs/op
```

**解读**：

| 字段 | 含义 |
|:---|:---|
| `68543210` | 每秒可执行的基准测试迭代次数 |
| `15.3 ns/op` | **每次路由匹配耗时 15.3 纳秒**（极致性能！） |
| `0 B/op` | 每次操作内存分配为 0（零内存分配，GC 友好） |
| `0 allocs/op` | 每次操作无堆内存分配 |

**面试话术**：

> "通过 `go test -bench` 量化测试，我们的 Radix Tree 路由引擎在单次匹配中耗时约 **15ns**，内存分配为零，实现了无 GC 压力的高性能路由。"

### 5.2 使用 pprof 进行 CPU 分析

```bash
# 生成 CPU profile
go test -bench=BenchmarkRouter_Search -cpuprofile=cpu.prof

# 可视化分析
go tool pprof cpu.prof
(pprof) top10
(pprof) web   # 需要安装 Graphviz
```

---

## 6. 结合 Prometheus 实时监控

> **前提**：在本地安装 Prometheus 并启用网关的 `/metrics` 端点。

### 6.1 关键指标查询语句（PromQL）

```promql
# 实时 QPS
rate(gateway_http_requests_total[1m])

# P99 延迟
histogram_quantile(0.99, rate(gateway_http_request_duration_seconds_bucket[5m]))

# P95 延迟
histogram_quantile(0.95, rate(gateway_http_request_duration_seconds_bucket[5m]))

# 各状态码分布
sum by (status) (rate(gateway_http_requests_total[1m]))

# 错误率
rate(gateway_http_requests_total{status=~"5.."}[1m]) / rate(gateway_http_requests_total[1m])
```

### 6.2 JMeter + Prometheus 联动测试流程

```
1. 启动网关 → 2. 启动 Prometheus 采集 → 3. JMeter 发压 → 4. 对比两侧数据

JMeter 报告（客户端视角）：看到的是端到端延迟
Prometheus 数据（服务端视角）：看到的是网关内部处理延迟

两者差值 ≈ 网络传输耗时 + 序列化耗时
```

---

## 7. 量化数据填写模板

在完成测试后，请将实测数据填入以下表格，用于简历/项目展示：

```
===================================================
Go-Gateway 性能量化测试报告
测试时间：____年__月__日
测试机器：CPU __ 核，内存 __ GB，OS __
===================================================

【场景 1：基准路由性能】
  并发数：100，总请求：20,000
  - 吞吐量 (QPS)：________ req/s
  - 平均延迟：________ ms
  - P99 延迟：________ ms
  - P95 延迟：________ ms
  - 错误率：________ %

【场景 2：令牌桶限流触发率】
  并发数：200，限流阈值：100 QPS
  - 被拦截请求（429）比例：________ %
  - 被限流请求平均响应时间：________ ms（体现限流快速响应）
  - 正常请求平均响应时间：________ ms

【场景 5：极限压力测试】
  并发数：500，持续：30 秒
  - 峰值 QPS：________ req/s
  - P99 延迟：________ ms
  - 错误率：________ %
  - 内存占用峰值：________ MB

【Go Benchmark：路由引擎】
  - 单次匹配耗时：________ ns/op
  - 内存分配：________ B/op
  - 并行 Benchmark：________ ns/op
===================================================
```

---

## 快速命令速查

```bash
# 1. 一键完整测试（非 GUI 模式）
jmeter -n -t benchmark\Gateway_QuantitativeTest.jmx -l reports\result.jtl -e -o reports\dashboard

# 2. 仅运行 Go 单元测试
go test ./internal/engine/... -v

# 3. 运行 Benchmark 并输出完整报告
go test ./internal/engine/... -bench=. -benchmem -benchtime=10s -count=3

# 4. 生成已有 .jtl 的 HTML 报告
jmeter -g reports\result.jtl -o reports\html_report

# 5. 运行 stress_test.go 压测工具
go run scripts/stress_test.go
```
