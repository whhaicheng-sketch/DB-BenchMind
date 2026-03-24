# Templates Compact Tasks v0.1

> 文档路径：`docs/templates_compact/Templates_compact_tasks_v0.1.md`

## 执行规则

- 顺序执行，不跳任务。
- 恢复时优先继续 `doing`；若无 `doing`，执行第一个不是 `done` 的任务。
- 恢复前先读取：
  - `docs/templates_compact/Templates_compact_spec_v0.1.md`
  - `docs/templates_compact/Templates_compact_plan_v0.1.md`
  - `docs/templates_compact/Templates_compact_tasks_v0.1.md`
- 若代码已完成但文档未更新，先修正文档状态。
- 若发生阻塞，必须记录阻塞原因和影响范围。

## 当前任务集状态

- 总状态：`completed`
- 可收尾：`yes`
- 最新完成任务：`H1`

## Task List

### T1. 审计与文档落地

- Task ID: `T1`
- 标题：审计 Templates 现状并建立 compact 文档基线
- 目标：
  - 完成前后端与关键词审计。
  - 新增或更新 compact spec / plan / tasks 文档。
  - 建立后续可中断恢复的执行基线。
- 前置条件：
  - 无
- 涉及文件：
  - `docs/templates_compact/Templates_compact_spec_v0.1.md`
  - `docs/templates_compact/Templates_compact_plan_v0.1.md`
  - `docs/templates_compact/Templates_compact_tasks_v0.1.md`
  - 所有被审计的 Templates 前后端文件
- 实现步骤：
  - [x] 读取现有 Templates 旧文档与当前代码。
  - [x] 全局搜索 `tags` / `scope` / `status` / `last updated` / `updated_at` / `build` / `prepare` / `generate` / `warmup` / `run` / `verify` / `cleanup` / `delete`。
  - [x] 记录前端、后端、binding、seed、Create Task from Template 的现状结论。
  - [x] 写入 compact spec。
  - [x] 写入 compact plan。
  - [x] 写入 compact tasks。
- 测试步骤：
  - [x] 记录前端 Templates 相关测试运行方式与结果。
  - [x] 记录 Go Templates 相关测试运行方式与结果。
- 验收标准：
  - 三份文档存在且内容与当前代码一致。
  - 后续执行者可以仅依赖本文件继续 T2。
- 状态：`done`
- 备注：
  - 2026-03-20 已完成。
  - 基线验证结果：
    - `cd frontend && node --test tests/templatesTab.test.mjs` 尚未执行于本 task；前次错误命令为 `npm test -- --test-name-pattern=template`，该脚本不支持附加参数。
    - `go test ./internal/domain/template ./internal/app/usecase ./internal/transportwails/bindings -run 'Template|Seed'` 失败，存在仓库既有问题：`internal/domain/template` 的旧 JSON/Clone 测试与当前 canonical 校验不一致，`internal/app/usecase` 与 `internal/transportwails/bindings` 还有 Templates 无关的编译失败基线。

### T2. UI 紧凑化 + 删除冗余解释文案

- Task ID: `T2`
- 标题：收紧 Templates 弹窗与列表密度，清理冗余说明文案
- 目标：
  - 对齐 Connections 的浅色紧凑节奏。
  - 移除不必要的说明性 copy。
- 前置条件：
  - `T1` 完成
- 涉及文件：
  - `frontend/src/components/tabs/TemplatesTab.vue`
  - `frontend/src/components/template/TemplateEditorDialog.vue`
  - `frontend/src/components/template/TemplateBasicSection.vue`
  - `frontend/src/components/template/TemplateRuntimeSection.vue`
  - `frontend/src/components/template/TemplatePreviewSection.vue`
  - `frontend/src/components/template/TemplateToolbar.vue`
  - `frontend/tests/templatesTab.test.mjs`
