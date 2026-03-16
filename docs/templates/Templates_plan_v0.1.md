# Templates 模块 Plan v0.1

## 1. 文档信息

- 模块：Templates
- 文档类型：Plan
- 版本：v0.1
- 对应 Spec：`Templates_spec_v0.1.md`

---

## 2. 实施目标

本计划用于指导 Templates 模块前后端基础能力落地。

本阶段目标是：

1. 基于 spec 完成 Templates 页面重构
2. 形成可交互的模板管理工作台
3. 完成基础后端数据模型、持久化与 CRUD API
4. 完成前后端联调并保留前端 validation / normalize 兜底

---

## 3. 实施范围

### 3.1 本轮包含

- 页面整体布局重构
- 模板列表区
- 模板详情区
- 模板编辑态
- 阶段配置 UI
- 工具差异化表单区域
- 模板摘要预览卡片
- 搜索 / 筛选 / 空状态 / 无结果状态
- Built-in / User 模板区分
- 占位保存、复制、删除、创建任务按钮
- mock 数据、类型定义、前端状态管理

### 3.2 本轮不包含

- 真实 Import / Export
- 真实 Create Task 跳转与绑定
- 真实命令预览拼接
- 真实兼容性探测

---

## 4. 总体实施策略

采用以下顺序推进：

1. 先抽象数据模型和 capability matrix
2. 再搭建页面骨架与双栏布局
3. 再接入 mock data 和选择逻辑
4. 再实现详情编辑区与预览区
5. 再补后端持久化 / CRUD / duplicate
6. 最后切换前端到真实后端并保留 mock fallback

---

## 5. 页面结构计划

### 5.1 顶部区

- 标题
- 说明文字
- `New Template`
- `Import`
- `Export`

### 5.2 筛选工具栏

- 搜索框
- Tool 筛选
- DB 筛选
- Scope 筛选
- Tag 筛选
- Reset Filters

### 5.3 左侧模板列表区

- 模板卡片 / 列表项
- 选中态
- Built-in / User 标识
- 空列表态
- 无结果态

### 5.4 右侧详情区

- 空状态
- Basic Information
- Compatibility
- Phases
- Runtime
- Tool-specific Settings
- Preview Summary
- Footer Actions

---

## 6. 组件拆分计划

建议最少拆分为以下组件：

### 6.1 容器级

- `TemplatesTab.vue` / `TemplatesPage.vue`
  - 页面主容器
  - 管理列表、选中模板、编辑状态

### 6.2 顶部与筛选

- `TemplateHeader.vue`
- `TemplateToolbar.vue`

### 6.3 列表区

- `TemplateList.vue`
- `TemplateListItem.vue`
- `TemplateListEmptyState.vue`

### 6.4 详情区

- `TemplateDetailPanel.vue`
- `TemplateBasicSection.vue`
- `TemplateCompatibilitySection.vue`
- `TemplatePhasesSection.vue`
- `TemplateRuntimeSection.vue`
- `TemplateToolSpecificSection.vue`
- `TemplatePreviewCard.vue`
- `TemplateFooterActions.vue`

### 6.5 公共与辅助

- `TemplateEmptyState.vue`
- `TemplateTagGroup.vue`
- `TemplateModeSwitch.vue`

说明：
- 若项目已有通用空状态、标签、卡片、表单行组件，应优先复用。
- 不强制新增过多碎片组件，应以清晰边界和后续可维护性为主。

---

## 7. 状态管理计划

### 7.1 页面主状态

建议主状态至少包含：

- `templates`
- `selectedTemplateId`
- `editingTemplateDraft`
- `isCreating`
- `editorMode`
- `filters`
- `searchKeyword`

### 7.2 状态来源

本阶段建议：

- 使用本地 store 或页面级 composable
- mock 数据集中管理
- 不将 mock 数据散落在多个组件中

### 7.3 推荐结构

- `models/template.ts`
- `mock/templates.ts`
- `stores/template.js` 或 `stores/template.ts`
- `composables/useTemplateFilters.ts`
- `composables/useTemplateEditor.ts`

---

## 8. capability matrix 计划

需要建立一个前端内部能力矩阵，用于：

- 生成默认表单
- 决定可选 benchmark
- 决定哪些字段显示
- 决定哪些字段必填
- 决定工具专属面板内容

### 8.1 第一阶段必须覆盖的规则

