# Templates 前置设计文档 v0.1

## 1. 文档目的

本文档用于在正式进入 `spec / plan / tasks` 之前，先完成 Templates 模块的前置设计。

本轮目标不是直接实现后端逻辑，而是先把以下内容定义清楚：

- Templates 模块在产品中的职责边界
- Sysbench、Swingbench 2.7、HammerDB 三种压测工具在 GUI 中的承载方式
- 模板的数据抽象方式
- 模板页面和编辑器的总体结构
- 模板命名、阶段建模、能力约束、字段分层等核心设计原则

---

## 2. 设计背景

当前项目已经具备较成熟的 `Connections` 页面，但 `Templates` 页面仍处于非常早期的状态，尚未形成真正可用的模板管理与编辑能力。

本项目后续希望支持以下三种数据库压测工具：

- Sysbench
- Swingbench 2.7
- HammerDB

并希望 GUI 能尽量覆盖三种工具在数据库压测场景下的主要能力。

根据当前确认的产品边界：

1. Templates 只面向 **数据库压测**，不纳入 sysbench 的 CPU / memory / fileio / mutex / threads 等非数据库 benchmark。
2. Templates **不绑定具体 Connection**，只描述压测场景；真正运行时，在 `Tasks & Monitor` 中选择 Connection 进行绑定。
3. 结果导出、结果比较不放在 Templates 中处理，而由其他页面负责。
4. Swingbench 以 **2.7** 为能力基线，并以兼容 Oracle 11g 作为设计目标之一。

---

## 3. 需求设计

### 3.1 Templates 的产品定义

Templates 不能被定义成“若干表单参数的保存器”，而应定义为：

> **一个连接无关的数据库压测场景模板（Benchmark Scenario Template）。**

一个模板应当能够描述：

- 用哪个工具压测
- 面向哪类数据库
- 使用什么 workload / benchmark
- 涉及哪些执行阶段
- 运行时的并发与时长模型
- 工具专属的高级参数
- 能力边界与适用范围

### 3.2 核心设计目标

Templates 模块需要满足以下目标：

1. **可复用**：同一压测场景可以重复使用
2. **可扩展**：可以承接三种工具不同的能力模型
3. **可约束**：根据工具与数据库自动限制字段和选项
4. **可预览**：用户能清楚知道该模板将执行什么场景
5. **连接无关**：模板本身不绑定 host/user/password/connectionId
6. **可演进**：后续可以平滑接入后端持久化与执行逻辑

### 3.3 核心需求结论

#### 需求 A：模板必须是阶段化的

模板不能只保存 `threads + duration + benchmark name`，而必须支持阶段建模。

建议阶段集合：

- `build`
- `prepare`
- `generate`
- `warmup`
- `run`
- `verify`
- `cleanup`
- `delete`

原因：

- Sysbench 有 `prepare / run / cleanup`，数据库 OLTP 脚本还有 warmup 相关能力
- Swingbench 具备 wizard 阶段，如 `create / drop / generate`
- HammerDB 明确是 workflow 型工具，包含 `buildschema / checkschema / deleteschema / datagenrun / loadscript / vucreate / vurun`

补充说明：
- Oracle Swingbench 的 `prepare` 应映射到 `oewizard` 等 wizard/schema build 能力。
- Oracle Swingbench 的 `run` 应映射到 `charbench` 这类 workload frontend，而不是继续执行 schema create。
- 因此 `prepare` 阶段没有真实 TPS / TPM 时显示 `0` 可以接受，但 `run` 阶段必须有明确吞吐来源。

#### 需求 B：模板必须是“公共模型 + 工具专属模型”

三种工具不能被粗暴压平到一张简单表单中。

因此模板必须分层：

- **公共字段层**：名称、描述、工具、数据库类型、workload 类型、阶段开关、并发、时长等
- **工具专属字段层**：
  - Sysbench 专属参数
  - Swingbench 专属参数
  - HammerDB 专属参数

#### 需求 C：模板必须连接无关

模板仅描述 workload 场景，不保存：

- host
- port
- username
- password
- connectionId
- schema owner 账号信息

这些信息只在真正创建任务或执行任务时于 `Tasks & Monitor` 中绑定。

#### 需求 D：模板必须有能力约束与动态校验

GUI 不允许“所有字段都能填”。

系统必须根据：