- 实现步骤：
  - [ ] 先补/改前端测试，锁定紧凑浅色风格与文案删除要求。
  - [ ] 统一弹窗、列表、筛选条的间距、标题、按钮密度。
  - [ ] 删除章节副标题和多余说明文案。
  - [ ] 保留最短必要的只读/验证提示。
- 测试步骤：
  - [ ] `cd frontend && node --test tests/templatesTab.test.mjs`
- 验收标准：
  - 弹窗和列表为浅色紧凑风格。
  - 关键冗余说明文案不再出现。
- 状态：`done`
- 备注：
  - 2026-03-23 已完成。
  - 已先补前端测试并验证失败，再完成最小侵入实现。
  - 本轮仅收紧 Templates 弹窗/列表/筛选/表单密度并清理冗余说明文案，未扩展到 T3/T4/T5。
  - 验证结果：
    - `node --test frontend/tests/templatesTab.test.mjs` 通过。
    - `cd frontend && npm run build` 通过。

### T3. 删除 tags / scope / status / last updated 的前端展示与筛选

- Task ID: `T3`
- 标题：移除四组字段的前端展示、badge、metadata 与筛选
- 目标：
  - 前端可见层彻底不再暴露四组字段。
- 前置条件：
  - `T2` 完成
- 涉及文件：
  - `frontend/src/components/template/TemplateBasicSection.vue`
  - `frontend/src/components/template/TemplateFilterBar.vue`
  - `frontend/src/components/template/TemplateListItem.vue`
  - `frontend/src/components/template/TemplateList.vue`
  - `frontend/src/components/template/TemplatePreviewSection.vue`
  - `frontend/src/stores/template.js`
  - `frontend/src/mock/templates.js`
  - `frontend/tests/templatesTab.test.mjs`
- 实现步骤：
  - [ ] 先写前端失败测试。
  - [ ] 删除表单中的四组字段。
  - [ ] 删除列表、badge、preview、metadata 中的四组字段。
  - [ ] 删除筛选条件、选项和相关统计文案。
- 测试步骤：
  - [ ] `cd frontend && node --test tests/templatesTab.test.mjs`
- 验收标准：
  - 前端界面和 store 不再依赖这四组字段。
- 状态：`done`
- 备注：
  - 2026-03-23 已完成。
  - 执行前复核发现 T2 已顺带覆盖大部分前端展示删除；本轮继续完成 T3 审计、测试补强与 store 筛选白名单收口。
  - 本轮仅处理前端展示/筛选层，不触碰 T4 的后端字段删除。
  - 验证结果：
    - `node --test frontend/tests/templatesTab.test.mjs` 通过。
    - `cd frontend && npm run build` 通过。

### T4. 删除这四组字段的前后端数据链路

- Task ID: `T4`
- 标题：清理四组字段在 DTO / entity / repository / schema / seed 中的残留链路
- 目标：
  - Templates canonical model 前后端一致去除四组字段。
- 前置条件：
  - `T3` 完成
- 涉及文件：
  - `frontend/src/models/template.js`
  - `frontend/wailsjs/go/models.ts.body`
  - `internal/domain/template/*`
  - `internal/app/usecase/*template*`
  - `internal/infra/database/repository/template_repo.go`
  - `internal/infra/database/schema.sql`
  - `internal/domain/template/seeds.go`
  - `internal/transportwails/bindings/template.go`
  - 相关测试文件
- 实现步骤：
  - [ ] 先补 Go/前端测试锁定字段缺失。
  - [ ] 清理 domain/usecase/repository/binding 中的四组字段。
  - [ ] 校正 seed/mock/schema 注释与兼容逻辑。
  - [ ] 更新 Wails/TS 模型。
- 测试步骤：
  - [ ] 运行 Templates 相关 Go 测试。
  - [ ] 运行 `cd frontend && node --test tests/templatesTab.test.mjs`
- 验收标准：
  - 前后端 canonical model 不再暴露四组字段。
