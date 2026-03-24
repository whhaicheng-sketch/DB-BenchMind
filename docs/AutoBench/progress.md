# AutoBench Progress

当前模块: M3
当前任务: T3.2
下一任务: T3.3
已完成任务: T0.1, T0.2, T1.1, T1.2, T1.3, T2.1, T2.2, T2.3, T3.1, T3.2

## 本轮摘要
- 读取 `progress` 后确认本轮唯一任务为 `T3.2`，因此只增强 AutoBench Wizard 的 profile 选择区。
- `autobenchWizardDraft` helper 为 `test / cpu_bound / io_bound` 增加了本地 scope 元数据、用途说明与选择摘要函数，仍然完全不依赖后端。
- `AutoBenchTab.vue` 的 Profiles 区现在使用更明确的本地配置卡片和顺序摘要展示，但仍未进入执行计划预览或真实 suite 生成。

## 改动文件
- `frontend/src/components/tabs/AutoBenchTab.vue`
- `frontend/src/components/tabs/autobenchWizardDraft.mjs`
- `frontend/tests/autoBenchTab.test.mjs`
- `frontend/tests/autobenchWizardDraft.test.mjs`
- `frontend/tests/autobenchProgressDocs.test.mjs`
- `docs/AutoBench/progress.md`
- `docs/AutoBench/progress.json`

## 测试结果
- `node --test frontend/tests/autoBenchTab.test.mjs`: pass
- `node --test frontend/tests/autobenchWizardDraft.test.mjs`: pass
- `node --test frontend/tests/appNavigation.test.mjs frontend/tests/autobenchProgressDocs.test.mjs`: pass
- `cd frontend && node --test tests/*.test.mjs`: pass

## 已知风险
- AutoBench 当前仍然只完成本地占位 connection 和 profile 选择，没有真实 connection/template 数据源、suite 创建、执行监控、报告生成或历史集成。
- Profiles 区当前只提供本地 metadata 与顺序摘要，后续接 `T3.3` 执行计划预览时仍需避免引入真实执行语义。
- 仓库现有 `go test ./internal/domain/...` 会被 `internal/domain/connection/ssh_tunnel.go` 的既有问题阻断，不属于本轮改动范围。
