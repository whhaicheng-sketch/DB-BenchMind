# Templates Compact Spec v0.1

> 文档路径：`docs/templates_compact/Templates_compact_spec_v0.1.md`

## 背景与目标

本次整改面向现有 `Templates` 模块的正式收敛，不再继续扩张旧版“全量模板工作台”设想，而是把当前实现收敛为可维护、可中断恢复、可继续迭代的紧凑版本。

最终目标：

- Templates 创建/编辑弹窗改为更紧凑的浅色风格，密度与交互节奏接近 Connections。
- 删除冗余解释性文案，界面以操作为主。
- 删除 `tags` / `scope` / `status` / `last updated` 四组内容在表单、列表、筛选、badge、metadata 以及前后端数据链路中的存在。
- 基于代码证据审计 phase，仅保留真实可用且产品需要的 phase；目标态收敛为 `prepare` / `warmup` / `run` / `cleanup`。
- 不破坏 `Create` / `Edit` / `Save As` / `Create Task from Template` 主流程。

## 当前问题审计结论

### 前端审计结论

- 当前页面已局部朝紧凑化收敛，但仍保留旧版文案、旧文档和未完全收尾的交互表达。
- Templates 关键前端文件为：
  - `frontend/src/components/tabs/TemplatesTab.vue`
  - `frontend/src/components/template/TemplateEditorDialog.vue`
  - `frontend/src/components/template/TemplateBasicSection.vue`
  - `frontend/src/components/template/TemplateRuntimeSection.vue`
  - `frontend/src/components/template/TemplatePreviewSection.vue`
  - `frontend/src/components/template/TemplateToolbar.vue`
  - `frontend/src/components/template/TemplateFilterBar.vue`
  - `frontend/src/components/template/TemplateList.vue`
  - `frontend/src/components/template/TemplateListItem.vue`
  - `frontend/src/stores/template.js`
  - `frontend/src/models/template.js`
  - `frontend/src/mock/templates.js`
- 当前前端模型与 store 已基本去除 `tags` / `scope` / `status` / `updatedAt`，筛选器已缩减为 `search` / `dbFamily` / `tool`，说明前端主线已开始收敛。
- 当前弹窗、列表、筛选条已是浅色风格，但仍需以 Connections 的密度为标尺继续压缩留白、标题层级和冗余提示。

### 后端审计结论

- Templates 领域模型当前 canonical 字段已不包含 `scope` / `tags` / `status` / `updatedAt`，见：
  - `internal/domain/template/template.go`
  - `internal/transportwails/bindings/template.go`
- 持久化仍使用数据库表的 `updated_at` 作为基础审计字段，但该字段已不应再作为 Templates 产品展示字段；该列保留与否属于数据层实现决策，不属于前台产品字段。
- 内置 seed 当前仅保留一个默认测试模板 `tpl_test_mysql_sysbench`，见：
  - `internal/domain/template/seeds.go`
  - `frontend/src/mock/templates.js`
- Create Task from Template 当前为“任务壳”交接，不直接传输被删除的 metadata 字段，只传：
  - `templateId`
  - `templateName`
  - `tool`
  - `dbFamily`
  - `workloadFamily`
  - `source`
  见 `frontend/src/stores/template.js` 与 `frontend/src/stores/app.js`/`appState.mjs`。

### phase 审计结论

- 当前 Templates canonical model 只保留四个 phase：
  - `prepare`
  - `warmup`
  - `run`
  - `cleanup`
  证据：
  - `frontend/src/models/template.js`
  - `frontend/src/constants/templateCapabilities.js`
  - `internal/domain/template/template.go`
  - `frontend/wailsjs/go/models.ts.body`
- `build` / `generate` / `verify` / `delete` 目前不再作为 Templates canonical phase 出现在前端模型、能力矩阵、绑定 DTO 和领域结构中。
- `internal/domain/template/template.go` 的 `PhaseSet.UnmarshalJSON` 仍对旧字段做兼容吸收：
  - `build` / `generate` 被并入 `prepare`
  - `verify` / `delete` 被并入 `cleanup`
  - 旧 phase 的 `enabled` / `required` / `params` 会被吸收到 canonical phase，保证旧模板 JSON 读取后再保存时不会丢失同层 phase 配置
  这说明它们当前仅作为兼容输入，而非产品级 phase。
