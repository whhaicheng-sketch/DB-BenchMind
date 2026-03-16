# Templates 模块 Spec v0.1

## 1. 文档信息

- 模块：Templates
- 文档类型：Spec
- 版本：v0.1
- 当前阶段：前后端联调 / 基础后端已接入
- 关联前置文档：`Templates_前置设计文档_v0.1.md`

---

## 2. 背景

当前项目中的 `Connections` 页面已具备较成熟的交互与视觉基础，但 `Templates` 页面仍停留在极简占位状态，尚未形成模板管理、模板编辑、模板分类、模板预览等完整能力。

项目目标是支持以下三类数据库压测工具，并通过 GUI 尽量承载其数据库压测能力：

- Sysbench（仅数据库压测相关能力）
- Swingbench 2.7
- HammerDB

同时，产品边界已明确：

1. Templates 仅面向数据库压测场景。
2. Templates 不绑定具体 Connection。
3. 真正运行时，在 `Tasks & Monitor` 中选择并绑定 Connection。
4. 结果导出、结果比较不属于 Templates 模块范围。
5. 当前阶段已接入 Templates 基础后端持久化、CRUD API 与前端联调；任务执行映射仍暂缓。

---

## 3. 目标

### 3.1 产品目标

将 Templates 模块建设为：

> 一个连接无关、阶段化、支持三类压测工具公共模型与专属模型并存的数据库压测场景模板系统。

### 3.2 本阶段目标

本阶段重点目标：

- 建立清晰的模板数据模型
- 明确模板页面的信息架构与编辑模式
- 支持模板列表、模板详情、模板编辑、模板预览
- 支持内置模板与用户模板的区分
- 支持搜索、筛选、空状态、无结果状态
- 接入真实后端持久化与 CRUD API
- 为后续任务执行映射预留结构

---

## 4. 非目标

本阶段不做以下内容：

- 不接真实模板导入/导出解析逻辑
- 不接真实任务创建与下发逻辑
- 不接真实工具命令执行
- 不做结果导出、结果对比、报告生成
- 不处理连接测试、权限测试、真实兼容性探测

---

## 5. 用户与使用场景

### 5.1 目标用户

- 数据库压测工程师
- DBA / 性能测试工程师
- 需要重复执行标准化 benchmark 场景的用户

### 5.2 典型场景

1. 用户想快速创建一个 Sysbench MySQL OLTP 模板，并多次复用。
2. 用户想定义一个 Swingbench Oracle OrderEntry 场景模板，但稍后再在任务中绑定具体连接。
3. 用户想为 HammerDB 定义一套包含 build / run / delete 的 TPROC-C 场景模板。
4. 用户想从内置模板复制一份用户模板，再按环境进行微调。

---

## 6. 核心定义

### 6.1 Template 定义

Template 不是“参数保存器”，而是：

> 一个对数据库压测场景进行声明式描述的对象。

它描述：

- 使用的工具
- 面向的数据库类型
- 具体 benchmark / workload
- 执行阶段
- 运行参数
- 工具专属参数
- 兼容性约束

### 6.2 连接关系定义

- Template：只描述 workload 场景
- Connection：只描述连接能力
- Task：在运行阶段把 Template 与 Connection 进行绑定

---

## 7. 功能范围

### 7.1 必须支持的工具

- Sysbench（数据库压测部分）
- Swingbench 2.7
- HammerDB

### 7.2 必须支持的模板能力

- 新建模板
- 查看模板
- 编辑模板
- 复制模板
- 删除用户模板
- 后端持久化 CRUD
- Built-in / Readonly Shared 后端只读保护
- 区分内置模板 / 用户模板
- 按工具、数据库、标签、关键字筛选
- 模板摘要预览
- 模板阶段配置
- 模板基础校验

### 7.3 首阶段支持的模板家族

#### Sysbench
- MySQL / PostgreSQL
- 数据库类 OLTP 模板
- 典型脚本：
  - Read Write
  - Read Only
  - Write Only
  - Point Select
  - Update Index
  - Update Non-Index
  - Insert

