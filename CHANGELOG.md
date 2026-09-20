# Changelog

## 二开改动（Sndeok fork，基于 v1.0.2）

面向「Codex → new-api → cph → 上游」链路的排查与修复，核心侧改动只需替换二进制即可生效。

- Fixed 工具调用（function calling）链路完全不可用：OpenAI 方言上游只在首个增量带
  `tool_calls[].id/name`，`sdk/openaiup` 把后续增量的空 id 原样透传，而核心按 id 分组，
  导致同一调用被拆成「空 id 的新块」，客户端拿到残缺 tool_calls（表现为一般对话正常、
  一涉及查看链接等工具操作就没回复）。新增 `internal/gateway/toolcall.go` 统一归一，
  三协议（chat_completions / messages / responses）编码器共用；老插件无需重编即恢复。
- Fixed `/v1/responses` 流式缺 `response.output_item.done`（function_call）：Codex CLI
  只认 done 事件里的工具调用，缺失时工具永不执行。同时补齐
  `response.function_call_arguments.done`，`output_index` 改为按项递增，
  `response.completed` 带上 `output` 与 `usage`，`output_text.done` 回填完整文本。
- Fixed 插件宿主回调（日志 / 状态存储 / 读设置）在插件启动 5 秒后永久失效：
  go-plugin 的 broker 发出 ConnInfo 后只保留 5 秒，SDK 懒加载拨号必然错过，
  且失败被 `sync.Once` 缓存成终态；首次请求还会白等 5 秒。改为启动阶段异步
  warmup + 失败退避重试（需重新构建插件）。
- Fixed 插件子进程 stderr 被 go-plugin 吞掉（只记「收到 N 字节」）：现按行转发为核心
  日志并加 `[plugin:<name>]` 前缀，插件自身诊断不再丢失。
- Fixed 日志「总耗时」口径错误：原由各调用方传入局部 `time.Since(start)`，流式请求
  只统计首事件之后的输出阶段，出现「首字 5001ms、总耗时 1ms」。统一以 serve 入口为
  基准，在 `requestLogCtx.write` 内计算。
- Added 调用日志分页与多维过滤：`GET /admin/logs` 支持 page/page_size（≤200）、
  key_id、plugin_id、account_id、protocol、model、status(ok/error/4xx/5xx)、q、
  from/to、min_latency，返回 total/has_more；旧 `limit` 参数仍兼容。
  同时消除旧实现「每查一页日志就全表扫 keys」的 O(n) 开销（改为只反查当前页出现的 id）。
- Added 迁移 000003：`request_logs` 增加 requested_model（请求模型 / 路由别名）、
  route_id、group_id、stream、finish_reason、attempts（含重试换号降级的尝试次数）、
  error_type 与两个过滤索引。
- Added 日志保留策略 `logs.retention_days`（0 = 永久保留）+ 手动清理
  `POST /admin/logs/cleanup`（N 天前 / 全部），后台每 6 小时自动清理一次。
- Added 插件市场出站代理 `network.market_proxy`（也可用 `CPH_MARKET_PROXY` 设默认）：
  拉取索引与下载 .cphplugin 共用，支持 socks5 / socks5h / http / https，
  省略协议头按 socks5 处理、域名交代理解析；与原有 `network.github_proxy`
  （URL 前缀改写）并存。新增 `POST /admin/settings/test-market` 供页面自检连通性。
- Changed 默认插件市场地址改为自建插件仓库
  `https://raw.githubusercontent.com/Sndeok/ClawProxyHubPlugins/main/index.json`。
- Changed 仪表盘日志页重写：过滤条、服务端分页、请求模型/上游模型/流式/尝试次数列、
  行详情抽屉、自动刷新、清理日志；设置页新增日志保留天数与市场代理（含测试连接）。

## v1.0.2

- Fixed `CPH_SEED_API_KEY` 每次重启重复创建密钥：AES-GCM 随机 nonce 导致等值查重永不命中；改用 SHA-256 hash 判重
- Fixed API 密钥鉴权每次请求全表扫描解密（O(n)）：`keys` 表新增 `key_hash` 等值索引列，快路径 O(1) 命中，存量密钥自动回填 hash
- Fixed 网关重试/超时后插件事件流 goroutine 泄漏：gRPC Chat 调用改为传入可取消 context，重试前 cancel 旧流并后台排空 channel
- Fixed 插件 KV 状态存储全部插件共享且不持久化：每个插件实例持有独立 HostService，按插件名隔离 key 空间，持久化到 `plugin_storage` 表
- Fixed 删除分组后路由 `groups_json` / `failover_group_id` 残留引用：删除时事务式清理路由配置、降级指向、账号关联与代理绑定
- Fixed 流式响应中途上游失败客户端无感知：TaskFailed 时在 SSE 关闭前发出错误事件（含错误类型/消息/状态码）
- Fixed 请求体超 32MB 静默截断：预检 `Content-Length` + chunked 尾部检测，超限返回 413
- Fixed `truncStr` 截断中文产生无效 UTF-8：统一改用 `util.TruncStr`，回退到 rune 边界保证不切半个字符
- Fixed Windows 插件安装/升级时二进制文件锁导致删除失败：`removeWithRetry` 重试 3 次 × 200ms
- Added 出站代理参数校验：scheme 限定 `http`/`https`/`socks5`，port 限定 1–65535
- Added 数据库迁移 000002：`keys.key_hash` 列（带索引）+ `plugin_storage` 表（插件持久化 KV），存量数据自动兼容

## v1.0.1

- Fixed Codex CLI（`/v1/responses`）请求失败（表现为 502 / 上游 500）：`extractText` 未识别 `input_text` / `output_text` 内容块，导致消息文本被静默丢空、上游收到空请求
- Fixed `developer` 角色（Responses / Chat Completions）被透传给只认 `system/user/assistant/tool` 的上游而被拒；现归一为 `system`
- Fixed Codex 流式中断 `missing field input_tokens`：`response.completed` 的 `usage` 缺失时补零值（`input_tokens` / `output_tokens` / `total_tokens`）
- Added 从 Responses 请求提取 `reasoning.effort` 并透传上游 `reasoning_effort`（此前被丢弃）
- Added `TestParseResponsesRequestCodex` 回归测试，复刻 Codex CLI 真实请求

## v1.0.0

- 首个正式版本：Claw 类客户端统一管理反代网关，单二进制核心 + 插件化实现
- 三协议归一化入口（`/v1/messages`、`/v1/chat/completions`、`/v1/responses`），任意入口 × 任意上游方言全矩阵转换
- 路由（模型别名 → 分组 → 账号，多策略）、401 刷新换号、4xx/5xx 降级、首事件超时
- 账号多步登录、AES-256-GCM 凭据加密、429/无积分自动暂停、分组级出站代理
- 任务调度（interval/daily/once）、插件市场（在线索引 + sha256 校验 + 离线兜底）
- Vue3 + TDesign 双语暗色仪表盘，Docker 部署，SQLite + golang-migrate 自动迁移