- Swingbench 仅允许 Oracle
- Sysbench 仅显示数据库类模板配置
- HammerDB 的 TPROC-C / TPROC-H 区分字段集合
- Built-in 模板不可删除
- User 模板可编辑 / 可删除

### 8.2 可选实现方式

- `templateCapabilities.ts` 中维护声明式配置
- 以 `tool -> workload -> sections -> fields` 组织

---

## 9. mock 数据计划

### 9.1 首批 mock 模板

建议至少准备 8~12 条模板：

#### Sysbench
- MySQL OLTP RW Baseline
- MySQL OLTP RO Baseline
- PostgreSQL OLTP RW Baseline

#### Swingbench
- Oracle OrderEntry Baseline
- Oracle SalesHistory Baseline
- Oracle StressTest Baseline

#### HammerDB

## 10. Test 模板修复实施方案

本节描述 Templates 模块内 `Test` 模板缺失 / 不可见问题的增量修复方案。

### 10.1 根因分析方法

修复前必须完成以下审计，且结论以代码实现为准：

- 后端 seed 是否真实存在 `scope=test` 且可落库的模板
- repository / usecase 是否会保留并返回这些模板
- Wails binding 是否把 `scope` / `tags` / `database_types` / `tool` 原样传给前端
- 前端 store 是否按 `scope`、`tags`、`dbFamily`、`tool` 进行本地过滤
- Templates 到 Tasks & Monitor 的交接是否只传模板 ID，还是依赖额外字段
- 任务创建时实际消费哪些参数键

### 10.2 当前高风险断裂点

本次修复重点排查并修复以下风险：

- 仅 mock 数据存在 `Test` 模板，后端 seed 并不存在
- `scope=test` 与 `tags=["test"]` 语义不统一
- 前端显示 “Test” 但过滤比较值不是 `test`
- `database_types` 与 `dbFamily` 不一致，导致任务页或筛选页行为错位
- 模板能显示但默认参数未映射到真实执行参数键
- 模板能创建任务壳，但进入 prepare / run 链路后因默认值不合理失败

### 10.3 seed 补齐方案

在默认 seed 中补齐每个真实支持数据库至少 1 个 Test 模板：

- MySQL / Sysbench
- PostgreSQL / Sysbench
- Oracle / Swingbench
- SQL Server / HammerDB

每个模板必须：

- 使用 `scope=test`
- `tags` 含 `test`
- 名称清晰标识数据库 + 工具 + 用途
- 采用最小可运行参数
- 与 `dbFamily`、`database_types`、`toolConfig`、`runtime` 保持一致

### 10.4 字段统一方案

统一模板元数据语义：

- `scope=test` 表示系统预置的测试用途模板
- `tags` 补充搜索与标签过滤，不替代 `scope`
- `dbFamily` 作为主数据库类型字段
- `database_types` 作为兼容字段，与 `dbFamily` 强一致
- `tool` 作为唯一 benchmark tool 标识

需要确保 Normalize、repository 存储、binding DTO、前端 normalize 逻辑都不会破坏这种一致性。

### 10.5 前端过滤对齐方案

Templates 页面继续采用本地 store 过滤，但需保证：

- `scope` 过滤直接匹配后端返回的 `scope`
- `tag` 过滤直接匹配后端返回的 `tags`
- `dbFamily` 与 `tool` 使用同一套枚举值
- `visible / total` 使用相同数据源统计

若发现 `Scope / Tag` 组合交互让用户难以选到 `Test`，需在不破坏现有交互前提下修正可见性与筛选结果。

### 10.6 创建任务链路核对方案

修复过程中必须验证：

- Template editor 的 “Create Task from Template” 仅做模板 handoff，不应丢失模板身份
- Tasks & Monitor 通过模板 ID 再次读取模板
- 参数解析来自后端 `resolveParams()`
- Test 模板的 `runtime` / `toolConfig` 能正确映射到 adapter 消费的参数键

重点核对：

- Sysbench: `tables` / `table_size` / `threads` / `time`
- Swingbench: `virtual_users` / `time` / `config_file`
- HammerDB: `virtual_users` / `warehouses` / `scale|duration|rampup|iterations`

### 10.7 验证方案

修复完成后必须完成以下验证：

- 后端自动化测试：
  - seed 数量与覆盖
  - usecase 加载
  - repository builtin/custom 查询
  - task 参数映射
- 前端自动化或脚本验证：
  - 仅 `Test`
  - `Database + Test`
  - `Database + Tool + Test`
