# Templates Compact Plan v0.1

> 文档路径：`docs/templates_compact/Templates_compact_plan_v0.1.md`

## 总体说明

本计划用于执行 Templates 模块“分阶段、可中断恢复”的正式整改。执行要求：

- 先维护文档，再动代码。
- 一次只做一个 task。
- 若中断，后续仅根据 `docs/Templates_compact_tasks_v0.1.md` 继续下一个未完成任务。

## 阶段 1：审计与文档基线

- 目标：
  - 审计 Templates 前后端现状。
  - 建立 compact spec / plan / tasks 三份基线文档。
  - 确认 phase 与字段删除的代码证据。
- 涉及文件：
  - `docs/Templates_compact_spec_v0.1.md`
  - `docs/Templates_compact_plan_v0.1.md`
  - `docs/Templates_compact_tasks_v0.1.md`
  - 以及所有审计源文件
- 改动策略：
  - 只写文档，不扩散到后续代码任务。
  - 若发现旧文档与现状冲突，以代码现状为准增量修正。
- 测试策略：
  - 记录当前 Templates 相关测试/构建基线。
  - 不把仓库已有无关失败误判为本阶段引入。
- 风险控制：
  - 不回滚当前工作树已有改动。
  - 所有审计结论必须有文件级证据。
- 完成标志：
  - 三份 compact 文档已存在。
  - T1 状态已完成。

## 阶段 2：UI 紧凑化与说明文案清理

- 目标：
  - Templates 弹窗、列表、筛选条视觉密度对齐 Connections。
  - 删除冗余解释文案，保留最短必要提示。
- 涉及文件：
  - `frontend/src/components/tabs/TemplatesTab.vue`
  - `frontend/src/components/template/TemplateEditorDialog.vue`
  - `frontend/src/components/template/TemplateBasicSection.vue`
  - `frontend/src/components/template/TemplateRuntimeSection.vue`
  - `frontend/src/components/template/TemplatePreviewSection.vue`
  - `frontend/src/components/template/TemplateToolbar.vue`
  - 相关样式与测试文件
- 改动策略：
  - 先写或更新前端测试，再做最小样式与文案收敛。
  - 不混入字段删除和后端链路删除之外的重构。
- 测试策略：
  - 运行 Templates 相关前端单测。
  - 必要时补充静态源码断言测试。
- 风险控制：
  - 保持浅色主题。
  - 不改变核心交互顺序。
- 完成标志：
  - 弹窗、列表、筛选条为紧凑浅色风格。
  - 冗余说明文案已清理。

### 阶段 2 当前状态

- 状态：`done`
- 完成时间：`2026-03-23`
- 本轮结果：
  - Templates 编辑弹窗收紧了 overlay、header、body、footer、按钮与 badge 密度。
  - Basic / Runtime / Preview section 的表单与摘要卡片间距已压缩。
  - 筛选条、列表头、列表项操作区已向 Connections 的密度靠拢。
  - 空状态文案与 Create Task 成功提示已改为更短的操作导向文案。
- 验证：
  - `node --test frontend/tests/templatesTab.test.mjs`
  - `cd frontend && npm run build`

## 阶段 3：前端删除 tags / scope / status / last updated 展示与筛选

- 目标：
  - 前端界面、筛选、badge、metadata 不再出现四组字段。
- 涉及文件：
  - `frontend/src/components/template/TemplateBasicSection.vue`
  - `frontend/src/components/template/TemplateFilterBar.vue`
  - `frontend/src/components/template/TemplateListItem.vue`
  - `frontend/src/components/template/TemplateList.vue`
  - `frontend/src/components/template/TemplatePreviewSection.vue`
  - `frontend/src/stores/template.js`
  - `frontend/src/mock/templates.js`
  - `frontend/tests/templatesTab.test.mjs`
- 改动策略：
  - 先补/改前端测试验证字段缺失。
  - 再删除展示和筛选残留。
- 测试策略：
  - 运行 `frontend/tests/templatesTab.test.mjs`。
- 风险控制：
  - 保持列表统计、空状态、创建流程不回退。
- 完成标志：
  - 前端界面和 store 不再依赖四组字段。

### 阶段 3 当前状态

- 状态：`done`
- 完成时间：`2026-03-23`
- 本轮结果：
  - 复核 Templates 列表、卡片/行项、编辑弹窗、Preview、Toolbar、FilterBar、store、mock、前端模型相关链路，确认前端界面已不再展示 `tags` / `scope` / `status` / `last updated`。
  - 补充前端测试，锁定四组字段相关展示/筛选缺失以及 store 仅接受 `search` / `dbFamily` / `tool` 三个筛选键。
  - 以最小侵入方式收紧 `frontend/src/stores/template.js` 的筛选入口，防止旧筛选键重新注入前端状态。