- 状态：`done`
- 备注：
  - 2026-03-23 已完成。
  - 本轮先完成前后端链路审计，再补 domain 兼容测试并修正 legacy/canonical 判定。
  - 已去除 Templates 业务链路中仍带 `scope` 语义的方法命名残留；`updated_at` 仅保留为数据库底层审计列。
  - 审计确认 `frontend/wailsjs/go/models.ts.body` 的 Templates 相关类已不暴露四组字段，因此无需额外改动生成模型。
  - 为了让 T4 相关 Go 测试可执行，本轮同时修正了少量与 Templates 无关但阻塞验证的测试基线：
    - `internal/app/usecase/benchmark_usecase_test.go`
    - `internal/transportwails/bindings/connection_test.go`
    - `internal/transportwails/bindings/oracle_binding_mode_test.go`
  - 验证结果：
    - `node --test frontend/tests/templatesTab.test.mjs` 通过。
    - `go test ./internal/domain/template ./internal/app/usecase ./internal/transportwails/bindings ./internal/infra/database/repository -run 'Template|Seed'` 通过。
    - `cd frontend && npm run build` 通过。
    - `go build ./...` 通过。
  - `updated_at` 若仅保留为数据库审计列，不视为任务未完成。

### T5. phase 审计与收敛

- Task ID: `T5`
- 标题：以代码证据完成 phase 收敛
- 目标：
  - 明确并验证仅保留 `prepare` / `warmup` / `run` / `cleanup`。
- 前置条件：
  - `T4` 完成
- 涉及文件：
  - `frontend/src/models/template.js`
  - `frontend/src/constants/templateCapabilities.js`
  - `internal/domain/template/template.go`
  - `frontend/wailsjs/go/models.ts.body`
  - seed/mock/测试文件
- 实现步骤：
  - [ ] 先写 phase 相关失败测试。
  - [ ] 审核兼容映射与 canonical phase 暴露。
  - [ ] 删除错误的产品 phase 暴露或补齐文档结论。
  - [ ] 验证 Create Task from Template 主流程未受影响。
- 测试步骤：
  - [ ] 运行 Templates 相关前端与 Go 测试。
- 验收标准：
  - 代码与文档均明确四阶段 canonical model。
- 状态：`done`
- 备注：
  - `generate` 在 Swingbench `wizardOperation` 中出现不视为 phase 保留依据。
  - 2026-03-23 已完成。
  - 本轮先完成 phase 审计，再补测试，再以最小实现修正 legacy phase 兼容吸收逻辑；未扩展到 T6。
  - 审计确认：
    - `prepare` / `warmup` / `run` / `cleanup` 仍是 Templates canonical phase。
    - `build` / `generate` / `verify` / `delete` 不进入 Templates phase 展示或 DTO 链路；仅存在于旧 JSON 兼容、工具内部步骤、日志或删除动作语义。
  - 实现修正：
    - `internal/domain/template/template.go` 的 `PhaseSet.UnmarshalJSON` 现会把 legacy phase 的 `enabled` / `required` / `params` 吸收到 canonical `prepare` / `cleanup`，重保存时仍只输出四阶段。
  - 验证结果：
    - `node --test frontend/tests/templatesTab.test.mjs` 通过。
    - `go test ./internal/domain/template ./internal/app/usecase ./internal/transportwails/bindings ./internal/infra/database/repository -run 'Template|Seed'` 通过。
    - `cd frontend && npm run build` 通过。
    - `go build ./...` 通过。

### T6. 回归验证与文档收尾

- Task ID: `T6`
- 标题：回归验证并收尾 compact 文档
- 目标：
  - 完成本轮整改验证并更新文档最终状态。
- 前置条件：
  - `T5` 完成
- 涉及文件：
  - 三份 compact 文档
  - 相关测试文件
- 实现步骤：
  - [ ] 运行 Templates 相关前后端测试。
  - [ ] 运行 `cd frontend && npm run build`。
  - [ ] 运行 `go build` 或等价构建。
  - [ ] 把验证结果和遗留问题写回文档。