- 手工验收：
  - 从 Templates 可见模板
  - 进入 Tasks & Monitor
  - 创建任务预览 / 创建任务
  - 验证 prepare / run / cleanup 或 full pipeline 的最小链路
- Oracle TPROC-C Baseline
- PostgreSQL TPROC-C Baseline
- MySQL TPROC-C Baseline
- PostgreSQL TPROC-H Baseline

### 9.2 mock 数据要求

- 字段完整
- 覆盖多工具、多数据库、多 scope
- 至少包含 built-in 和 user 两种模板
- 至少包含一个可演示高级参数的模板

---

## 10. 交互实现计划

### 10.1 新建模板

- 点击 `New Template`
- 创建 user draft
- 默认选中并进入编辑态
- 使用合理默认值

### 10.2 选择模板

- 点击左侧列表项
- 加载到右侧详情区
- 若当前编辑有未保存改动，先以本地规则处理

### 10.3 保存模板

- 仅更新本地 draft / store
- 显示保存成功占位提示

### 10.4 复制模板

- 复制当前模板为 user scope
- 自动生成新 id 和名称副本

### 10.5 删除模板

- 仅允许 user template
- 确认后从本地列表移除

### 10.6 Create Task from Template

- 当前为占位行为
- 预留事件回调，未来用于导航或传参

---

## 11. 样式与视觉计划

### 11.1 基本原则

- 与现有深色主题一致
- 参考 `Connections` 页面作为视觉基线
- 使用统一的间距、边框、圆角、hover/focus 行为

### 11.2 重点打磨点

- 左右双栏比例
- 列表选中态
- 表单分组标题层级
- 标签视觉统一
- 空状态不可过于空洞
- 大面积空白区需要承担说明和预览职责

### 11.3 避免问题

- 不做成只有一个下拉框的页面
- 不把所有字段堆成超长单表单
- 不让 Sysbench / Swingbench / HammerDB 混在同一组静态字段中
- 不做出与现有导航页割裂的样式

---

## 12. 实施阶段划分

### 阶段 1：调研与文档

产出：
- 前置设计文档
- Spec
- Plan
- Tasks

### 阶段 2：数据模型与状态层

产出：
- template 类型定义
- capability matrix
- mock 数据
- store / composables

### 阶段 3：页面骨架

产出：
- 双栏布局
- 顶部区
- 筛选工具栏
- 空状态

### 阶段 4：列表与详情区

产出：
- 模板列表区
- 详情区基本信息
- 基础编辑态

### 阶段 5：工具差异化与预览

产出：
- Phases UI
- Runtime UI
- Tool-specific UI
- Preview card

### 阶段 6：交互完善与打磨

产出：
- 搜索 / 筛选
- 新建 / 保存 / 复制 / 删除占位
- 无结果状态
- 样式细化

---

## 13. 风险与应对

### 风险 1：三工具差异大，界面容易失控

应对：
- 采用公共模型 + 工具专属模型
- 使用 capability matrix
- 引入 standard / advanced / expert 模式

### 风险 2：字段过多导致页面难用

应对：
- 分组展示
- 默认只展示高频字段
- 高级参数折叠或切模式显示

### 风险 3：前端先行后，后端对接时字段冲突

应对：
- 在类型与模型层统一命名
- 预留 toolConfig 扩展对象
- 保持阶段模型清晰

### 风险 4：Swingbench Oracle 11g 兼容性表述过满

应对：
- 文档中明确写“以实际集成验证为准”
- 模板设计支持该目标，但实现阶段保留兼容性标记

---

## 14. 后端对接预留点

本阶段需预留以下接口边界：

- `loadTemplates()`
- `createTemplate()`
- `updateTemplate()`
- `deleteTemplate()`
- `duplicateTemplate()`
- `exportTemplate()`
- `importTemplate()`
- `createTaskFromTemplate()`

说明：
- 这些方法当前可以是空实现、mock 实现或 store 层占位。
- 关键是前端调用边界要先形成。

---

## 15. 完成定义（Definition of Done）

满足以下条件视为本阶段完成：

1. 文档齐全且相互一致
2. Templates 页面完成双栏重构
3. mock 模板可浏览、可选中、可编辑
4. Built-in / User 模板可区分
5. 页面支持新建、复制、删除、保存占位流程
6. 页面支持搜索、筛选、空状态、无结果状态
7. UI 风格与现有系统统一
8. 代码结构为后续接入后端留有明确边界
