# AutoBench Progress

当前模块: M1
当前任务: T1.3
下一任务: T2.1
已完成任务: T0.1, T0.2, T1.1, T1.2, T1.3

## 本轮摘要
- 读取 `progress` 后确认本轮唯一任务为 `T1.3`。
- 补齐 AutoBench 领域模型与 use case 骨架测试，覆盖空连接校验、支持 profile 列表、快照隔离与 not-found 语义。
- 为匹配新增测试，仅补了一个最小校验：`CreateSuite` 在连接列表归一化后为空时返回明确错误。

## 改动文件
- `internal/app/usecase/autobench_usecase.go`
- `internal/app/usecase/autobench_usecase_test.go`
- `docs/AutoBench/progress.md`
- `docs/AutoBench/progress.json`

## 测试结果
- `go test ./internal/app/usecase -run AutoBenchSuiteUseCase`: pass
- `go test ./internal/domain/autobench`: pass
- `cd frontend && node --test tests/connectionFormRemote.test.mjs tests/connectionsTabAggregatedAiTest.test.mjs tests/connectionsTabBadges.test.mjs tests/templatesTab.test.mjs tests/tasksMonitorConnectionDefaults.test.mjs tests/tasksMonitorTemplateSelection.test.mjs tests/tasksMonitorPerformanceMetrics.test.mjs tests/tasksMonitorOracleSwingbench.test.mjs tests/tasksMonitorStatusStrip.test.mjs tests/tasksMonitorTaskState.test.mjs tests/taskBinding.test.mjs tests/appNavigation.test.mjs tests/appStore.test.mjs`: pass

## 已知风险
- AutoBench 当前仍只完成领域模型与 use case 骨架，尚未进入 `T2.*` 页面入口与布局实现。
- `CreateSuite` 当前采用内存级存储与占位状态语义，尚未接真实持久化或执行态，这是本阶段刻意限制。
- 仓库现有 `go test ./internal/domain/...` 会被 `internal/domain/connection/ssh_tunnel.go` 的既有问题阻断，不属于本轮改动范围。
