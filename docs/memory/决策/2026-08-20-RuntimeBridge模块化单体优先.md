# RuntimeBridge 模块化单体优先

- 状态：已确认
- 日期：2026-08-20

## 结论

XCode 采用以下稳定边界，当前继续单进程部署：

```text
XCode ProductCore
        |
RuntimeBridge v1 contract
        |
Sub2API Driver
        |
上游服务
```

Sub2API 是第一个 Runtime Driver，负责账号池、OAuth 刷新、协议适配、调度、重试、冷却、失败切换和上游请求。ProductCore 负责平台、模型能力、API Key 授权、订阅、余额、倍率、定价、扣费和使用记录。Runtime 不接收 API Key、订阅、余额、套餐或 Ent 实体，只接收标量关联 ID 与不可变路由快照。

## 原因

现有账号仓库、OAuth 状态、调度缓存、协议转换和用量终结器仍在同一 Go 进程内协作。现在直接拆成独立进程会把隐式依赖变成网络鉴权、重试、一致性和流式断线问题。先建立包级 contract、LocalRuntime 和可替换 Driver，可以在不改变数据库、HTTP API、前端、Docker 和计费规则的前提下提高可替换性。

## 影响

- 默认部署仍是一个应用容器，不新增 Runtime 容器、RPC 或数据库迁移。
- OpenAI Chat、Responses、Messages 已开始通过纯 Sub2API Driver 迁移；同步、媒体、Live、WebSocket 等未完成端点必须显式列入延期集合，不得静默回退。
- 未来只有在所有生产端点不依赖 Gin/公开 Handler、RuntimeStore/ControlPort、HTTP/SSE conformance、服务鉴权和单进程/外置黑盒等价全部满足后，才允许外置 RuntimeBridge。
- 暂不切换 CLIProxyAPI；新增 Driver 只能实现同一 contract 和 conformance tests，不得修改 ProductCore、前端或业务 schema。

## 相关文件

- `docs/superpowers/specs/2026-08-20-runtimebridge-sub2api-separation-design.md`
- `docs/superpowers/plans/2026-08-20-runtimebridge-sub2api-separation.md`
- `backend/pkg/runtimebridge/v1/`
- `backend/internal/runtimebridge/`
- `backend/internal/runtime/sub2api/`
- `backend/internal/architecture/runtime_sub2api_purity_guard_test.go`