#### Swingbench 2.7
- Oracle-only
- 首批重点 benchmark：
  - OrderEntry
  - SalesHistory
  - StressTest
- 其他 benchmark 作为能力预留：
  - JSON
  - TPC-H Like
  - TPC-DS Like
  - MovieStream

#### HammerDB
- 首批重点：TPROC-C
- 同步建模但不追求首版深度易用：TPROC-H
- 支持数据库家族：
  - Oracle
  - SQL Server
  - Db2
  - PostgreSQL
  - MySQL
  - MariaDB

## 7.4 Test 模板预置与筛选一致性修复

本节定义 Templates 模块内对 `Test` 模板能力的增量修复要求。该修复属于现有 Templates 模块范畴，不单独建立平行 spec。

### 7.4.1 判定原则

- 支持矩阵必须以代码实现为准，不以 README、既有 spec、用户假设为准。
- “Test 模板已支持” 的定义必须同时满足：
  - 模板真实存在于 seed / repository / API 返回数据中。
  - Templates 页面在 `Test` 相关筛选下可见。
  - 模板可从 Templates 页面进入任务创建链路。
  - 模板参数被真实执行链路消费，而不是仅停留在展示层。
- 最终结论必须给出代码定位点。

### 7.4.2 支持矩阵确认要求

实现时必须先通过代码确认以下事实：

- 系统真实支持哪些数据库类型。
- 每种数据库类型对应哪些 benchmark tool 能被真实执行。
- Templates 页面允许用户为指定数据库选择哪些工具。
- 任务创建与执行链路实际消费哪些模板字段与参数键。

支持矩阵若与既有文档冲突，以代码事实为准，并同步修正文档。

### 7.4.3 Test 模板覆盖要求

对于每个代码中真实支持、且在当前产品范围内对外暴露的数据库，系统必须至少预置 1 个 `Test` 模板。

首批必须覆盖的组合以当前代码能力为基线：

- MySQL -> Sysbench Test
- PostgreSQL -> Sysbench Test
- Oracle -> Swingbench Test
- SQL Server -> HammerDB Test

若某数据库在代码中不支持真实执行，则不得伪造模板；必须在实现与汇报中明确说明不支持原因及代码证据。

### 7.4.4 Test 模板命名规范

Test 模板命名必须让用户一眼识别数据库、工具与用途，推荐格式：

- `<Database Label> - <Tool Label> Test`

例如：

- `MySQL - Sysbench Test`
- `PostgreSQL - Sysbench Test`
- `Oracle - Swingbench Test`
- `SQL Server - HammerDB Test`

### 7.4.5 元数据统一规则

所有 Test 模板必须统一以下语义：

- `scope` 必须为 `test`
- `tags` 中必须包含 `test`
- `dbFamily` / `database_types` 必须与实际数据库一致
- `tool` 必须与真实 benchmark tool 一致
- `name` / `description` 必须体现其为最小可运行的功能验证模板

前端筛选使用的字段值必须与后端返回值完全一致，不允许出现仅 label 显示为 `Test`、但 value 与实际存储不一致的情况。

### 7.4.6 最小可运行参数原则

Test 模板不是性能基线模板，而是功能链路模板。所有 Test 模板必须满足：

- 数据规模最小但非 0
- 并发最小但非 0
- 执行时长尽可能短
- 资源占用低
- 能覆盖 prepare / run / monitor / result 的最小链路

各工具应遵循以下约束：

- Sysbench：
  - 最少表数
  - 最小 `table_size`
  - 最小 `threads`
  - 短时 `time`
  - 参数键必须与执行链路消费的 `tables` / `table_size` / `threads` / `time` 对齐
- Swingbench：
  - 最小可运行 `scale`
  - 最小 `virtual_users`
  - 最短可接受运行时长
  - 参数键必须与执行链路消费的 `virtual_users` / `time` / `config_file` 等键一致
- HammerDB：
  - 最小 `warehouses`
  - 最小 `virtual_users`
  - 最短 `duration`
  - 最低必要的 buildschema / run 参数
  - 参数键必须与执行链路消费的 `virtual_users` / `warehouses` / `duration` / `rampup` / `iterations` 等键一致

