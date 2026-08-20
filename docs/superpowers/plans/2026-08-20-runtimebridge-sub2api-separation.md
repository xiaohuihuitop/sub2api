# RuntimeBridge 与 Sub2API Driver 分离实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans (recommended) to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在不改变当前 HTTP API、数据库、前端、计费和 Sub2API 行为的前提下，把 Runtime contract 提升为可独立版本化的边界，并将 Sub2API 收口为可替换 Driver。

**Architecture:** 先保持单进程部署。新增 `pkg/runtimebridge/v1` 作为纯 Go contract，XCode 通过 `internal/runtimebridge` 调用本地 Driver；Sub2API executor 从 Handler 组合中逐步迁移到 `internal/runtime/sub2api`。只有所有端点通过 contract conformance tests 后，才允许另行实现 HTTP/SSE 外置 Bridge。

**Tech Stack:** Go 1.26、标准库 `context`/`net/http`/`encoding/json`、现有 `gatewayruntime`/`applicationgateway`、Gin 仅限 ingress、现有 Sub2API Service、Go unit/integration tests。

---

## 文件地图

| 单元 | 新增/修改文件 | 责任 |
| --- | --- | --- |
| 公共 contract | `backend/pkg/runtimebridge/v1/` | 只放可跨进程、可跨仓库使用的值对象和事件 |
| 内部兼容层 | `backend/internal/gatewayruntime/` | 在迁移期为旧调用提供别名/转换，不新增业务字段 |
| Bridge 组合 | `backend/internal/runtimebridge/` | ProductCore 决策到 Driver 请求的转换、本地执行和终态投递 |
| Sub2API Driver | `backend/internal/runtime/sub2api/` | Endpoint registry、executor 和 Sub2API Service 端口适配 |
| 产品组合 | `backend/internal/applicationgateway/`、`backend/internal/handler/` | Handler ingress 和 ApplicationGateway 只依赖 Runtime contract |
| 架构守卫 | `backend/internal/architecture/` | 防止 Gin、Ent、旧 Handler、产品实体回流 Runtime 边界 |
| 契约/回归测试 | 各新增包的 `*_test.go` 及现有 Handler/Service 测试 | 固定行为与迁移等价性 |

## 任务 0：建立基线与迁移清单

> 执行状态：已完成（基线、端点清单和迁移文档已建立；历史提交已保留）。

**Files:**
- Read: `docs/superpowers/specs/2026-08-20-runtimebridge-sub2api-separation-design.md`
- Read: `backend/internal/architecture/sub2api_adapter_purity_guard_test.go`
- Read: `backend/internal/handler/sub2api_runtime_composition.go`
- Read: `backend/internal/handler/sub2api_runtime_adapter.go`
- Test: `backend/internal/gatewayruntime/`, `backend/internal/applicationgateway/`

- [ ] **Step 1: 记录当前测试基线**

Run from `backend`:

```powershell
$env:GOFLAGS='-p=1'
$env:GOMAXPROCS='2'
go test -tags=unit ./internal/gatewayruntime ./internal/applicationgateway ./internal/architecture -count=1 -timeout=20m
```

Expected: all selected packages PASS. If a package fails, record the exact existing failure before touching Runtime code; do not classify it as a migration regression.

- [ ] **Step 2: 固定当前生产注册入口清单**

Record the current endpoints from `Sub2APIRuntimeRegistry` and `NewSub2APIProductionApplicationGateway`: Messages, Chat Completions, Responses, Count Tokens, Embeddings, Alpha Search, Gemini Native, Images, Videos and Live. The list becomes the expected registry in later architecture tests.

- [ ] **Step 3: Commit the baseline note**

```powershell
git add docs/superpowers/plans/2026-08-20-runtimebridge-sub2api-separation.md
git commit -m "docs(runtime): 拆分 RuntimeBridge 实施任务"
```

## 任务 1：建立公共 RuntimeBridge v1 contract

> 执行状态：已完成。`pkg/runtimebridge/v1`、contract 测试和脱敏/终态校验已通过。