- tool
- dbFamily
- workloadFamily
- phase
- editor mode

动态决定：

- 哪些 benchmark 可选
- 哪些字段显示
- 哪些字段必填
- 哪些组合非法
- 哪些选项仅在特定数据库上出现

---

## 4. 总体设计

### 4.1 顶层抽象

建议把 Template 定义为：

> **一个数据库压测场景的声明式描述对象。**

它由五层构成：

1. 元数据层
2. 兼容性层
3. 阶段层
4. 运行层
5. 工具扩展层

### 4.2 元数据层

负责模板的身份与管理属性，不直接承载执行细节。

建议字段：

- `id`
- `name`
- `description`
- `tool`
- `dbFamily`
- `workloadFamily`
- `scope`（`builtin | user`）
- `tags`
- `status`（`draft | ready | deprecated`）
- `version`
- `createdAt`
- `updatedAt`

### 4.3 兼容性层

负责描述“模板适用于哪些对象”。

建议字段：

- `supportedDatabases`
- `supportedVersions`
- `compatibilityNotes`
- `requiresPrivileges`
- `constraints`

说明：

- Swingbench 设计上按 Oracle-only 工具对待
- 目标兼容 Oracle 11g，但最终仍以集成验证结果为准

### 4.4 阶段层

负责描述模板会经历哪些执行阶段。

建议结构：

```json
{
  "phases": {
    "build": { "enabled": false, "required": false, "params": {} },
    "prepare": { "enabled": false, "required": false, "params": {} },
    "generate": { "enabled": false, "required": false, "params": {} },
    "warmup": { "enabled": false, "required": false, "params": {} },
    "run": { "enabled": true, "required": true, "params": {} },
    "verify": { "enabled": false, "required": false, "params": {} },
    "cleanup": { "enabled": false, "required": false, "params": {} },
    "delete": { "enabled": false, "required": false, "params": {} }
  }
}
```

### 4.5 运行层

负责描述压力模型与执行节奏。

建议字段：

- `concurrency.mode`：`threads | users | virtualUsers`
- `concurrency.value`
- `durationSeconds`
- `warmupSeconds`
- `rampUpSeconds`
- `iterations`
- `rateLimit`
- `reportInterval`
- `percentile`
- `validationEnabled`
- `notes`

### 4.6 工具扩展层

负责承载三种工具的专属参数。

建议结构：

```json
{
  "toolConfig": {
    "sysbench": {},
    "swingbench": {},
    "hammerdb": {}
  }
}
```

---

## 5. 三种工具的能力建模

### 5.1 Sysbench 模板设计

#### 设计边界

仅纳入数据库 benchmark，不纳入：

- cpu
- memory
- fileio
- mutex
- threads

#### 建议支持的数据库类型

- `mysql`
- `postgresql`

#### 建议支持的 workload family

- `oltp_read_write`
- `oltp_read_only`
- `oltp_write_only`
- `oltp_point_select`
- `oltp_update_index`
- `oltp_update_non_index`
- `oltp_insert`

#### 建议字段

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

#### 特别说明

Sysbench 存在两种“预热”概念，应区分建模：

1. `phase.warmup`：阶段级 warmup
2. `runtime.warmupSeconds`：运行级 warmup time

两者不应混为一个字段。

---

### 5.2 Swingbench 2.7 模板设计

#### 设计边界

- 按 Oracle-only 工具设计
- 以 2.7 为能力基线
- 以兼容 Oracle 11g 为设计目标，但最终以集成验证为准

#### 建议支持的 benchmark family

- `orderEntry`
- `salesHistory`
- `stressTest`
- `json`
- `tpcdsLike`
- `tpchLike`
- `movieStream`

#### 首批推荐内置模板建议

为了兼顾 Oracle 11g 兼容目标与首版可控性，建议首批把以下 benchmark 作为一等公民：

- `orderEntry`
- `salesHistory`
- `stressTest`

其余 benchmark 可以先进入高级模板能力矩阵，但不一定作为首版推荐模板。

#### 建议字段

- `benchmark`
- `frontend`：`swingbench | minibench | charbench`
- `configMode`：`managed | importXml | overrideOnly`
- `wizardType`
- `wizardOperation`：`create | drop | generate`
- `driverType`
- `userCount`
- `runTime`
- `minThinkTime`
- `maxThinkTime`
- `minSleepTime`
- `maxSleepTime`
- `autoStart`
- `cpuMonitorLocation`
- `verboseMetrics`
- `xmlOverrides`

