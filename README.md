> ## ⚠️ 本仓库已废弃（DEPRECATED）
>
> 这里是最早的 **Vue 3 + TDesign** 版本，只保留历史参考，**不再接收功能更新与修复**。
>
> 请改用新仓库 👉 **[Sndeok/ClawProxyHub-Next](https://github.com/Sndeok/ClawProxyHub-Next)**
>
> 新仓库包含：Next.js 重构前端、账号新增向导、路由/账号编辑弹窗、移动端适配、
> Codex/new-api 协议兼容修复（错误帧、参数透传、缓存写入 token 统计），Docker 部署方式一致。
>
> This repository is **deprecated** and kept for historical reference only. Use
> [Sndeok/ClawProxyHub-Next](https://github.com/Sndeok/ClawProxyHub-Next) instead.

<div align="center">

# ClawProxyHub（旧版，已废弃）

**把官方 AI 桌面客户端，变成你自己的 OpenAI 兼容网关**

一个插件化的 AI 反代网关：用官方客户端账号登录，对外提供 OpenAI Chat Completions / Responses /
Anthropic Messages 三种协议；账号池、模型路由、密钥、调用日志、定时任务一站式管理。

`Go 1.24` · `Vue 3 + TDesign` · `gRPC 插件` · `SQLite` · `Docker`

</div>

---

## 这是什么

ClawProxyHub-Next 是 [ClawProxyHub](https://github.com/Sndeok/ClawProxyHub) 的二开重构版（下称"本项目"），
上游为 [ShadowSmallBaby/ClawProxyHub](https://github.com/ShadowSmallBaby/ClawProxyHub)。

它解决的问题很具体：**官方桌面客户端（LobsterAI / WorkBuddy 等）的账号，能不能变成标准 API 用？**

答案是能，而且做得比"能"多一点：

- 官方客户端怎么登录，你就怎么登录（手机验证码 / 浏览器授权 / 凭据文件），凭据由本项目加密托管
- 对外只暴露标准协议，`Codex CLI`、`Claude Code`、`CC Switch`、`new-api` 这类客户端直接对接
- 多账号自动轮换、按分组加权、失败降级、会话粘性、快过期积分优先，都是配置项
- 每次调用都有结构化日志：走的哪个账号、命中哪条路由、首字耗时、总耗时、缓存命中、重试次数

## 和上游的关系

本项目是**独立演进**的分支，不是简单的镜像同步：

| 维度 | 说明 |
|---|---|
| 协议 | 向上游兼容（`sdk/proto/cph.proto`），插件仓库也用同一份 SDK |
| Go module | 仍为 `github.com/ShadowSmallBaby/ClawProxyHub`，**故意不改**——改了所有已发布插件都要重发 |
| 迁移编号 | 已发布的迁移只增不改；上游新增 schema 一律按序追加新编号（避免存量库"永远执行不到"） |
| 差异 | 见 [CHANGELOG.md](./CHANGELOG.md)，逐条记录了与上游的每一处分歧及原因 |

## 功能一览

### 账号与凭据
- 多插件账号池：登录 / 刷新 / 过期标记 / 自动重新登录
- 账号级出站代理（优先级：账号 > 分组），支持 `socks5` / `socks5h` / `http(s)`
- 积分快照：剩余 / 总 / 已用，**分池明细与到期时间**（账号页可直接看到"快过期积分"倒计时）
- 今日消耗：按调用日志聚合 token + 积分（积分优先取逐次上报，缺失时用余额差值估算）

### 路由
- 对外模型名（别名）→ 分组（权重）→ 账号（策略）三级解析
- 负载策略：`轮询` / `随机` / `最少使用` / `会话粘性` / `过期优先`
- 会话粘性：同一会话固定账号，多轮对话更连贯，上游 prompt 缓存命中率更高（保持时长与清理周期可配）
- 失败降级：按 4xx / 5xx 触发，切到备用分组与备用模型
- 一键同步上游模型：把账号可见模型并集补齐为路由（默认会话粘性）

### 网关
- 三种协议入口 + 协议间自动转换（Responses ↔ Chat ↔ Messages）
- 多模态内容贯通：图片 / 文件 / 音频 / 视频按 content block 透传
- 工具调用链路：跨协议重组 function_call，增量参数按 index 归位
- 用量口径统一：`input_tokens` 含缓存命中，`cached_tokens` 是其中的子集
- 首字超时（全局 + 路由级）、重试与换号、流式错误下发

### 可观测
- 调用日志：服务端分页 + 多维过滤（时间 / 状态 / 协议 / 密钥 / 插件 / 账号 / 模型 / 关键词 / 慢请求）
- 日志保留策略：按天自动清理（每 6 小时巡检）+ 手动清理入口
- 概览：今日请求、成功率、总 Token、账号/插件/密钥计数、请求趋势、模型分布、渠道积分概览、最近请求（分页）

### 运维
- 任务调度：签到等维护任务，支持周期 / 每日 / 一次性；一键全部执行 + 执行历史
- 插件市场：自建仓库索引，支持 SOCKS5/HTTP 代理拉取与下载
- 出站标识按插件配置（UA / 客户端名称 / 版本 / CLI 版本），默认对齐官方分发包
- 会话粘性策略、出站标识都支持"保存即生效"，不用重启

## 架构

```
                 ┌──────────── 客户端 ────────────┐
                 │ Codex CLI / Claude Code / ... │
                 └───────────────┬───────────────┘
             /v1/chat/completions│/v1/responses│/v1/messages
                                 ▼
┌──────────────────────────────────────────────────────────────┐
│                        ClawProxyHub-Next                     │
│                                                              │
│  鉴权(密钥)  →  路由解析(Route→Group→Account)  →  出站代理    │
│       │                    │                        │         │
│       │            会话粘性 / 过期优先              │         │
│       ▼                    ▼                        ▼         │
│   调用日志          信封协议(protobuf)          gRPC 插件     │
│  (SQLite)                                          │         │
└────────────────────────────────────────────────────┼─────────┘
                                                     ▼
                              LobsterAI / WorkBuddy / ... 官方上游
```

核心进程只做编排：**登录、路由、记账、日志**。真正的协议适配在上游侧由插件完成，
所以新增一个上游 = 新增一个插件，不用改核心。

## 快速开始

### Docker Compose（推荐）

```bash
git clone https://github.com/Sndeok/ClawProxyHub.git
cd ClawProxyHub
# 首次运行会创建数据目录与初始管理员密码
docker compose up -d --build
```

打开 `http://<服务器>:8080`，按引导设置管理员密码，然后：

1. **插件页** → 从市场安装 `LobsterAI` / `WorkBuddy`
2. **账号页** → 添加账号（扫码 / 验证码 / 凭据文件）
3. **路由页** → 「同步上游模型」一键把账号可见模型建成路由
4. **密钥页** → 创建调用密钥
5. 客户端里把 Base URL 指到 `http://<服务器>:8080/v1`，API Key 填刚创建的密钥

### 本地开发

```bash
# 后端（需要 CGO，SQLite 驱动依赖）
go run ./cmd/cph

# 前端（Vite dev server，代理到 :8080）
cd web && pnpm install && pnpm dev
```

前端产物通过 `go:embed` 打进二进制，改了前端记得 `pnpm build`。

## 环境变量

| 变量 | 默认 | 说明 |
|---|---|---|
| `CPH_ADDR` | `:8080` | 监听地址 |
| `CPH_DATA_DIR` | `./data` | 数据目录（SQLite / 插件 / 密钥） |
| `CPH_DATABASE_DSN` | `<data>/cph.db` | 数据库 DSN |
| `CPH_PLUGIN_DIR` | `<data>/plugins` | 插件安装目录 |
| `CPH_MARKETPLACE_URL` | 自建仓库 index.json | 插件市场地址（首次启动写入设置，之后以界面为准） |
| `CPH_MARKET_PROXY` | 空 | 市场出站代理（`socks5://host:port` / `http://host:port`） |
| `CPH_ADMIN_PASSWORD` | 空 | 首次启动引导管理员密码（之后在设置页修改） |

界面里可配的项（设置页）：网关首字超时、日志保留天数、插件市场地址与代理、GitHub 加速代理、
会话粘性策略、以及**按插件**的出站标识。

## 目录结构

```
cmd/cph/            进程入口：装配数据库 / 插件 / 路由 / 网关 / 管理后台
internal/
  account/          账号域：登录、刷新、积分快照、凭据加解密
  admin/            管理后台 API + 内嵌前端资源
  database/         SQLite 打开与迁移（migrations/*.sql，已发布编号只增不改）
  gateway/          协议入口与出口：chat / responses / anthropic 三种转换
  model/            GORM 模型
  plugin/           插件宿主：子进程、gRPC broker、市场安装
  router/           Route → Group → Account 解析 + 会话粘性
  setting/          设置 KV（含热生效的出站标识 / 粘性策略）
  task/             任务调度引擎
sdk/                插件 SDK（proto + 上游适配器：openaiup / anthropicup）
web/                Vue 3 前端（TDesign + 自研设计令牌）
```

## 插件

插件是独立进程，通过 gRPC 与本项目通信，SDK 就是本仓库的 `sdk/` 目录。

- 官方插件仓库：[Sndeok/ClawProxyHubPlugins](https://github.com/Sndeok/ClawProxyHubPlugins)
- 插件用 `sdk/openaiup` / `sdk/anthropicup` 做上游协议适配，用 `sdk/sdk.go` 起服务
- 新增上游厂商 = 实现 `pb.ClawPluginServer`，声明鉴权方式与任务能力，打包成 `.cphplugin`

> 改了 `sdk/`（例如用量解析）必须重新编译插件，否则插件二进制里还是老逻辑。

## 安全说明

- 账号凭据以 AES-256-GCM 加密落库，密钥文件在数据目录（`.cph/key`），**务必备份**
- 管理后台走 Bearer Token（管理员密码 bcrypt 存储）
- 出站标识字段做了换行与长度校验，避免请求头注入
- 本项目用于个人账号的自动化代理，请自行评估上游服务条款

## 许可

与上游一致，见 [LICENSE](./LICENSE)。