- 代码中仍出现 `generate` / `verify` / `delete` 文本，但结论如下：
  - `generate`：主要残留为 Swingbench `wizardOperation` 默认值，不是模板 phase。
  - `verify`：主要出现在 benchmark 执行链日志/步骤语义中，不是模板 phase。
  - `delete`：主要是模板删除动作，不是 benchmark 生命周期 phase。
  - `build`：主要出现在“build command”日志语义中，不是模板 phase。

## 本次整改范围

- Templates 页面、列表、筛选条、编辑弹窗、预览区的紧凑化与浅色风格统一。
- 删除冗余解释文案。
- 删除 `tags` / `scope` / `status` / `last updated` 的前端展示、筛选、badge、metadata。
- 删除上述四组字段的前后端数据链路残留，包括 store、mock、DTO、entity、repository、schema 注释/seed/binding 相关实现。
- 以代码证据完成 phase 审计，并将产品定义明确为四阶段模型。
- 回归 `Create` / `Edit` / `Save As` / `Create Task from Template` 主流程。

## 非范围项

- 不新增新的模板家族或新的 benchmark tool。
- 不改造 Tasks & Monitor 的整体交互。
- 不重写 benchmark 执行引擎。
- 不顺手进行大规模 schema 重构。
- 不把数据层通用审计列一并视为产品字段删除，除非其确实构成 Templates 前后端链路的一部分。

## UI 紧凑化要求

- 弹窗采用浅色主题，不引入新的深色面板。
- 视觉密度对齐 Connections：
  - 更小的标题层级
  - 更少的垂直空白
  - 更短的段落与说明
  - 更直接的主操作按钮布局
- 表单以两列或适配移动端的一列紧凑栅格呈现。
- 预览区保持轻量摘要，不扩展为第二套说明面板。
- 列表项以“名称 + 核心配置 + 操作”为主，不恢复已删除的冗余 metadata。

### 当前实现状态（T2 / 2026-03-23）

- 已完成 Templates 弹窗、筛选条、列表、Basic / Runtime / Preview section 的首轮紧凑化。
- 已确认保持浅色主题，未引入新的深色面板。
- 已将空状态文案和 Create Task 成功提示收敛为更短的操作导向文案。
- `Create` / `Edit` / `Save As` / `Create Task from Template` 主流程未在本轮中改动其行为语义。

## 冗余解释文案清理要求

- 删除对“连接无关”“将来绑定任务”“未来扩展”等重复说明。
- 删除章节顶部的解释性副标题，只保留必要标题。
- 空状态与提示文案只保留完成当前操作所需的信息。
- 内置模板提示保留最短必要说明，例如“Built-in template. Use Save As to edit a copy.”

### 当前实现状态（T2 / 2026-03-23）

- 已删除或缩短的 T2 相关文案包括：
  - 空列表长说明文案
  - 无结果页按筛选扩展的解释句
  - Create Task from Template 的长成功提示
- 本轮未触碰字段删除、phase 收敛或后端链路整改文案。

## 删除 tags / scope / status / last updated 的产品要求

- 表单中不得再出现这四组字段。
- 列表项中不得再展示这四组字段。
- 预览区和 metadata/badge 中不得再展示这四组字段。
- 筛选器中不得再出现这四组字段。
- 前端/后端 DTO、实体、mock、store、binding、repository、seed、schema 相关产品链路不得依赖这四组字段。
- 若数据库物理列 `updated_at` 仅作为通用持久化审计字段存在，可保留，但不得外溢回 Templates 产品模型与前端展示。

## 删除上述 4 组字段相关筛选、展示、前后端链路的要求

- 前端：
  - 移除 `scope` / `tag` 类筛选条件、选项来源和统计文案。
  - 移除列表项 `status`、`updatedAt`、`tag`、`scope` 展示。
  - 移除编辑弹窗中这些字段的输入、只读显示或说明。