### 7.4.7 Templates 页面筛选验收标准

下列筛选必须返回可见模板，不得出现“0 visible / N total 但理论上存在 Test 模板”的情况：

- 仅选择 `Test`
- 选择 `Database Type + Test`
- 选择 `Benchmark Tool + Test`
- 选择 `Database Type + Benchmark Tool + Test`

`visible / total` 的统计必须以页面当前模板集合与前端实际过滤结果为准。

### 7.4.8 任务创建与运行验收标准

每个 Test 模板都必须可用于真实创建任务，而非仅展示：

- 可以从 Templates 页面进入任务创建链路
- 在 Tasks & Monitor 中可被选中
- 支持当前系统允许的 prepare / run / cleanup / full pipeline 行为
- 能产出基础监控或结果数据
- 不因模板默认参数设计不当而天然失败

### 7.4.9 回归测试要求

至少覆盖以下层级：

- 后端：
  - 默认 seed 模板初始化
  - repository 读写与 builtin/custom 查询
  - template usecase 加载与覆盖验证
  - task 参数解析与 template snapshot 验证
- 前端：
  - Templates 页 `Test` 过滤
  - `Database + Test`
  - `Database + Tool + Test`
  - `visible / total` 结果正确

若前端缺少单元测试框架，则必须补可重复执行的页面级验证脚本，并在最终汇报中给出手工验证路径与结果。

---

## 8. 数据模型

## 8.1 顶层模型

```json
{
  "id": "tpl_xxx",
  "name": "Sysbench-MySQL-OLTP_RW-10x100k-32th-300s",
  "description": "MySQL OLTP read/write baseline template",
  "tool": "sysbench",
  "dbFamily": "mysql",
  "workloadFamily": "oltp-read-write",
  "scope": "builtin",
  "status": "ready",
  "tags": ["mysql", "sysbench", "oltp", "baseline"],
  "version": "1.0.0",
  "compatibility": {},
  "phases": {},
  "runtime": {},
  "toolConfig": {},
  "createdAt": "",
  "updatedAt": ""
}
```

### 8.2 元数据字段

- `id`
- `name`
- `description`
- `tool`
- `dbFamily`
- `workloadFamily`
- `scope`：`builtin | user`
- `status`：`draft | ready | deprecated`
- `tags`
- `version`
- `createdAt`
- `updatedAt`

### 8.3 兼容性字段

- `supportedDatabases`
- `supportedVersions`
- `compatibilityNotes`
- `requiresPrivileges`
- `constraints`

### 8.4 阶段字段

统一阶段集合：

- `build`
- `prepare`
- `generate`
- `warmup`
- `run`
- `verify`
- `cleanup`
- `delete`

每阶段包含：

- `enabled`
- `required`
- `params`

### 8.5 运行字段

- `concurrency.mode`
  - `threads`
  - `users`
  - `virtualUsers`
- `concurrency.value`
- `durationSeconds`
- `warmupSeconds`
- `rampUpSeconds`
- `rateLimit`
- `iterations`
- `reportIntervalSeconds`
- `percentile`
- `validationEnabled`
- `notes`

### 8.6 工具专属字段

#### Sysbench
- `dbDriver`
- `scriptType`
- `tables`
- `tableSize`
- `rangeSize`
- `pointSelects`
- `simpleRanges`
- `sumRanges`
- `orderRanges`
- `distinctRanges`
- `indexUpdates`
- `nonIndexUpdates`
- `deleteInserts`
- `skipTrx`
- `secondary`
- `createSecondary`
- `mysqlStorageEngine`
- `pgsqlVariant`
- `extraCliArgs`

#### Swingbench
- `benchmark`
- `frontend`
- `configMode`
- `wizardType`
- `wizardOperation`
- `driverType`
- `userCount`
- `runTimeSeconds`
- `minThinkTime`
- `maxThinkTime`
- `minSleepTime`
- `maxSleepTime`
- `autoStart`
- `cpuMonitorLocation`
- `verboseMetrics`
- `xmlOverrides`

