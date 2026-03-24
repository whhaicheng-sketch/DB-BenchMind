# AutoBench Progress

当前模块: M3
当前任务: T3.3
下一任务: T3.4
已完成任务: T0.1, T0.2, T1.1, T1.2, T1.3, T2.1, T2.2, T2.3, T3.1, T3.2, T3.3

## 本轮摘要
- 读取 `progress` 后确认本轮唯一任务为 `T3.3`，因此只实现 AutoBench 的本地执行计划预览。
- `autobenchWizardDraft` helper 新增 `buildLocalPlanPreview`，基于本地已选 connection 和 profile 生成稳定顺序的纯前端预览列表。
- `AutoBenchTab.vue` 新增 `Plan Preview` 区块，只做只读展示，不创建 suite、不调用执行能力、不引入真实执行语义。

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
- AutoBench 当前仍然只完成本地占位 connection/profile 选择和本地 plan preview，没有真实 connection/template 数据源、suite 创建、执行监控、报告生成或历史集成。
- Plan Preview 只是本地草图，后续进入 `T3.4` 时应继续把测试补强限制在前端预览层，不要提前接入执行链路。
- 仓库现有 `go test ./internal/domain/...` 会被 `internal/domain/connection/ssh_tunnel.go` 的既有问题阻断，不属于本轮改动范围。