**Files:**
- Create: `backend/pkg/runtimebridge/v1/types.go`
- Create: `backend/pkg/runtimebridge/v1/events.go`
- Create: `backend/pkg/runtimebridge/v1/errors.go`
- Create: `backend/pkg/runtimebridge/v1/contract_test.go`
- Modify: `backend/go.mod` only if package layout requires no dependency change (do not add a dependency)

- [ ] **Step 1: 写 contract 失败测试**

Add tests that require:

```go
func TestRequestRejectsMissingIdentity(t *testing.T)
func TestRequestClonesMutableFields(t *testing.T)
func TestTerminalEventAllowsExactlyOneFinalEvent(t *testing.T)
func TestRuntimeErrorNeverSerializesCredentialFields(t *testing.T)
```

The tests must assert that a request has `ContractVersion`, `RequestID`, positive platform ID, endpoint, and copied payload/header maps; a terminal collector accepts the first terminal event and rejects the second with a stable error.

- [ ] **Step 2: 运行测试确认失败**

```powershell
go test ./pkg/runtimebridge/v1 -count=1
```

Expected: FAIL because the package and types do not exist yet.

- [ ] **Step 3: 实现最小 contract**

Define the following public types without importing `internal/*`, Gin, Ent or Sub2API:

```go
type Endpoint string
type EventKind string
const (
    EventResponseStarted EventKind = "response_started"
    EventResponseChunk EventKind = "response_chunk"
    EventResponseFinished EventKind = "response_finished"
    EventRuntimeFailed EventKind = "runtime_failed"
    EventUsageFinal EventKind = "usage_final"
    EventStreamCancelled EventKind = "stream_cancelled"
)
type PlatformRoute struct {
    ID int64
    Code string
    RuntimeAdapter string
    RequestedModel string
    UpstreamModel string
    EndpointCapabilities []string
}
type OwnerRef struct { UserID int64; APIKeyID int64 }
type SessionMetadata struct {
    SessionID string
    UserAgent string
    ClientIP string
    RequestPayloadHash string
}
type Request struct {
    ContractVersion string
    RequestID       string
    Platform        PlatformRoute
    Endpoint        Endpoint
    Stream          bool
    Payload         []byte
    Headers         map[string]string
    Owner           OwnerRef
    Session         SessionMetadata
}
type Result struct {
    StatusCode       int
    ResponseHeaders  map[string][]string
    Body             []byte
    Streamed         bool
    AccountID        int64
    UpstreamEndpoint string
    UpstreamModel    string
    Usage            UsageFacts
}
type UsageFacts struct {
    AccountID int64
    PlatformID int64
    Endpoint string
    RequestedModel string
    UpstreamModel string
    InputTokens int
    OutputTokens int
    CacheCreationTokens int
    CacheReadTokens int
    FirstTokenMilliseconds int64
    DurationMilliseconds int64
    TerminalStatus string
    RequestWasClientStream bool
    ResponseWasPartiallySent bool
}
type RuntimeError struct {
    Category string
    Message string
    Retryable bool
    AttemptedAccountIDs []int64
    UpstreamStatus int
}
type Event struct {
    Sequence uint64
    Kind EventKind
    Result *Result
    Usage *UsageFacts
    Error *RuntimeError
}
```

Implement `Validate`, `Clone`, `TerminalEvent`, and `RecordTerminal` so unknown contract versions fail closed and mutable byte/header fields are copied.

- [ ] **Step 4: 运行测试确认通过**

