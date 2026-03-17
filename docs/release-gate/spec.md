# DB-BenchMind 自动化回归门禁规范

> 文档类型：Spec
> 状态：强制执行
> 生效范围：所有功能开发、缺陷修复、构建调整、依赖升级、发布验收

## 1. 目标

DB-BenchMind 的自动化回归测试是发布门禁，不是建议流程。任何会影响功能、构建、配置、依赖、启动链路或发布产物的变更，都必须通过自动化门禁后才允许合并、发布和验收通过。

## 2. 强制规则

1. 未执行自动化门禁，不允许声称任务完成。
2. 自动化门禁任一环节失败，不允许合并，不允许发布。
3. 新功能没有自动化测试，不算完成。
4. bug 修复没有回归测试，不算完成。
5. 构建脚本、启动脚本、依赖版本、配置结构发生变更时，必须执行完整回归门禁。
6. 发布记录必须附上自动化门禁执行证据。

## 3. 当前强制门禁范围

### 3.1 基础构建门禁

- `./scripts/entry/build` 必须成功
- 必须产出 `./bin/db-benchmind`
- 前端生产构建必须成功
- Wails 构建必须成功

### 3.2 后端测试门禁

- `scripts/gates/test_backend_gate.sh` 定义的核心后端测试必须通过
- 覆盖范围至少包括：usecase、核心 domain、repository、adapter、Wails transport、collector

### 3.3 前端测试门禁

- `npm --prefix frontend run test` 必须通过
- 覆盖范围至少包括：核心导航、核心状态流、任务监控关键状态规则

### 3.4 启动 smoke test 门禁

- `scripts/gates/test_script_layout.sh` 必须通过
- `scripts/gates/test_start_cleanup.sh` 必须通过
- `scripts/gates/test_smoke_gate.sh` 必须通过
- smoke test 至少验证：
  - 程序能启动到 Wails startup
  - Benchmark / Monitor / Task binding 完成上下文注入
  - 系统监控启动
  - 不再出现 “Wails applications will not build without the correct build tags.”

### 3.5 统一门禁入口

- `./scripts/entry/regression`
- `make regression`
- `make release-gate`

以上任一入口均属于正式门禁入口，结果必须等价。

## 4. 强制触发条件

以下任一情况必须执行自动化回归门禁：

- 新增功能
- 修改已有逻辑
- 修复 bug
- 改 UI 交互
- 改 store / 状态流
- 改后端 usecase / domain / repository
- 改数据库适配器
- 改构建脚本
- 改依赖版本
- 改配置结构
- 改启动流程
- 准备发布
- 合并高风险分支

## 5. 通过标准

只有在以下条件全部满足时，才允许认定门禁通过：

1. `./scripts/entry/build` 成功
2. `scripts/gates/test_script_layout.sh` 成功
3. `scripts/gates/test_backend_gate.sh` 成功
4. `scripts/gates/test_frontend_gate.sh` 成功
5. `scripts/gates/test_start_cleanup.sh` 成功
6. `scripts/gates/test_smoke_gate.sh` 成功
6. 没有 blocker / critical 级问题
7. 文档与实现保持一致

## 6. 当前阶段边界

当前项目尚未具备“一条 `go test ./...` 全绿”条件，因此第一阶段门禁采用“稳定核心子集 + 构建 + smoke”模式。以下内容属于已知缺口，不允许忽略，但允许在后续阶段补齐：

- `internal/transportwails/bindings` 的历史测试失效
- `test/integration` 的旧集成用例与当前规则不一致
- `internal/domain/connection` 仍存在全量测试不可直接纳入门禁的问题

这些缺口必须在后续计划中逐步收敛，但不阻止当前阶段先建立强制门禁。
