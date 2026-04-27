# Nexus Gateway 性能压测指南

## 前置要求

- [wrk](https://github.com/wg/wrk) 压测工具已安装
- 网关服务已启动且可访问

## 基础压测

```bash
# 使用默认参数 (100 并发, 30s 持续, 4 线程)
./benchmark/bench.sh

# 自定义参数
./benchmark/bench.sh -u http://localhost:8080/api/test -c 200 -d 60s -t 8
```

### 参数说明

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `-u` | 目标 URL | `http://localhost:8080` |
| `-c` | 并发连接数 | `100` |
| `-d` | 测试持续时间 | `30s` |
| `-t` | 线程数 | `4` |

## 限流压测

使用 wrk Lua 脚本模拟不同 IP 的并发请求，验证滑动窗口限流的准确性：

```bash
wrk -t4 -c200 -d30s -s benchmark/bench_ratelimit.lua http://localhost:8080/api/test
```

输出示例：

```
========= Rate Limit Benchmark Results =========
Total Requests:     125000
  200 OK:           98000 (78.4%)
  429 Rate Limited: 26500 (21.2%)
  403 Blacklisted:  500 (0.4%)
  Other:            0 (0.0%)
==================================================
```

## 报告模板

每次压测报告应包含以下指标：

| 指标 | 说明 | 目标值 |
|------|------|--------|
| **QPS** | 每秒请求数 | > 10,000 |
| **P50 延迟** | 50% 请求的延迟 | < 5ms |
| **P99 延迟** | 99% 请求的延迟 | < 50ms |
| **错误率** | 非 2xx/429 响应占比 | < 0.1% |
| **内存占用** | 网关进程 RSS | < 100MB |

## 报告存放

压测报告自动保存到 `benchmark/reports/` 目录，文件名格式：`bench_YYYYMMDD_HHMMSS.txt`。