- 验证：
  - `node --test frontend/tests/templatesTab.test.mjs`
  - `cd frontend && npm run build`

## 阶段 4：删除四组字段的前后端数据链路

- 目标：
  - DTO、entity、repository、schema 注释、seed/mock、binding 统一清理四组字段残留。
- 涉及文件：
  - `frontend/src/models/template.js`
  - `frontend/wailsjs/go/models.ts.body`
  - `internal/domain/template/*`
  - `internal/app/usecase/*template*`
  - `internal/infra/database/repository/template_repo.go`
  - `internal/infra/database/schema.sql`
  - `internal/transportwails/bindings/template.go`
  - 相关测试文件
- 改动策略：
  - 先补后端测试和绑定测试，再做实现修正。
  - 将“物理审计列”与“产品字段”分开处理。
- 测试策略：
  - Go 单测覆盖 domain/usecase/repository/binding。
- 风险控制：
  - 不误删执行链仍在使用的通用字段。
  - 不破坏 builtin seed 加载与 CRUD。
- 完成标志：
  - 前后端 canonical model 不再暴露四组字段。

### 阶段 4 当前状态

- 状态：`done`
- 完成时间：`2026-03-23`
- 本轮结果：
  - 完成 Templates 前后端链路审计，确认前端 canonical model、Wails Templates DTO、后端 domain/usecase/repository 已不再把 `tags` / `scope` / `status` / `last updated` 作为业务字段流转。
  - 修正 `internal/domain/template/template.go` 的 canonical/legacy 判定，恢复 legacy 模板 JSON 与 `Clone()` 的兼容路径，保证旧 payload 带四组字段时不会导致主流程报错。
  - 将 `internal/app/usecase/template_usecase.go` 中仍带 `scope` 业务语义的方法调用改为只读判定命名，移除 Templates 业务链路中的 `scope` 命名残留。
  - 审计后确认 `frontend/wailsjs/go/models.ts.body` 的 Templates 类已不暴露四组字段，因此本轮无需额外改动该文件。
- 验证：
  - `node --test frontend/tests/templatesTab.test.mjs`
  - `go test ./internal/domain/template ./internal/app/usecase ./internal/transportwails/bindings ./internal/infra/database/repository -run 'Template|Seed'`
  - `cd frontend && npm run build`
  - `go build ./...`

## 阶段 5：phase 审计与收敛

- 目标：
  - 确认仅保留 `prepare` / `warmup` / `run` / `cleanup` 作为产品 phase。
- 涉及文件：
  - `frontend/src/models/template.js`
  - `frontend/src/constants/templateCapabilities.js`
  - `internal/domain/template/template.go`
  - `frontend/wailsjs/go/models.ts.body`
  - seed/mock/测试文件
  - Create Task from Template 相关链路涉及的模板使用点
- 改动策略：
  - 先补测试证明 canonical phase 集合与兼容映射行为。
  - 再清理 phase 残留或修正文档。
- 测试策略：
  - 前端模型测试。
  - Go domain/binding/usecase 测试。
- 风险控制：
  - 只删除“产品 phase”暴露，不删除必要兼容逻辑。
  - 不能凭文本搜索结果直接下结论。
- 完成标志：
  - phase 相关实现与文档一致。

### 阶段 5 当前状态

- 状态：`done`
- 开始时间：`2026-03-23`
- 完成时间：`2026-03-23`
- 本轮结果：
  - 完成八个 phase 在 Templates UI、前端 model/store/mock、Wails DTO、后端 template domain/DTO/repository、Create Task handoff、benchmark 执行链与 adapter 中的文件级审计。
  - 复核确认 canonical phase 继续只保留 `prepare` / `warmup` / `run` / `cleanup`；`build` / `generate` / `verify` / `delete` 仅存在于兼容层、工具内部步骤、日志/动作语义，不属于 Templates phase。
  - 补充 phase 测试，锁定前端仅展示四阶段、Create Task handoff 不携带 legacy phase，以及旧模板 JSON 的 legacy phase 配置可被兼容读取但 canonical 保存不再写回。
  - 修正 `internal/domain/template/template.go` 的 legacy phase 吸收逻辑，使旧 `build` / `generate` / `verify` / `delete` 的 `enabled` / `required` / `params` 能并入 canonical `prepare` / `cleanup`。