- 测试步骤：
  - [ ] `cd frontend && node --test tests/templatesTab.test.mjs`
  - [ ] `cd frontend && npm run build`
  - [ ] `go build ./...`
- 验收标准：
  - 文档状态、代码状态、验证结果一致。
- 状态：`done`
- 备注：
  - 2026-03-23 已完成。
  - 本轮按规则先从 `todo` 切到 `doing`，再执行完整回归；未发现与 T1-T5 目标直接相关、需要进入 T6 修复分支的阻塞。
  - 回归结论：
    - UI 与交互：创建/编辑弹窗维持浅色紧凑风格，解释性文案未回退；`Create` / `Edit` / `Save As` / `Create Task from Template` 对应链路验证通过。
    - 字段裁剪：前端展示、筛选、DTO/binding、domain/usecase/repository 不再依赖 `tags` / `scope` / `status` / `last updated`；legacy payload 可兼容读取。
    - phase 收敛：canonical phase 仍仅为 `prepare` / `warmup` / `run` / `cleanup`；legacy phase 可兼容吸收但不会重新写回。
    - 数据与模板：默认模板仅保留 `tpl_test_mysql_sysbench`；fresh install / existing data / 去重 / 稳定更新既有默认模板记录的场景均成立。
  - 验证结果：
    - `cd frontend && node --test tests/templatesTab.test.mjs` 通过。
    - `go test ./internal/domain/template ./internal/app/usecase ./internal/transportwails/bindings ./internal/infra/database/repository ./internal/infra/database -run 'Template|Seed|SQLite|Builtin'` 通过。
    - `cd frontend && npm run build` 通过。
    - `go build ./...` 通过。
  - 当前任务集 T1-T6 已全部完成，本轮整改可整体收尾，无需新增 `T7`。

### H1. 修复 template phase 读取/归一化导致的运行时崩溃

- Task ID: `H1`
- 标题：修复默认模板在 partial/legacy phase 数据下的前端运行时崩溃
- 目标：
  - 修复验收期发现的 `undefined is not an object (evaluating 's.run.required')`。
  - 保证默认模板 `View` / `Edit` / `Copy` 在 partial/legacy phase 数据下不崩溃。
- 前置条件：
  - `T6` 完成
- 涉及文件：
  - `frontend/src/models/template.js`
  - `frontend/tests/templatesTab.test.mjs`
  - `docs/templates_compact/Templates_compact_spec_v0.1.md`
  - `docs/templates_compact/Templates_compact_plan_v0.1.md`
  - `docs/templates_compact/Templates_compact_tasks_v0.1.md`
- 实现步骤：
  - [x] 审计 `s.run.required` 的源码访问链路与触发条件。
  - [x] 先补前端失败测试，复现 partial phase 输入触发的崩溃。
  - [x] 修补前端 phase 归一化，确保缺失 phase 子对象时自动补齐四阶段安全结构。
  - [x] 回写 spec / plan / tasks 的 hotfix 记录。
- 测试步骤：
  - [x] `cd frontend && node --test tests/templatesTab.test.mjs`
  - [x] `cd frontend && npm run build`
- 验收标准：
  - 默认模板 `View` 不再报错。
  - 默认模板 `Edit` / `Copy` 不再因 phase 结构缺失而报错。
  - 前端拿到 canonical / legacy / partial phase 数据时，都会先归一化再渲染。
- 状态流转：
  - `todo` -> `doing` -> `done`
- 状态：`done`
- 备注：
  - 2026-03-23 已完成。
  - 根因不在默认 seed 本身；后端默认 seed 仍是完整四阶段结构。真正问题是前端 `createPhaseState` 对 partial phase 输入使用了对象浅覆盖，导致 `run` 可能被覆盖成 `undefined`。
  - 本轮仅修复前端 phase 归一化层，未回退 T5 phase 收敛，也未扩展到无关优化。
  - 验证结果：
    - `cd frontend && node --test tests/templatesTab.test.mjs` 通过。
    - `cd frontend && npm run build` 通过。
