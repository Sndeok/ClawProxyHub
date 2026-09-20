# Changelog

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