- 验证：
  - `node --test frontend/tests/templatesTab.test.mjs`
  - `go test ./internal/domain/template ./internal/app/usecase ./internal/transportwails/bindings ./internal/infra/database/repository -run 'Template|Seed'`
  - `cd frontend && npm run build`
  - `go build ./...`

## 阶段 6：回归验证与文档收尾

- 目标：
  - 做一轮针对 Templates 模块的前后端回归验证。
  - 更新 spec/plan/tasks 的完成态与实际偏差说明。
- 涉及文件：
  - 三份 compact 文档
  - 受影响测试文件
- 改动策略：
  - 只修复本轮整改引入的问题。
  - 不在收尾阶段扩需求。
- 测试策略：
  - Templates 相关前端测试。
  - Templates 相关 Go 测试。
  - `npm run build`
  - `go build` 或等价目标构建。
- 风险控制：
  - 明确记录仍然存在但不属于本轮引入的问题。
- 完成标志：
  - 验证结果已记录。
  - tasks 全部收口为 `done` 或明确 `blocked`。

### 阶段 6 当前状态

- 状态：`done`
- 开始时间：`2026-03-23`
- 完成时间：`2026-03-23`
- 本轮结果：
  - 完成 T1-T5 目标的整体验证，覆盖 Templates UI 紧凑化、文案收敛、字段裁剪、phase 收敛、默认模板收口、legacy payload/phase 兼容与 Create Task handoff 相关链路。
  - 复核确认 fresh install、existing data、默认模板去重、默认模板稳定更新而非危险重建等场景均有测试支撑。
  - 本轮未发现需要额外修复的 Templates 直接相关阻塞，因此未引入 T6 回归修正代码。
  - 已同步更新 spec / plan / tasks 三份文档至最终完成态，并标记本轮整改可收尾。
- 验证：
  - `cd frontend && node --test tests/templatesTab.test.mjs`
  - `go test ./internal/domain/template ./internal/app/usecase ./internal/transportwails/bindings ./internal/infra/database/repository ./internal/infra/database -run 'Template|Seed|SQLite|Builtin'`
  - `cd frontend && npm run build`
  - `go build ./...`

## 整体完成结论

- `Templates_compact` 的 T1-T6 已全部完成，当前任务集无 `doing`、无 `blocked`。
- 当前实现满足本轮既定范围，且未发现与本轮目标直接相关的残留问题。
- 本轮整改已进入可收尾状态，无需新增 `T7`。

## Hotfix 阶段：H1 phase 归一化回归修复

- 目标：
  - 以最小侵入方式修复验收期发现的默认模板弹窗崩溃。
  - 保持 T5 四阶段 canonical 收敛不回退。
- 涉及文件：
  - `frontend/src/models/template.js`
  - `frontend/tests/templatesTab.test.mjs`
  - `docs/templates_compact/Templates_compact_spec_v0.1.md`
  - `docs/templates_compact/Templates_compact_plan_v0.1.md`
  - `docs/templates_compact/Templates_compact_tasks_v0.1.md`
- 改动策略：
  - 先做根因审计，确认 `s.run.required` 的源码访问链路。
  - 先补前端失败测试锁定“缺失 phase 子对象时自动补齐”的行为。
  - 仅修补前端 phase 归一化层，不扩展到无关后端整改。
- 测试策略：
  - `cd frontend && node --test tests/templatesTab.test.mjs`
  - `cd frontend && npm run build`
- 风险控制：
  - 不推翻 T1-T6 已完成状态。
  - 不把 legacy phase 重新暴露为 UI phase。
  - 不做顺手重构。
- 完成标志：
  - 默认模板 `View` / `Edit` / `Copy` 不再因 phase 缺失而崩溃。
  - 前端接收 partial/legacy phase 数据时会补齐四阶段安全结构。

### Hotfix 当前状态

- 状态：`done`
- 开始时间：`2026-03-23`
- 完成时间：`2026-03-23`
- 本轮结果：
  - 根因定位到 `frontend/src/models/template.js` 的 `createPhaseState` / `normalizePhasesForTool` 链路；旧实现对缺失 `run` 的 partial phase 输入进行了错误的浅覆盖。
  - 已将 phase 归一化改为逐 phase 补齐默认结构，保证 `prepare` / `warmup` / `run` / `cleanup` 都始终存在，且 `run.required` 读取安全。
  - 已补前端回归测试，覆盖缺失 `run`、缺失 `warmup` 等 partial phase 输入场景。
- 验证：
  - `cd frontend && node --test tests/templatesTab.test.mjs`
  - `cd frontend && npm run build`