说明：
- `prepare` 对应 Swingbench wizard 语义，当前 Oracle 基线场景使用 `oewizard` 负责 schema build / data generation。
- `run` 对应 workload frontend 语义，自动化执行默认使用 `charbench`，而不是继续停留在 wizard/create 阶段。
- `configMode = managed` 时，内置 Oracle OrderEntry 模板必须能解析出官方配置文件，不要求用户再手填 `config_file`。
- `xmlOverrides` 用于覆盖官方/托管 XML；若指定则优先于 managed config。

#### HammerDB
- `benchmark`
- `buildSchemaEnabled`
- `checkSchemaEnabled`
- `deleteSchemaEnabled`
- `dataGenEnabled`
- `loadScriptEnabled`
- `virtualUsers`
- `delayMs`
- `repeatDelayMs`
- `iterations`
- `showOutput`
- `jobsEnabled`
- `metricsEnabled`
- `txCounterEnabled`
- `warehouses`
- `useAllWarehouses`
- `timeProfile`
- `fixedThroughput`
- `xmlConnectPool`
- `stepTesting`
- `storedProcedureMode`
- `prepareStatements`
- `purgeWhenComplete`
- `scaleFactor`
- `powerTest`
- `throughputTest`
- `geometricMeanMode`
- `datagenDirectory`
- `buildVirtualUsers`

---

## 9. 名称规范

### 9.1 命名原则

名称用于快速识别，不承载全部执行语义。真正参与执行的内容必须结构化存储。

### 9.2 推荐格式

```text
{Tool}-{DB}-{Workload}-{Scale}-{Concurrency}-{Duration}
```

### 9.3 示例

- `Sysbench-MySQL-OLTP_RW-10x100k-32th-300s`
- `Swingbench-Oracle-OE-Medium-64u-30m`
- `HammerDB-PostgreSQL-TPROC-C-1000W-64vu-20m`

### 9.4 必须进入名称的信息

- Tool
- DB family
- Workload
- Dataset scale
- Concurrency
- Duration

### 9.5 不应进入名称的信息

- Connection 名称
- Host / Port
- 用户名 / 密码
- 结果文件路径
- 输出目录

---

## 10. 页面信息架构

### 10.1 页面整体结构

Templates 页面采用：

- 顶部标题区
- 筛选工具栏
- 左侧模板列表区
- 右侧模板详情 / 编辑区

### 10.2 标题区

包含：

- 页面标题：Templates
- 一句话说明
- 操作按钮：
  - New Template
  - Import
  - Export

### 10.3 筛选工具栏

包含：

- Search
- Tool Filter
- DB Filter
- Scope Filter
- Tags Filter
- Reset Filters

### 10.4 左侧模板列表

列表项展示：

- 模板名称
- 副标题 / 描述摘要
- Tool Tag
- DB Tag
- Scope Tag
- Updated At

列表区支持：

- 选中高亮
- 空列表态
- 搜索无结果态
- Built-in / User 视觉区分

### 10.5 右侧详情区

未选中模板时显示空状态。

选中模板后展示：

1. Basic Information
2. Compatibility
3. Phases
4. Runtime Settings
5. Tool-specific Settings
6. Preview Summary
7. Footer Actions

---

## 11. 编辑模式

模板编辑器采用三层模式：

### 11.1 Standard

面向大多数用户，仅展示高频字段：

- 工具
- 数据库
- workload
- phases
- 并发
- 时长
- 规模参数

### 11.2 Advanced

展示工具高级参数：

- Sysbench workload mix / storage engine
- Swingbench wizard / think time / frontend
- HammerDB driver options / metrics / step testing

### 11.3 Expert

保留为能力兜底：

- Sysbench extra CLI args
- Swingbench XML overrides
- HammerDB 高级脚本/参数覆盖

---

## 12. 状态流转

### 12.1 页面状态

- 初始未选中模板
- 已加载模板列表
- 列表为空
- 搜索无结果
- 新建模板态
- 编辑模板态
- 占位保存成功提示态
- 占位未实现提示态

### 12.2 模板编辑状态

- `view`
- `editing`
- `dirty`
- `saving-placeholder`
- `saved-placeholder`