---

### 5.3 HammerDB 模板设计

#### 设计边界

HammerDB 必须按 workflow 工具建模，而不是只做单次 run 表单。

#### 建议支持的数据库类型

- `oracle`
- `sqlserver`
- `db2`
- `postgresql`
- `mysql`
- `mariadb`

#### 建议支持的 workload family

- `tproc-c`
- `tproc-h`

#### 建议公共字段

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

#### TPROC-C 建议字段

- `warehouses`
- `useAllWarehouses`
- `timeProfile`
- `fixedThroughput`
- `xmlConnectPool`
- `stepTesting`
- `storedProcedureMode`
- `prepareStatements`
- `purgeWhenComplete`

#### TPROC-H 建议字段

- `scaleFactor`
- `powerTest`
- `throughputTest`
- `geometricMeanMode`
- `datagenDirectory`
- `buildVirtualUsers`

#### 建议落地策略

- 数据模型同时支持 `TPROC-C + TPROC-H`
- 首版 GUI 优先把 `TPROC-C` 做深
- `TPROC-H` 先做到“可建模、可编辑、可预留执行映射”

---

## 6. Name Templates 设计

### 6.1 设计原则

`name` 用于快速识别模板，不承载全部业务语义。

所有真正会影响执行和校验的参数，必须结构化存储，而不是依赖名字解析。

### 6.2 推荐命名模板

```text
{Tool}-{DB}-{Workload}-{Scale}-{Concurrency}-{Duration}
```

### 6.3 示例

- `Sysbench-MySQL-OLTP_RW-10x100k-32th-300s`
- `Swingbench-Oracle-OE-Medium-64u-30m`
- `HammerDB-PostgreSQL-TPROC-C-1000W-64vu-20m`

### 6.4 建议必须进入名字的值

- Tool
- DB family
- Workload
- Dataset scale
- Concurrency
- Duration

### 6.5 条件性追加到名字的值

仅在确实有区分价值时追加：

- Phase mode：`RunOnly / PrepareRun / BuildRun`
- Variant：`RO / RW / WO / Power / Throughput`
- Access mode：`JDBC / PLSQL`
- Advanced mode：`Step / Pool / FixedRate`

例如：

- `HammerDB-Oracle-TPROC-C-1000W-64vu-30m-Step`
- `Swingbench-Oracle-OE-PLSQL-128u-15m`
- `Sysbench-PG-RO-20x1M-64th-600s-FixedRate`

### 6.6 不应进入名字的值

以下信息不应进入模板名：

- connection 名称
- host / port
- username
- password
- output path
- result file path

---

## 7. 概要设计

### 7.1 编辑器模式设计

为了同时兼顾易用性和能力覆盖，模板编辑器建议采用三层模式：

#### 标准模式

面向大多数用户，仅暴露高频字段：

- 工具
- 数据库类型
- benchmark / workload
- 阶段选择
- 并发
- 时长
- 基础规模参数

#### 高级模式

暴露工具专属的高级参数：

- Sysbench：workload mix、storage engine、pgsql variant 等
- Swingbench：frontend、wizard、think/sleep time、monitor 等
- HammerDB：driver options、jobs、metrics、step testing、fixed throughput 等

#### 专家模式

作为能力兜底：

- Sysbench：extra CLI args
- Swingbench：XML overrides
- HammerDB：driver script / CLI override

说明：

如果没有专家模式，就很难说 GUI 已经“尽量包边三种工具能力”；
如果一开始就把所有参数平铺，又会让普通用户几乎无法使用。

### 7.2 能力矩阵设计

需要构建内部 capability matrix，用于驱动 GUI 的字段显示与校验逻辑。

能力矩阵至少根据以下维度工作：

- `tool`
- `dbFamily`
- `workloadFamily`
- `phase`
- `editorMode`

典型规则：

- `tool=swingbench` 时，只允许 Oracle 分支
- `tool=sysbench` 时，不出现非数据库 benchmark
- `tool=hammerdb + workload=tproc-h` 时，显示 `scaleFactor / powerTest / throughputTest`
- `tool=hammerdb + workload=tproc-c` 时，显示 `warehouses / useAllWarehouses / stepTesting`

### 7.3 模板分类

建议将模板分成四类：

