# 🚀 Nexus Gateway (高并发分布式 API 网关)

Nexus Gateway 是一个基于 Go (Gin) + Redis + MySQL 构建的轻量级、高性能分布式 API 网关系统。该项目专为微服务架构设计，提供动态路由、分布式限流、双层安全黑名单拦截以及丰富的数据观测面板。是实习/求职展示个人对高并发、架构设计及中间件应用能力的优秀实战项目。

## ✨ 核心特性 (Core Features)

* **🚦 动态路由转发**: 基于 `httputil.ReverseProxy` 实现，支持路由规则（前缀/精确匹配）的热更新，数据由 Redis 提供毫秒级缓存读取。
* **🛡️ 分布式安全防线 (Blacklist)**: 支持动态拉黑 IP 和请求路径（精确或前缀匹配）。高并发下直接命中 Redis 集合，非法请求在网关层即被拦截 (403 Forbidden)。
* **⏱️ 双层分布式限流 (Rate Limiting)**: 
  * **第一层（稍微频繁）**: 触发 429 Too Many Requests 限流，保护后端服务不被打垮。
  * **第二层（过于频繁）**: 触发系统**自动拉黑**机制，该客户端 IP 会被直接扔进 Redis 黑名单池，并异步落库。管理台支持一键关闭或调节双层阈值。
* **📊 全链路监控与日志 (Metrics & Observability)**: 网关层自带 Metrics 统计中间件，实时计算 QPS、拦截次数及状态码分布。拦截动作均支持异步记录到 MySQL 中，供后台安全审计。
* **💻 沉浸式管理控制台 (Admin Dashboard)**: 内置一个高级感满满（Glassmorphism 拟物毛玻璃风格）的前端控制台，支持管理规则、查看实时流量大盘与拦截明细。
* **🐳 极简部署 (Docker Ready)**: 提供极致优化的多阶段构建 (Multi-stage build) Dockerfile 及 `docker-compose.yml`，一键拉起网关、管理端、MySQL、Redis 及 Nginx 反向代理集群。

---

## 🏗️ 架构设计 (Architecture)

请求在到达网关侧的核心处理流水线 (Pipeline) 如下：
`[Client] -> [Nginx] -> [Gateway Metrics] -> [Security Blacklist] -> [Distributed Rate Limit] -> [Dynamic Proxy] -> [Upstream Services]`

* **Go (Gin)**: 承载极高的并发，协程处理请求，内存开销极低。
* **Redis**: 充当所有热点数据的缓冲层（路由表、黑名单集合、限流计数器），保证了网关转发的极致低延迟。
* **MySQL**: 作为持久化配置中心。后台配置变更后，系统自动同步更新 Redis。

---

## 🚀 快速启动 (Quick Start)

项目提供了极其简便的 Docker Compose 部署方案，只需两步即可在本地启动完整的网关栈。

### 1. 域名配置
为了方便本地模拟生产环境，请修改本机的 `hosts` 文件（Windows: `C:\Windows\System32\drivers\etc\hosts`，Linux/Mac: `/etc/hosts`），添加以下域名映射：
```text
127.0.0.1   gateway.nym.asia
127.0.0.1   manage.nym.asia
```

### 2. 容器启动
在项目根目录执行：
```bash
docker-compose up -d
```

启动完成后：
* **业务网关入口**: `http://gateway.nym.asia` (将根据你在后台配置的规则进行转发)
* **管理控制台**: `http://manage.nym.asia`
  * 默认管理员账号：`admin`
  * 默认管理员密码：`admin123`

---

## 📂 目录结构 (Directory Structure)

```text
Go-Gateway/
├── admin-frontend/       # Vue 3 原生前端管理界面（极光/毛玻璃UI）
├── cmd/                  # main.go 入口文件
├── conf/                 # Nginx 等外部组件配置文件
├── internal/
│   ├── config/           # 环境变量与应用配置
│   ├── controllers/      # 管理台 API 接口 (CRUD、系统配置、登录)
│   ├── middleware/       # 网关核心中间件 (鉴权、黑名单、限流、代理、Metrics)
│   ├── models/           # 数据库 GORM 实体模型定义
│   └── routes/           # Gin 路由编排与拦截器流水线注册
├── pkg/
│   ├── jwt/              # JWT Token 工具
│   ├── mysql/            # MySQL 连接池与自动迁移
│   └── redis/            # Redis 客户端配置
├── scripts/              # 数据库初始化脚本
├── Dockerfile            # Go 应用的多阶段构建镜像打包文件
└── docker-compose.yml    # 服务全家桶一键编排文件
```

---

## 🧪 压测建议 (Benchmarking)

得益于 Go 语言特性和 Redis 的内存级命中，系统占用资源极低（百兆级内存消耗）。你可以使用 `wrk` 工具对其限流和代理性能进行验证：
```bash
# 模拟 100 个并发连接，持续 30 秒，发起高频请求触发拉黑防线
wrk -c 100 -t 10 -d 30s http://gateway.nym.asia/api/test
```
在压测时，可以打开 `manage.nym.asia` 管理后台，实时观察 QPS 攀升以及被系统“自动拉黑”的 IP 名单。