- 后端：
  - DTO 不再暴露这些字段。
  - 领域模型不再依赖这些字段。
  - repository/save/load 不再显式维护这些字段。
  - seeds/mock 不再生成这些字段。
- 兼容层：
  - 若历史 JSON 中仍含旧字段，可被忽略或吸收，但不得重新写回 canonical model。

### 当前实现状态（T3 / 2026-03-23）

- 已确认 Templates 前端以下链路不再展示 `tags` / `scope` / `status` / `last updated`：
  - `TemplateBasicSection.vue`
  - `TemplateFilterBar.vue`
  - `TemplateList.vue`
  - `TemplateListItem.vue`
  - `TemplatePreviewSection.vue`
  - `TemplateEditorDialog.vue`
  - `TemplateToolbar.vue`
- 已确认空状态与筛选提示文案不再引用这四组字段。
- 已确认 mock 数据仅保留单个默认测试模板，前端 mock 不再生成这四组字段。
- 已补充前端测试锁定 store 仅接受 `search` / `dbFamily` / `tool` 三个筛选键，避免旧 `scope/tag/status/updatedAt` 键重新进入前端状态链路。
- 本轮未处理后端 DTO / binding / repository / schema 残留；该部分继续留给 T4。

### 当前实现状态（T4 / 2026-03-23）

- 已审计 Templates 前后端数据链路，结论如下：
  - 前端 `frontend/src/models/template.js` / `frontend/src/stores/template.js` / `frontend/src/mock/templates.js` 已不再把 `tags` / `scope` / `status` / `updatedAt` 作为 canonical 字段使用。
  - `frontend/wailsjs/go/models.ts.body` 的 Templates 相关 `TemplateDTO` / `TemplateListResult` / `TemplateResult` 已不再暴露这四组字段；文件内其它 `status` 命中属于非 Templates 类。
  - 后端 `internal/domain/template/template.go` / `internal/transportwails/bindings/template.go` / `internal/infra/database/repository/template_repo.go` 已不再把这四组字段作为 Templates 业务字段读写。
  - `internal/app/usecase/template_usecase.go` 中原有 `IsReadonlyScope` 命名残留已改为不带 `scope` 业务语义的只读判定。
- 已补充 domain 兼容测试，确认旧模板 JSON 即使仍包含 `scope` / `tags` / `status` / `updatedAt` 也可读取，且 canonical 序列化结果不会再写回这些字段。
- 已确认 `updated_at` 仅继续保留为数据库底层审计列，未重新暴露为 Templates 产品字段。
- 已确认 `Create` / `Edit` / `Save As` / `Create Task from Template` 所依赖的前后端链路不再依赖这四组字段。
- 本轮未处理 phase 收敛，也未改动与 Templates 无关的业务模块。

## phase 审计原则

- 只能依据代码级证据判定 phase 是否真实有用。
- “真实有用”必须同时满足至少一项：
  - 出现在当前 canonical model；
  - 被模板编辑 UI 暴露；
  - 被绑定 DTO 透传；
  - 被执行链明确按 phase 消费，而不是日志、命令构建或兼容映射语义。
- 单纯出现在注释、测试名称、日志文本、变量名或兼容转换逻辑中，不构成保留依据。

## phase 保留/删除判定规则

- 保留：
  - `prepare`：执行链真实使用，UI 真实暴露。
  - `warmup`：执行链真实使用，UI 真实暴露。
  - `run`：核心必选 phase。
  - `cleanup`：执行链真实使用，UI 真实暴露。
- 删除为产品 phase：
  - `build`
  - `generate`
  - `verify`
  - `delete`
- 对于已删除 phase：
  - 若历史数据输入包含这些字段，可在兼容层映射为 `prepare`/`cleanup`。
  - 不得继续在 UI、DTO、mock、文档和测试基线中把它们当作 canonical phase 暴露。

## 数据兼容/迁移要求