1. `Built-in Recommended`
2. `Built-in Advanced`
3. `User Custom`
4. `Compatibility / Experimental`

这样有利于在 UI 中区分：

- 推荐模板
- 高级模板
- 用户自定义模板
- 需额外验证的模板族

### 7.4 校验设计

模板保存前至少需要三类校验：

#### 结构校验

- 字段是否齐全
- 类型是否正确
- 阶段冲突是否存在

#### 能力校验

- 所选工具是否支持该数据库类型
- 所选 workload 是否支持该参数组合
- 所选 editor mode 是否允许该字段出现

#### 可执行性校验

- 模板是否具备最小执行信息
- 是否存在 run 阶段缺失必填参数的问题
- 是否存在 build / prepare / delete 阶段参数不完整的问题

---

## 8. 页面结构建议

### 8.1 顶层定位

Templates 页面不应只是一个 dropdown，而应是一个“模板管理与编辑工作台”。

### 8.2 建议页面结构

```text
┌──────────────────────────────────────────────────────────────────────┐
│ Templates                                               [New][Import][Export] │
│ Manage reusable benchmark templates for different tools and databases │
├──────────────────────────────────────────────────────────────────────┤
│ Search...   [Tool ▼] [DB ▼] [Category ▼] [Scope ▼]      [Reset]     │
├──────────────────────────────┬───────────────────────────────────────┤
│ Template List                │ Template Editor                       │
│ - Built-in Recommended       │ - Basic Information                   │
│ - Built-in Advanced          │ - Compatibility                       │
│ - User Custom                │ - Phases                             │
│                              │ - Runtime                            │
│                              │ - Tool-specific Settings             │
│                              │ - Preview Summary                    │
│                              │ - Save / Save As                     │
└──────────────────────────────┴───────────────────────────────────────┘
```

### 8.3 右侧详情区建议模块

- Basic Information
- Compatibility
- Phase Configuration
- Runtime Settings
- Tool-specific Settings
- Preview Summary
- Footer Actions

### 8.4 未选中空状态

未选中模板时，应显示：

- `No template selected`
- 说明文字
- `Create Template`
- `Import Template`

---

## 9. 首版设计建议

### 9.1 首版重点

建议 V1 先把以下内容做对：

1. 统一数据模型
2. capability matrix
3. 页面信息架构
4. 模板编辑器的分层结构
5. 内置模板族的分类

### 9.2 各工具的首版策略建议

#### Sysbench

- 先把数据库 OLTP 模板族做完整
- 非数据库 benchmark 不纳入
- 自定义 Lua 数据库脚本先预留数据模型，不急于首版 UI 完整开放

#### Swingbench

- 以 2.7 为能力基线
- 首批重点支持 `OrderEntry / SalesHistory / StressTest`
- 以 Oracle-only 方式进入模板能力矩阵

#### HammerDB

- 模型同时支持 `TPROC-C + TPROC-H`
- 首版 UI 优先把 `TPROC-C` 做深
- `TPROC-H` 先做到结构与字段层面可编辑

---

## 10. 当前阶段结论

基于当前讨论，可以先冻结以下产品定义：

> Templates 是一个“连接无关、阶段化、能力受约束、支持三工具公共模型与专属模型并存”的数据库压测场景模板系统。

只有在这个定义上达成一致，后续的：

- spec
- plan
- tasks
- 页面原型
- 前端实现
- 后端执行映射

才不会反复返工。

---

## 11. 后续建议工作流

建议下一阶段按以下顺序推进：

1. 输出正式 `需求规格（Spec）`
2. 输出 `实现计划（Plan）`
3. 输出 `任务拆解（Tasks）`
4. 再启动 Templates 页面前端实现
5. 最后才接入后端逻辑

---

## 12. 参考资料

- [Sysbench 官方仓库](https://github.com/akopytov/sysbench)
- [Sysbench OLTP 公共脚本](https://github.com/akopytov/sysbench/blob/master/src/lua/oltp_common.lua)
- [Swingbench 官方站点](https://www.dominicgiles.com/swingbench/)
- [Swingbench 文档与示例页](https://www.dominicgiles.com/simplebenchmark.html)
- [HammerDB 官方文档首页](https://hammerdb.com/docs/index.html)
- [HammerDB CLI 文档示例](https://www.hammerdb.com/docs/ch09s03.html)
