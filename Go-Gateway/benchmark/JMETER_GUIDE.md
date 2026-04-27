# Nexus Gateway JMeter 测试指南

这份指南将教你如何利用 JMeter 对网关进行深度的性能测试，并量化优化成果。

## 1. 测试准备
- **工具**: 安装 [Apache JMeter](https://jmeter.apache.org/download_jmeter.cgi)。
- **启动**: 在 JMeter 中打开 `benchmark/Nexus_Gateway_Test_Plan.jmx`。
- **环境**: 确保 Redis 和 Go-Gateway 已在本地或 Docker 中运行。

## 2. 核心测试场景

### 场景 A: 滑动窗口限流稳定性测试
- **目的**: 验证 Lua 原子限流是否能平滑处理并发请求。
- **配置**:
    1. 在 `模拟不同IP` (Header Manager) 中，将 Header 的值改为固定 IP（例如 `1.1.1.1`）。
    2. 运行脚本，观察“汇总报告”中的 Error %。
- **预期**: 当请求频率超过 Redis 中配置的 `slightly_freq_threshold` 时，JMeter 应该开始记录 `429` 响应。

### 场景 B: Trie 树 vs Redis 遍历 性能对比
- **目的**: 量化 Trie 树在海量路由下的 O(K) 优势。
- **步骤**:
    1. 通过 Admin API 批量插入 500 条路由。
    2. 分别使用旧代码（Redis HGetAll）和新代码（Trie）运行相同的 200 并发压测。
    3. 对比 JMeter 报告中的 `Average` 和 `99% Line` (P99 延迟)。
- **简历量化建议**: "在大规模路由场景下（500+规则），将路由匹配耗时由线性增长优化为 O(1) 内存操作，P99 延迟降低了 X ms。"

### 场景 C: 热更新同步延迟
- **目的**: 测试 Redis Pub/Sub 同步路由的速度。
- **步骤**:
    1. 保持 JMeter 对一个未定义的 URL 发送请求（返回 404）。
    2. 运行中手动调用 API 创建该路由。
    3. 观察 JMeter 结果树，记录从 404 变为 200 的瞬间。

## 3. 常用性能指标速查
- **Throughput (TPS)**: 每秒处理的请求数，代表网关吞吐量。
- **Latency (P99)**: 99% 的请求都在此耗时以内，代表长尾响应质量。
- **Error %**: 拦截比例。在限流测试中，这代表限流器的触发频率。

---
*提示：建议在非 GUI 模式下进行正式压测以获得最准数据：*
`jmeter -n -t benchmark/Nexus_Gateway_Test_Plan.jmx -l benchmark/reports/result.jtl -e -o benchmark/reports/dashboard`