- 历史模板 JSON 若仍含 `build` / `generate` / `verify` / `delete`，读取时允许兼容吸收：
  - `build` / `generate` -> `prepare`
  - `verify` / `delete` -> `cleanup`
  - 若这些旧 phase 仍带 `enabled` / `required` / `params`，兼容吸收后应并入 canonical `prepare` / `cleanup`，而不是在重保存时直接丢失
- 历史模板 JSON 若仍含 `tags` / `scope` / `status` / `updatedAt`，读取时不得报错，但 canonical 保存结果中不再保留这些字段。
- 数据库表中的 `updated_at` 若仍作为通用审计列存在，不视为产品字段迁移阻塞。
- migration/seed 逻辑必须保证 builtin template 注册后仍只得到符合四阶段 canonical model 的模板。

### 当前实现状态（T5 / 2026-03-23）

- 已完成 Templates phase 代码级审计，确认 canonical phase 仍只暴露 `prepare` / `warmup` / `run` / `cleanup`，覆盖前端 editor/preview/mock/model、Wails 生成模型、后端 domain/DTO/repository/seed 以及 Create Task handoff。
- 已确认真实执行链消费的 phase 为：
  - `prepare` / `run` / `cleanup`：由 task binding、benchmark usecase、adapter command builder 直接消费。
  - `warmup`：由 benchmark usecase 的 `TaskOptions.WarmupTime` 与 adapter run command 参数消费，但不作为独立模板 action 暴露到 task binding。
- 已确认 `build` / `generate` / `verify` / `delete` 不属于 Templates product phase：
  - `build`：仅出现在“build command/buildschema”日志或工具内部步骤语义中。
  - `generate`：仅作为 Swingbench `wizardOperation`/oewizard 语义存在，不是 phase。
  - `verify`：仅作为 Swingbench prepare 内部 cleanup verification 步骤存在，不是 phase。
  - `delete`：仅作为模板删除动作或旧 JSON 兼容字段存在，不是 benchmark/template phase。
- 已补兼容测试并修正 `PhaseSet.UnmarshalJSON`，使旧模板 JSON 的 legacy phase 配置在读取时吸收到 canonical `prepare` / `cleanup`，但 canonical 序列化仍只写回四阶段。

### 最终实现状态（T6 / 2026-03-23）

- 已完成针对 T1-T5 最终结果的回归验证，未发现需要额外进入 T6 修复分支的阻塞问题。
- UI 与交互回归结论：
  - Templates 创建弹窗、编辑弹窗保持浅色且更紧凑。
  - 解释性文案保持收敛，未回退到旧版长说明。
  - `Create` / `Edit` / `Save As` / `Create Task from Template` 主流程对应前端 store、后端 usecase/binding 与静态断言测试结果一致，构建通过。
- 字段裁剪回归结论：
  - 前端展示、筛选、DTO/binding、domain/usecase/repository 不再把 `tags` / `scope` / `status` / `last updated` 作为 Templates 产品字段使用。
  - 旧 payload 中这四组字段仍可兼容读取，canonical 保存结果不再写回。
- phase 收敛回归结论：
  - canonical phase 继续只保留 `prepare` / `warmup` / `run` / `cleanup`。
  - `build` / `generate` / `verify` / `delete` 仅保留在兼容输入或非 Templates phase 语义上下文。
  - 真实执行链与 Create Task handoff 未受影响。
- 数据与模板收敛回归结论：
  - 系统默认 seed 与前端 fallback 仅保留 `tpl_test_mysql_sysbench`。
  - fresh install / existing data / 默认模板不重复 / 稳定更新既有默认模板记录等场景均已由 repository/usecase/sqlite 测试覆盖。
  - 当前实现不依赖“先删光再重建默认模板”的危险策略；默认模板记录可原位更新并保持依赖关系。

## 验收期 Hotfix 补充（H1 / 2026-03-23）

### 回归现象

- 验收期发现：在 Templates 列表操作默认模板时，`View` / `Edit` / `Copy` 入口会出现前端弹窗错误。
- 报错信息为：`undefined is not an object (evaluating 's.run.required')`

### 根因审计结论