### 12.3 作用域状态

- `builtin`：可查看、可复制，不可直接删除
- `user`：可编辑、可删除、可复制

---

## 13. 动态显示与能力约束

系统必须维护 capability matrix，用于控制：

- 可选 benchmark
- 可选数据库
- 可启用阶段
- 必填字段
- 高级参数显示条件

### 13.1 规则示例

- `tool = swingbench` 时，仅允许 Oracle 模板族。
- `tool = sysbench` 时，不显示非数据库 benchmark 相关项。
- `tool = hammerdb && benchmark = tproc-c` 时，显示 warehouses 系列字段。
- `tool = hammerdb && benchmark = tproc-h` 时，显示 scale factor / power / throughput 系列字段。

---

## 14. 交互规则

### 14.1 新建模板

- 创建一个默认 user template
- 默认进入编辑态
- 默认填入推荐初始值

### 14.2 复制模板

- 内置模板可复制为用户模板
- 用户模板可复制为新用户模板
- 复制后的名称自动追加 `Copy`

### 14.3 删除模板

- 仅允许删除用户模板
- 内置模板不允许删除
- 删除前需确认

### 14.4 保存模板

当前阶段为真实后端行为：

- 调用 Templates 后端 Create / Update
- 持久化到 SQLite
- 显示成功提示

### 14.5 Create Task from Template

当前阶段为占位能力：

- 触发前端提示
- 为后续跳转到 `Tasks & Monitor` 预留接口

---

## 15. 校验规则

### 15.1 通用校验

- `name` 必填
- `tool` 必填
- `dbFamily` 必填
- `workloadFamily` 必填
- `run` 阶段至少为 enabled
- `concurrency.value` 必须大于 0
- `durationSeconds` 或等效运行参数必须有效

### 15.2 工具校验

#### Sysbench
- `dbDriver` 必填
- `scriptType` 必填
- `tables` / `tableSize` 在 prepare 场景下必须有效

#### Swingbench
- `benchmark` 必填
- `frontend` 必填
- `wizardOperation` 在 wizard 场景下必须有效
- 模板内部可保留 Swingbench wizard 参数，但模板对 Tasks & Monitor 暴露的执行动作语义必须统一为 `prepare / run / cleanup`
- `configMode = managed` 时，`run` 阶段必须能解析出有效的官方 XML 配置；若无法解析，模板应视为未就绪

#### HammerDB
- `benchmark` 必填
- `virtualUsers` 在 run 场景下必须有效
- `warehouses` 或 `scaleFactor` 依据 benchmark 分别校验

---

## 16. 内置模板建议

### 16.1 Sysbench
- MySQL OLTP Read Write Baseline
- MySQL OLTP Read Only Baseline
- PostgreSQL OLTP Read Write Baseline

### 16.2 Swingbench
- Oracle OrderEntry Baseline
- Oracle SalesHistory Baseline
- Oracle StressTest Baseline

### 16.3 HammerDB
- Oracle TPROC-C Baseline
- PostgreSQL TPROC-C Baseline
- MySQL TPROC-C Baseline
- PostgreSQL TPROC-H Baseline

---

## 17. 验收标准

### 17.1 文档层

- Spec、Plan、Tasks 三份文档齐备
- 与前置设计文档保持一致
- 明确边界、阶段、模型、页面结构

### 17.2 页面层

- 页面从单下拉框升级为模板管理工作台
- 页面具备列表区、详情区、筛选工具栏、空状态
- 能表现内置模板与用户模板
- 能展示模板预览摘要

### 17.3 状态层

- 支持后端模板数据加载
- 支持新建、编辑、复制、删除（真实后端 CRUD）
- 支持筛选、搜索、无结果态
- 支持阶段配置与工具差异化表单

### 17.4 范围层

- 已接入真实后端持久化与基础 CRUD
- 为后续任务执行映射保留接口与结构
- 不将模板和 connection 强绑定

---

## 18. 后续演进方向

- 模板导入 / 导出
- 模板版本历史
- 模板校验与兼容性探测
- 从模板直接创建任务
- 模板与任务执行映射可视化