```powershell
gofmt -w pkg/runtimebridge/v1
go test ./pkg/runtimebridge/v1 -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit**

```powershell
git add backend/pkg/runtimebridge/v1
git commit -m "feat(runtime): 建立 RuntimeBridge v1 契约"
```

## 任务 2：建立内部 Bridge Port 与本地 runner

> 执行状态：已完成。`internal/runtimebridge` LocalRuntime、Exchange context 和用量终态回归已通过。

**Files:**
- Create: `backend/internal/runtimebridge/port.go`
- Create: `backend/internal/runtimebridge/local.go`
- Create: `backend/internal/runtimebridge/terminal.go`
- Create: `backend/internal/runtimebridge/local_test.go`
- Modify: `backend/internal/applicationgateway/gateway.go`
- Modify: `backend/internal/gatewayruntime/runtime.go`

- [ ] **Step 1: 写本地 Bridge 行为测试**

Add tests for a stub Driver:

```go
func TestLocalBridgeCopiesProductRouteWithoutBillingAsset(t *testing.T)
func TestLocalBridgePublishesOneTerminalUsageEvent(t *testing.T)
func TestLocalBridgeReturnsMissingTerminalAsError(t *testing.T)
func TestLocalBridgePropagatesDriverErrorAfterTerminalEvent(t *testing.T)
```

The first test must verify that `BillingAsset`, rate multiplier, subscription ID and balance are not present in the Runtime request. The terminal tests must assert that the Product UsageSink receives exactly one event.

- [ ] **Step 2: 定义 Port**

Create the local interface:

```go
type Driver interface {
    Dispatch(context.Context, v1.Request, EventSink) (v1.Result, error)
}
type EventSink interface {
    Publish(context.Context, v1.Event) error
}
type RuntimePort interface {
    Dispatch(context.Context, productcore.Decision, v1.Request, gatewayruntime.UsageSink) (v1.Result, error)
}
```

The implementation must clone the product route into `v1.Platform`, never pass a product entity to the Driver, and convert only the final `usage_final` event into `gatewayruntime.UsageEvent`.

- [ ] **Step 3: 运行定向测试**

```powershell
go test -tags=unit ./internal/runtimebridge ./internal/applicationgateway -run 'LocalBridge|Terminal|RuntimePort' -count=1
```

Expected: PASS.

- [ ] **Step 4: Commit**

```powershell
git add backend/internal/runtimebridge backend/internal/applicationgateway/gateway.go backend/internal/gatewayruntime/runtime.go
git commit -m "feat(runtime): 增加本地 RuntimeBridge 端口"
```

## 任务 3：把 Sub2API registry/adapter 移到 Driver 包

> 执行状态：已完成。`internal/runtime/sub2api` registry/adapter 已建立，旧 Handler 包装仅作迁移桥。

**Files:**
- Create: `backend/internal/runtime/sub2api/registry.go`
- Create: `backend/internal/runtime/sub2api/adapter.go`
- Create: `backend/internal/runtime/sub2api/adapter_test.go`
- Modify: `backend/internal/handler/sub2api_runtime_registry.go`
- Modify: `backend/internal/handler/sub2api_runtime_adapter.go`
- Modify: `backend/internal/handler/sub2api_runtime_registry_test.go`

- [ ] **Step 1: 复制 registry 行为测试到新包**

The new package must retain tests for duplicate endpoint, unknown endpoint, nil executor and empty adapter. The expected endpoint set must match Task 0 exactly.

- [ ] **Step 2: 实现 Driver adapter**

Implement:

```go
type EndpointExecutor interface {
    Execute(context.Context, v1.Request, EventSink) (v1.Result, error)
}
type Adapter struct { executors map[v1.Endpoint]EndpointExecutor }
func NewAdapter(registry Registry) (*Adapter, error)
func (a *Adapter) Dispatch(context.Context, v1.Request, EventSink) (v1.Result, error)
```

The adapter must own terminal recording, reject a missing event, and return a structured unavailable error. It must not import `handler` or `gin`.

- [ ] **Step 3: 保留迁移期兼容包装**

Change the old Handler adapter to a thin type alias or forwarding wrapper used only by tests and unfinished endpoints. Do not add a second production fallback. The wrapper must have a comment identifying the endpoint migration it protects.

- [ ] **Step 4: 运行测试与架构守卫**

```powershell
go test -tags=unit ./internal/runtime/sub2api ./internal/handler ./internal/architecture -run 'Sub2API.*Runtime|Runtime.*Registry|AdapterPurity' -count=1 -timeout=30m
```

Expected: PASS, and no new legacy call site outside the single explicitly recorded compatibility file.

- [ ] **Step 5: Commit**

```powershell
git add backend/internal/runtime/sub2api backend/internal/handler/sub2api_runtime_registry.go backend/internal/handler/sub2api_runtime_adapter.go
git commit -m "refactor(runtime): 将 Sub2API registry 收口为 Driver"
```

## 任务 4：迁移生产组合到 Bridge Client

> 执行状态：已完成。生产组合已通过 LocalRuntime/RuntimeBridge，Task 4 定向 Handler/Service 回归通过。

**Files:**
- Modify: `backend/internal/handler/sub2api_runtime_composition.go`
- Modify: `backend/internal/handler/runtime_ingress.go`
- Modify: `backend/internal/handler/sub2api_execution.go`
- Modify: `backend/internal/applicationgateway/gateway.go`
- Create: `backend/internal/runtimebridge/composition_test.go`

- [ ] **Step 1: 写组合边界测试**

Assert that the production composition receives a `runtimebridge.Driver`/local runner and that `ApplicationGateway.Dispatch` passes only a cloned `v1.Request` to it. Assert that the composition does not require `GatewayHandler` or `OpenAIGatewayHandler` as a runtime dependency.

- [ ] **Step 2: 改造 composition**

Replace the current constructor shape:

```go
func NewSub2APIProductionApplicationGateway(
    decisions applicationgateway.DecisionProvider,
    driver runtimebridge.Driver,
    usage applicationgateway.UsageSinkFactory,
) *applicationgateway.Gateway
```

The constructor passes `driver` as the `gatewayruntime.GatewayRuntime` implementation through a local Bridge runner and passes the already-composed `UsageSinkFactory` unchanged. The HTTP handlers remain responsible for route registration and ingress error envelopes. They no longer become dependencies of the Runtime adapter. Service-specific construction of the `UsageSinkFactory` remains in the application composition root; it is not part of the Driver contract.

- [ ] **Step 3: 运行 OpenAI 生产路径回归**

```powershell
go test -tags=unit ./internal/handler ./internal/service ./internal/runtimebridge ./internal/runtime/sub2api -run 'OpenAI|ChatCompletions|Responses|Messages|Usage|Failover|RuntimeBoundary' -count=1 -timeout=40m
```

Expected: PASS with unchanged endpoint status, usage facts, account ID and terminal behavior.

- [ ] **Step 4: Commit**

```powershell
git add backend/internal/handler/sub2api_runtime_composition.go backend/internal/handler/runtime_ingress.go backend/internal/handler/sub2api_execution.go backend/internal/applicationgateway/gateway.go backend/internal/runtimebridge/composition_test.go
git commit -m "refactor(runtime): 让生产组合通过 RuntimeBridge"
```

## 任务 5：按端点家族移除 Gin/旧 Handler 依赖

> 执行状态：部分完成（本轮范围已完成）。OpenAI Chat/Responses/Messages、Images、Embeddings、Alpha Search、Count Tokens 已接入纯 Driver；Gemini Native、Grok 媒体、Videos、Live、WebSocket 仍保留显式延期集合。

**Files:**
- Modify: `backend/internal/runtime/sub2api/` endpoint executors
- Modify: `backend/internal/service/*_runtime.go` for explicit context/exchange entry points
- Modify: `backend/internal/handler/sub2api_legacy_dispatch.go` only to retain the last unfinished endpoint set
- Test: `backend/internal/architecture/sub2api_adapter_purity_guard_test.go`

- [x] **Step 1: 先迁移已纯化的 OpenAI 家族**

Move Chat, Responses, Messages, Images and Count Tokens executor implementations behind `v1.Request`/`v1.EventSink`. The executor entry point must remain:

```go
Execute(context.Context, v1.Request, runtimebridge.EventSink) (v1.Result, error)
```

It must not call `legacyEndpointExecutor`, `dispatchLegacyEndpoint`, `ginContextCarrier`, `GinContext()` or a public Handler.

- [x] **Step 2: 迁移 OpenAI 同步与媒体端点**

Move OpenAI Embeddings, Alpha Search, Count Tokens and Images using the same entry point. Preserve endpoint-specific JSON/SSE envelope conversion and usage facts. Gemini Native, Grok media and Videos remain outside this scope and stay in the deferred set.

- [x] **Step 3: 明确延期端点**

Live, WebSocket, Passthrough and any remaining compatibility path stay registered through the existing Driver until they have a pure exchange implementation. They must be listed in a test-visible migration set; they must not silently fall back from a new executor to an old Handler.

- [x] **Step 4: 收紧架构守卫**

Update `sub2api_adapter_purity_guard_test.go` so the allowed legacy bridge set decreases only when the corresponding endpoint test is green. Add a guard that `backend/internal/runtime/sub2api` contains no `gin-gonic` import and no `handler` import.

- [x] **Step 5: 运行端点分组回归**

```powershell
go test -tags=unit ./internal/handler ./internal/service ./internal/runtime/sub2api ./internal/architecture -run 'OpenAI|Images|Messages|Responses|Embeddings|Gemini|AlphaSearch|Video|RuntimeBoundary|Sub2APIAdapter|Purity' -count=1 -timeout=60m
```

Expected: PASS; all unchanged endpoints keep their existing response and billing tests.

- [ ] **Step 6: Commit**

```powershell
git add backend/internal/runtime/sub2api backend/internal/service backend/internal/handler/sub2api_legacy_dispatch.go backend/internal/architecture/sub2api_adapter_purity_guard_test.go
git commit -m "refactor(runtime): 清除 Sub2API Driver 的 Handler 依赖"
```

## 任务 6：固定账号事实与 Product UsageSink 终态

> 执行状态：已完成实现与验证。LocalRuntime、Product UsageSink 终态、实际账号、失败切换、延迟和缓存 Token conformance 已通过；提交仍按当前授权边界暂缓。

**Files:**
- Modify: `backend/internal/runtimebridge/terminal.go`
- Modify: `backend/internal/service/sub2api_product_usage_finalizer.go`
- Modify: `backend/internal/handler/runtime_usage_bridge.go`
- Create: `backend/internal/runtimebridge/usage_conformance_test.go`
- Create: `backend/internal/architecture/runtime_contract_guard_test.go`

- [x] **Step 1: 写终态和计费回归测试**

Add cases for:

```go
func TestFailedAttemptDoesNotReachProductUsageSink(t *testing.T)
func TestSuccessfulFailoverUsesFinalAccountID(t *testing.T)
func TestTerminalUsageIsExactlyOnce(t *testing.T)
func TestUsagePreservesLatencyAndCacheTokens(t *testing.T)
func TestUsageSinkFailureIsReturned(t *testing.T)
```

- [x] **Step 2: 收口 terminal event 转换**

Convert only the final contract event to `gatewayruntime.UsageEvent`. Preserve `AccountID`, `UpstreamEndpoint`, `UpstreamModel`, token fields, `DurationMilliseconds`, `FirstTokenMilliseconds`, `TerminalStatus`, stream state and partial-response state. Do not reconstruct billing assets inside Runtime.

- [x] **Step 3: 保持现有扣费实现但隔离入口**

Keep `Sub2APIProductUsageFinalizer` as the current Product UsageSink implementation. Its public input must be `DecisionSnapshot + UsageEvent`; the Driver must never call `RecordUsage` directly. Any missing account ID or successful event without usage facts must fail closed.

- [x] **Step 4: 运行服务回归**

```powershell
go test -tags=unit ./internal/runtimebridge ./internal/service ./internal/handler -run 'Usage|Billing|Failover|Latency|Cache|Terminal' -count=1 -timeout=40m
```

Expected: PASS, with no duplicate usage record and no charge for failed attempts.

- [ ] **Step 5: Commit**

```powershell
git add backend/internal/runtimebridge backend/internal/service/sub2api_product_usage_finalizer.go backend/internal/handler/runtime_usage_bridge.go backend/internal/architecture/runtime_contract_guard_test.go
git commit -m "fix(usage): 固定 Runtime 终态与实际账号归属"
```

## 任务 7：完成 contract conformance 与外置门禁

> 执行状态：已完成实现与验证。跨实现 contract conformance、纯度守卫和外置条件门禁已建立并通过；外置 RuntimeBridge 本身仍未启用。

**Files:**
- Create: `backend/pkg/runtimebridge/v1/conformance_test.go`
- Create: `backend/internal/runtimebridge/conformance_test.go`
- Modify: `backend/internal/architecture/sub2api_adapter_purity_guard_test.go`
- Modify: `backend/internal/architecture/legacy_configuration_guard_test.go`
- Modify: `backend/internal/architecture/my2_release_gate_test.go`
- Create: `docs/memory/决策/2026-08-20-RuntimeBridge模块化单体优先.md`
- Modify: `docs/memory/当前状态.md`

- [x] **Step 1: 增加跨实现 contract conformance tests**

The test suite must run the same request matrix against the local Sub2API Driver and a deterministic fake Driver. It must compare endpoint, response status, stream event order, terminal sequence, account facts and usage facts; it must not compare internal service types.

- [x] **Step 2: 增加外置条件门禁**

The architecture test must fail if an external transport is enabled before all of these are true: no production Gin carrier use, contract package has no forbidden import, RuntimeStore/ControlPort exists, terminal conformance passes, and service authentication/version checks are present.

- [x] **Step 3: 记录长期决策与当前状态**

Record that XCode uses Sub2API as the first Driver, prefers modular monolith deployment, and only later permits an external Bridge. Update `docs/memory/当前状态.md` to mark design complete and implementation Task 1 as the next stage. Do not record speculative CLIProxyAPI integration.

- [x] **Step 4: 运行架构测试**

```powershell
go test -tags=unit ./internal/architecture ./internal/runtimebridge ./pkg/runtimebridge/v1 -run 'Contract|Architecture|Purity|ReleaseGate' -count=1 -timeout=30m
```

Expected: PASS.

- [ ] **Step 5: Commit**

```powershell
git add backend/pkg/runtimebridge/v1 backend/internal/runtimebridge backend/internal/architecture docs/memory
git commit -m "test(architecture): 固化 RuntimeBridge 可替换门禁"
```

## 任务 8：完整验证与发布准备

> 执行状态：本地门禁已完成，发布外部动作待授权。Docker/离线包和真实服务器验收未在本轮执行，因此不得将本轮标记为可发布版本。

**Files:**
- Test only: backend and frontend existing suites
- Modify: `docs/memory/当前状态.md` with verified results

- [x] **Step 1: 后端单元与集成测试**

```powershell
Set-Location backend
$env:GOFLAGS='-p=1'
$env:GOMAXPROCS='2'
make test-unit
make test-integration
go build ./cmd/server
```

Expected: all commands exit with code 0.

- [x] **Step 2: 前端门禁**

```powershell
pnpm --dir frontend run test:run
pnpm --dir frontend run typecheck
pnpm --dir frontend run lint:check
pnpm --dir frontend run build
```

Expected: all commands exit with code 0; existing non-fatal test network warnings are recorded but not treated as failures.

- [ ] **Step 3: Docker 与离线包验证**

Run the repository's existing XCode release workflow locally or in GitHub Actions. Verify the produced image is `xcode:latest`, the offline archive can be loaded with `docker load`, and no database/Redis volume is included in the image.

- [ ] **Step 4: 真实链路验收**

Against the existing test server, run two rounds each for GPT Chat, GPT Responses, GLM Chat, one account failover, subscription-first billing, balance fallback and usage-record latency/cache fields. Record HTTP status, actual account ID, endpoint, token fields and billing result; never put API keys in logs or documents.

- [x] **Step 5: Final review and release decision**

Run `git diff --check`, inspect the final dependency graph, and verify the user-owned `frontend/pnpm-lock.yaml` and `my2.0.drawio` remain outside the implementation staging set. A release commit/tag still requires the separate Docker and real-server gates above; do not tag or deploy a partial RuntimeBridge migration.