- 直接访问链路位于前端 phase 归一化层，而不是 Runtime/Preview section 模板本身：
  - `frontend/src/models/template.js`
  - `normalizeTemplateRecord(...)`
  - `createDefaultTemplate(...)`
  - `normalizePhasesForTool(...)`
  - `createPhaseState(...)`
- 报错中的 `s` 对应的是 phase state 归一化后的对象；压缩后访问 `s.run.required`，源码对应 `normalized.run.required`。
- 旧实现中 `createPhaseState` 采用对象浅展开：
  - 先给出四阶段默认结构；
  - 再用 `...overrides` 覆盖。
- 当输入为 legacy/旧数据/部分 phase 数据时，如果 `run`、`warmup` 等 key 缺失，或某个 key 被显式传成 `undefined`，浅展开会把默认 `run` 整体覆盖成 `undefined`，随后 `normalizePhasesForTool` 读取 `normalized.run.required` 直接崩溃。
- 默认模板验收场景之所以命中，是因为前端当前需要兼容：
  - canonical 四阶段模板；
  - legacy phase JSON；
  - 缺少部分 phase 字段的旧模板记录。
  当前前端 phase canonical 收敛后，对“部分 phase 输入”的归一化保护不足，属于前端归一化遗漏；它体现为 legacy/旧数据兼容遗漏，而不是默认 seed 自身结构错误。

### 最终修复原则

- 不回退 T5 的四阶段 canonical 收敛结果。
- 不把 `build` / `generate` / `verify` / `delete` 重新暴露到 UI。
- 优先在前端模型归一化层修复，而不是扩大到后端重构。
- 前端无论拿到 canonical、legacy、缺字段 phase 数据，都必须先补齐：
  - `prepare`
  - `warmup`
  - `run`
  - `cleanup`
- 每个 phase 都必须归一化成稳定结构：
  - `enabled`
  - `required`
  - `params`
- View / Edit / Copy / Preview / Runtime 渲染只能消费归一化后的 phase state，不再假设 `run` 等子对象一定已存在。

## 验收标准

- UI：
  - Templates 编辑弹窗为浅色紧凑风格。
  - 无冗余说明文案。
  - 不显示 `tags` / `scope` / `status` / `last updated`。
- 前端数据面：
  - `frontend/src/models/template.js`
  - `frontend/src/stores/template.js`
  - `frontend/src/mock/templates.js`
  - `frontend/wailsjs/go/models.ts.body`
  中不再把上述字段作为 canonical 字段使用。
- 后端数据面：
  - `internal/domain/template/*`
  - `internal/app/usecase/*template*`
  - `internal/infra/database/repository/template_repo.go`
  - `internal/transportwails/bindings/template.go`
  中不再把上述字段作为 Templates 产品链路字段使用。
- phase：
  - canonical phase 仅为 `prepare` / `warmup` / `run` / `cleanup`。
  - `build` / `generate` / `verify` / `delete` 仅允许存在于兼容逻辑或非 phase 语义上下文。
- 主流程：
  - `Create`、`Edit`、`Save As`、`Create Task from Template` 可正常工作。

## 风险与边界

- 仓库当前已有未提交的 Templates 相关改动，本轮整改必须基于现状增量推进，不能直接覆盖或回滚。
- 当前仓库存在与 Templates 无关的测试/构建失败基线，本轮必须区分“当前 task 影响”与“仓库既有问题”。
- 后端 schema 中的通用审计列与业务字段删除不是同一层级，整改时不得混淆。
- Create Task from Template 目前是任务壳交接，不等同于完整执行链；整改不能破坏该 handoff。

### 最终边界与遗留风险（T6 / 2026-03-23）

- 本轮整改至此收口，不继续扩展新功能、新 UI 改造、新字段裁剪或新一轮 phase 改造。
- T6 已完成前端源码断言测试、后端 domain/usecase/binding/repository/sqlite 相关测试、`frontend` 构建与 `go build`；未额外引入浏览器端 E2E 自动化。
- 截至本轮收尾，未发现与 `templates_compact` 目标直接相关的遗留阻塞项；可进入整体收尾状态。
