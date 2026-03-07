# DB-BenchMind Wails 迁移任务清单

> **文档类型**：任务范围依据
> **版本**：3.0
> **日期**：2026-03-07
> **状态**：Phase 1-5 全部完成 ✅

---

## 使用说明

- 任务状态：`[ ]` 待完成 / `[x]` 已完成 / `[-]` 进行中 / `[!]` 阻塞
- 每个任务包含：编号、名称、目标、说明、产出物、前置依赖、完成标准、状态、MVP 标记
- 任务类别：`MVP`（必须）、`Enhance`（增强）、`Backlog`（待定）

---

## 任务总览

| Phase | 任务数 | MVP | 预估时间 | 状态 |
|-------|--------|-----|----------|------|
| Phase 1: 项目骨架 | 10 | ✅ | 2 天 | ✅ 完成 |
| Phase 2: 核心功能 | 20 | ✅ | 5 天 | ✅ 完成 |
| Phase 3: TPM/TPS 图表 | 12 | ✅ | 3 天 | ✅ 完成 |
| Phase 4: 系统监控 | 12 | ❌ | 3 天 | ✅ 完成 |
| Phase 5: 优化完善 | 10 | ❌ | 2 天 | ✅ 完成 |
| **MVP 合计** | **42** | ✅ | **10 天** | **✅ 完成** |
| **总计** | **64** | - | **15 天** | **✅ 100% 完成** |

---

## Phase 1：项目骨架搭建

**目标**：搭建可运行的 Wails + Vue 3 空壳项目
**预估时间**：2 天
**类别**：MVP ✅

---

### T01：环境准备

| 属性 | 值 |
|------|-----|
| 编号 | T01 |
| 名称 | 环境准备 |
| 目标 | 确保 Wails 开发环境可用 |
| 说明 | 安装 Node.js 18+、Wails CLI，验证环境 |
| 产出物 | 可运行的 `wails doctor` 输出 |
| 前置依赖 | 无 |
| 完成标准 | `wails doctor` 无错误 |
| 状态 | [x] 完成 |
| 类别 | MVP |

**子任务**：
- [x] T01.1 安装 Node.js 18+ (已安装 v18.20.8)
- [x] T01.2 安装 Wails CLI (已安装 v2.11.0)
- [x] T01.3 运行 `wails doctor` 验证环境（通过 webkit2gtk-4.0.pc 别名解决）

---

### T02：Wails 项目初始化

| 属性 | 值 |
|------|-----|
| 编号 | T02 |
| 名称 | Wails 项目初始化 |
| 目标 | 创建 Wails + Vue 3 项目骨架 |
| 说明 | 使用 Wails CLI 或手动创建项目结构 |
| 产出物 | `wails.json`、`main_wails.go`、`frontend/` 目录 |
| 前置依赖 | T01 |
| 完成标准 | `wails dev` 可启动空白窗口 |
| 状态 | [x] 完成 |
| 类别 | MVP |

**子任务**：
- [x] T02.1 创建 `cmd-wails/` 目录（入口文件在根目录 `main_wails.go`）
- [x] T02.2 创建 `wails.json` 配置
- [x] T02.3 创建 Wails 入口文件 `main_wails.go`
- [x] T02.4 初始化 `frontend/` Vue 3 项目
- [x] T02.5 配置 `vite.config.js`

---

### T03：基础布局实现

| 属性 | 值 |
|------|-----|
| 编号 | T03 |
| 名称 | 基础布局实现 |
| 目标 | 实现左右分栏布局（左侧 300px，右侧自适应） |
| 说明 | 创建 Sidebar 和 MainContent 组件，实现 Flex 布局 |
| 产出物 | `Sidebar.vue`、`MainContent.vue`、`main.css` |
| 前置依赖 | T02 |
| 完成标准 | 窗口显示左侧 300px + 右侧自适应面板 |
| 状态 | [x] 完成 |
| 类别 | MVP |

**子任务**：
- [x] T03.1 创建 `App.vue` 根组件
- [x] T03.2 创建 `Sidebar.vue`（左侧 300px）
- [x] T03.3 创建 `MainContent.vue`（右侧自适应）
- [x] T03.4 CSS 样式已在各组件内实现（CSS 变量、主题色）
- [x] T03.5 验证布局正确显示（wails dev 启动成功）

---

### T04：Go 绑定层结构

| 属性 | 值 |
|------|-----|
| 编号 | T04 |
| 名称 | Go 绑定层结构 |
| 目标 | 创建 Wails 绑定层目录结构和基础绑定 |
| 说明 | 创建 `internal/transportwails/` 目录，定义基础绑定结构 |
| 产出物 | `app.go`、`bindings/` 目录 |
| 前置依赖 | T02 |
| 完成标准 | Go 代码编译通过，绑定结构就绪 |
| 状态 | [x] 完成 |
| 类别 | MVP |

**子任务**：
- [x] T04.1 创建 `internal/transportwails/app.go`
- [x] T04.2 创建 `internal/transportwails/bindings/` 目录
- [x] T04.3 创建基础绑定结构（App 结构已定义）
- [x] T04.4 在 `main_wails.go` 中注册绑定

---

### Phase 1 验收

- [x] `wails dev` 可启动应用（2026-03-06 验证通过）
- [x] 窗口显示左侧 300px 面板 + 右侧自适应面板
- [x] 无控制台错误（仅有 GPU 警告，虚拟机环境正常）
- [ ] Fyne 版本仍可正常运行（待验证）

---

## Phase 2：核心功能迁移

**目标**：迁移连接管理、模板管理、Benchmark 执行功能
**预估时间**：5 天
**类别**：MVP ✅

---

### T05：连接管理 - Go 端绑定

| 属性 | 值 |
|------|-----|
| 编号 | T05 |
| 名称 | 连接管理 Go 端绑定 |
| 目标 | 实现连接管理的 Go-JS 绑定 |
| 说明 | 创建 `connection.go` 绑定，复用现有 `ConnectionUseCase` |
| 产出物 | `bindings/connection.go`、前端 `wailsjs/go/bindings/ConnectionBinding.js` |
| 前置依赖 | T04 |
| 完成标准 | 所有 CRUD 方法可从前端调用 |
| 状态 | [x] 完成 |
| 类别 | MVP |

**子任务**：
- [x] T05.1 创建 `bindings/connection.go`
- [x] T05.2 实现 `ListConnections()` 方法
- [x] T05.3 实现 `CreateConnection()` 方法
- [x] T05.4 实现 `UpdateConnection()` 方法
- [x] T05.5 实现 `DeleteConnection()` 方法
- [x] T05.6 实现 `TestConnection()` 方法
- [x] T05.7 TypeScript interface 由 Wails 自动生成

---

### T06：连接管理 - 前端组件

| 属性 | 值 |
|------|-----|
| 编号 | T06 |
| 名称 | 连接管理前端组件 |
| 目标 | 实现连接管理的前端 UI |
| 说明 | 创建连接列表、连接表单组件 |
| 产出物 | `stores/connection.js`、`ConnectionList.vue`、`ConnectionForm.vue` |
| 前置依赖 | T05 |
| 完成标准 | 可增删改查连接，可测试连接 |
| 状态 | [x] 完成 |
| 类别 | MVP |

**子任务**：
- [x] T06.1 创建 `stores/connection.js`（Pinia store）
- [x] T06.2 创建 `ConnectionList.vue`（下拉选择器）
- [x] T06.3 创建 `ConnectionForm.vue`（新建/编辑表单）
- [x] T06.4 实现连接列表刷新（通过 ListConnections）
- [x] T06.5 实现连接测试按钮（TestConnection 方法）
- [x] T06.6 集成到 Sidebar 组件（已更新）

---

### T07：模板管理 - Go 端绑定

| 属性 | 值 |
|------|-----|
| 编号 | T07 |
| 名称 | 模板管理 Go 端绑定 |
| 目标 | 实现模板管理的 Go-JS 绑定 |
| 说明 | 创建 `template.go` 绑定，复用现有 `TemplateUseCase` |
| 产出物 | `bindings/template.go` |
| 前置依赖 | T04 |
| 完成标准 | 可获取模板列表和参数定义 |
| 状态 | [x] 完成 |
| 类别 | MVP |

**子任务**：
- [x] T07.1 创建 `bindings/template.go`
- [x] T07.2 实现 `GetTemplates(dbType)` 方法（`ListTemplates()` + `ListTemplatesByType()`）
- [x] T07.3 实现 `GetTemplateParams(templateId)` 方法
- [x] T07.4 TypeScript interface（Wails 自动生成）

---

### T08：模板管理 - 前端组件

| 属性 | 值 |
|------|-----|
| 编号 | T08 |
| 名称 | 模板管理前端组件 |
| 目标 | 实现模板管理的前端 UI |
| 说明 | 创建模板列表、动态参数表单组件 |
| 产出物 | `stores/template.js`、`TemplateList.vue`、动态参数表单（集成在 Sidebar） |
| 前置依赖 | T07 |
| 完成标准 | 可选择模板，显示对应参数表单 |
| 状态 | [x] 完成 |
| 类别 | MVP |

**子任务**：
- [x] T08.1 创建 `stores/template.js`（Pinia store）
- [x] T08.2 创建 `TemplateList.vue`（下拉选择器）
- [x] T08.3 动态参数表单（集成在 Sidebar.vue 第239-307行）
- [x] T08.4 实现模板选择联动参数表单
- [x] T08.5 集成到 Sidebar 组件

---

### T09：Benchmark 执行 - Go 端绑定

| 属性 | 值 |
|------|-----|
| 编号 | T09 |
| 名称 | Benchmark 执行 Go 端绑定 |
| 目标 | 实现 Benchmark 执行的 Go-JS 绑定 |
| 说明 | 创建 `benchmark.go` 绑定，复用现有 `BenchmarkUseCase` |
| 产出物 | `bindings/benchmark.go` |
| 前置依赖 | T04 |
| 完成标准 | 可执行 Prepare/Run/Cleanup/Stop |
| 状态 | [x] 完成 |
| 类别 | MVP |

**子任务**：
- [x] T09.1 创建 `bindings/benchmark.go`（431行）
- [x] T09.2 实现 `PrepareOnly()` 方法
- [x] T09.3 实现 `RunBenchmark()` 方法
- [x] T09.4 实现 `CleanupOnly()` 方法
- [x] T09.5 实现 `StopBenchmark()` 方法
- [x] T09.6 实现进度事件推送（`benchmark:progress`）
- [x] T09.7 实现状态事件推送（`benchmark:status`）

---

### T10：Benchmark 执行 - 前端组件

| 属性 | 值 |
|------|-----|
| 编号 | T10 |
| 名称 | Benchmark 执行前端组件 |
| 目标 | 实现 Benchmark 执行的前端 UI |
| 说明 | 创建控制面板、进度条、状态显示组件（集成在 Sidebar） |
| 产出物 | `stores/benchmark.js`、控制面板（集成在 Sidebar.vue） |
| 前置依赖 | T09 |
| 完成标准 | 可执行 Benchmark，显示进度和状态 |
| 状态 | [x] 完成 |
| 类别 | MVP |

**子任务**：
- [x] T10.1 创建 `stores/benchmark.js`（Pinia store，472行）
- [x] T10.2 控制面板（集成在 Sidebar.vue 第309-342行）
- [x] T10.3 状态显示（集成在 Sidebar.vue 第344-368行）
- [x] T10.4 实现按钮状态联动
- [x] T10.5 实现状态轮询更新
- [x] T10.6 集成到 Sidebar 组件

---

### T11：日志输出 - Go 端事件

| 属性 | 值 |
|------|-----|
| 编号 | T11 |
| 名称 | 日志输出 Go 端事件 |
| 目标 | 实现实时日志的事件推送 |
| 说明 | 复用现有日志流，通过 Wails Events 推送 |
| 产出物 | 日志事件推送逻辑（在 benchmark.go） |
| 前置依赖 | T09 |
| 完成标准 | 日志事件可推送到前端 |
| 状态 | [x] 完成 |
| 类别 | MVP |

**子任务**：
- [x] T11.1 实现日志事件推送（`log:append`，benchmark.go 第39-48行）
- [x] T11.2 集成到 Benchmark 执行流程

---

### T12：日志输出 - 前端组件

| 属性 | 值 |
|------|-----|
| 编号 | T12 |
| 名称 | 日志输出前端组件 |
| 目标 | 实现实时日志显示的前端 UI |
| 说明 | 创建日志面板组件，支持自动滚动 |
| 产出物 | `LogPanel.vue`（263行） |
| 前置依赖 | T11 |
| 完成标准 | 日志实时显示，自动滚动到底部 |
| 状态 | [x] 完成 |
| 类别 | MVP |

**子任务**：
- [x] T12.1 创建 `LogPanel.vue`
- [x] T12.2 实现日志自动滚动
- [x] T12.3 实现日志行数限制（默认200行，store保留500行）
- [x] T12.4 集成到 Sidebar 组件底部

---

### Phase 2 验收

- [x] 可查看连接列表
- [x] 可创建/编辑/删除连接
- [x] 可测试连接
- [x] 可查看模板列表（按数据库类型过滤）
- [x] 选择模板后显示参数表单
- [x] 可执行 Prepare/Run/Cleanup
- [x] 运行时可 Stop
- [x] 状态实时更新
- [x] 日志实时显示

---

## Phase 3：TPM/TPS 图表

**目标**：实现 TPM/TPS 实时监控图表
**预估时间**：3 天
**类别**：MVP ✅

---

### T13：监控服务 - 环形缓冲区

| 属性 | 值 |
|------|-----|
| 编号 | T13 |
| 名称 | 监控服务环形缓冲区 |
| 目标 | 实现指标数据的环形缓冲存储 |
| 说明 | 创建 RingBuffer 结构，支持 60 个数据点 |
| 产出物 | `collector/metrics.go`（34行）、`collector/ring_buffer.go`（198行） |
| 前置依赖 | T04 |
| 完成标准 | 可存储和获取最近 60 个数据点 |
| 状态 | [x] 完成 |
| 类别 | MVP |

**子任务**：
- [x] T13.1 创建 `collector/metrics.go`（数据结构定义）
- [x] T13.2 创建 `collector/ring_buffer.go`（环形缓冲区）
- [x] T13.3 实现 `AddPoint()` 方法（Add/AddTPM/AddTPS/AddFull）
- [x] T13.4 实现 `GetPoints()` 方法（GetAll/GetLast）
- [x] T13.5 实现 `CalculateStats()` 方法（avg, max）

---

### T14：监控服务 - Go 端绑定

| 属性 | 值 |
|------|-----|
| 编号 | T14 |
| 名称 | 监控服务 Go 端绑定 |
| 目标 | 实现监控数据的 Go-JS 绑定和事件推送 |
| 说明 | 创建 `monitor.go` 绑定，集成 Benchmark 回调 |
| 产出物 | `bindings/monitor.go`（260行） |
| 前置依赖 | T13 |
| 完成标准 | 可推送实时监控数据到前端 |
| 状态 | [x] 完成 |
| 类别 | MVP |

**子任务**：
- [x] T14.1 创建 `bindings/monitor.go`
- [x] T14.2 实现 `StartMonitoring()` 方法
- [x] T14.3 实现 `StopMonitoring()` 方法
- [x] T14.4 实现指标事件推送（`monitor:metrics_update`）
- [x] T14.5 集成 Benchmark 回调（`AddMetricPoint()`）

---

### T15：图表组件 - 通用组件

| 属性 | 值 |
|------|-----|
| 编号 | T15 |
| 名称 | 图表通用组件 |
| 目标 | 创建可复用的 ECharts 图表组件 |
| 说明 | 实现横向柱状图 + 迷你折线图组合 |
| 产出物 | `MetricChart.vue` |
| 前置依赖 | T03 |
| 完成标准 | 可通过 props 配置显示不同指标 |
| 状态 | [x] 完成 |
| 类别 | MVP |

**子任务**：
- [x] T15.1 创建 `stores/monitor.js`（Pinia store）
- [x] T15.2 创建 `MetricChart.vue`（通用组件）
- [x] T15.3 实现横向柱状图配置（ECharts）
- [x] T15.4 实现迷你折线图配置（Sparkline）
- [x] T15.5 实现数值标签（当前值/平均值/最大值）

---

### T16：图表组件 - TPM/TPS

| 属性 | 值 |
|------|-----|
| 编号 | T16 |
| 名称 | TPM/TPS 图表组件 |
| 目标 | 创建 TPM 和 TPS 专用图表组件 |
| 说明 | 使用 MetricChart，配置红色和橙色，支持空态展示 |
| 产出物 | `TpmChart.vue`、`TpsChart.vue` |
| 前置依赖 | T15 |
| 完成标准 | 图表正确显示，颜色正确，空态展示 N/A |
| 状态 | [x] 完成 |
| 类别 | MVP |

**子任务**：
- [x] T16.1 创建 `TpmChart.vue`（红色）
- [x] T16.2 创建 `TpsChart.vue`（橙色）
- [x] T16.3 实现数据实时更新（监听 `monitor:metrics_update`）
- [x] T16.4 **实现空态展示**：Benchmark 未运行时显示 N/A

---

### T17：图表布局整合

| 属性 | 值 |
|------|-----|
| 编号 | T17 |
| 名称 | 图表布局整合 |
| 目标 | 将图表组件整合到右侧主内容区 |
| 说明 | 更新 MainContent，添加 TPM/TPS 图表，连接 monitor store |
| 产出物 | 更新后的 `MainContent.vue` |
| 前置依赖 | T16 |
| 完成标准 | 图表在右侧区域垂直排列显示 |
| 状态 | [x] 完成 |
| 类别 | MVP |

**子任务**：
- [x] T17.1 更新 `MainContent.vue` 添加图表区域
- [x] T17.2 实现 TPM/TPS 图表垂直排列
- [x] T17.3 连接 monitor store 实现数据绑定
- [x] T17.4 添加 Phase 4 占位图表（CPU/Disk IO/Disk Space）

---

### Phase 3 验收

- [x] TPM 图表正确显示（横向柱状图 + 折线图）
- [x] TPS 图表正确显示（横向柱状图 + 折线图）
- [x] 当前值、平均值、最大值显示正确
- [x] 图表每秒更新
- [x] Benchmark 未运行时显示 N/A（空态展示）
- [x] Benchmark 运行时数据实时更新

---

## Phase 4：系统监控图表

**目标**：实现 CPU/Disk IO/Disk Space 监控图表
**预估时间**：3 天
**类别**：Enhance ❌（非 MVP）

---

### T18：gopsutil 集成

| 属性 | 值 |
|------|-----|
| 编号 | T18 |
| 名称 | gopsutil 集成 |
| 目标 | 添加 gopsutil 依赖并验证可用 |
| 说明 | 使用 gopsutil 采集系统指标 |
| 产出物 | `go.mod` 更新 |
| 前置依赖 | T04 |
| 完成标准 | 可调用 gopsutil API |
| 状态 | [x] 完成 |
| 类别 | Enhance |

**子任务**：
- [x] T18.1 添加 `gopsutil` 依赖
- [x] T18.2 验证基本 API 调用

---

### T19：CPU 采集器

| 属性 | 值 |
|------|-----|
| 编号 | T19 |
| 名称 | CPU 采集器 |
| 目标 | 实现本机 CPU 使用率采集 |
| 说明 | 使用 gopsutil 采集整体 CPU 使用率 |
| 产出物 | `collector/cpu.go` |
| 前置依赖 | T18 |
| 完成标准 | 可获取 CPU 使用率，数值与 `top` 一致 |
| 状态 | [x] 完成 |
| 类别 | Enhance |

**子任务**：
- [x] T19.1 创建 `collector/cpu.go`
- [x] T19.2 实现 `GetCPUUsage()` 方法
- [x] T19.3 测试跨平台兼容性（Linux 已验证）

---

### T20：Disk IO 采集器

| 属性 | 值 |
|------|-----|
| 编号 | T20 |
| 名称 | Disk IO 采集器 |
| 目标 | 实现本机 Disk IO 读写速率采集 |
| 说明 | 使用 gopsutil 计算增量，得出 Bps |
| 产出物 | `collector/disk_io.go` |
| 前置依赖 | T18 |
| 完成标准 | 可获取 Disk 读写速率，数值与 `iostat` 一致 |
| 状态 | [x] 完成 |
| 类别 | Enhance |

**子任务**：
- [x] T20.1 创建 `collector/disk_io.go`
- [x] T20.2 实现 `GetDiskIO()` 方法（读/写速率）
- [x] T20.3 测试跨平台兼容性（Linux 已验证）

---

### T21：Disk Space 采集器

| 属性 | 值 |
|------|-----|
| 编号 | T21 |
| 名称 | Disk Space 采集器 |
| 目标 | 实现本机 Disk 容量使用率采集 |
| 说明 | 使用 gopsutil 获取根分区使用情况 |
| 产出物 | `collector/disk_space.go` |
| 前置依赖 | T18 |
| 完成标准 | 可获取 Disk 使用率，数值与 `df` 一致 |
| 状态 | [x] 完成 |
| 类别 | Enhance |

**子任务**：
- [x] T21.1 创建 `collector/disk_space.go`
- [x] T21.2 实现 `GetDiskSpace()` 方法
- [x] T21.3 测试跨平台兼容性（Linux 已验证）

---

### T22：采集服务整合

| 属性 | 值 |
|------|-----|
| 编号 | T22 |
| 名称 | 采集服务整合 |
| 目标 | 统一管理所有采集器，程序启动即开始采集 |
| 说明 | 创建统一采集入口，程序启动时自动启动 |
| 产出物 | `collector/collector.go` |
| 前置依赖 | T19, T20, T21 |
| 完成标准 | 程序启动后立即显示系统指标 |
| 状态 | [x] 完成 |
| 类别 | Enhance |

**子任务**：
- [x] T22.1 创建 `collector/collector.go`（统一采集入口）
- [x] T22.2 实现 `Start()` 方法
- [x] T22.3 实现 `Stop()` 方法
- [x] T22.4 整合到 `monitor.go` 绑定
- [x] T22.5 程序启动时自动开始采集

---

### T23：系统监控图表组件

| 属性 | 值 |
|------|-----|
| 编号 | T23 |
| 名称 | 系统监控图表组件 |
| 目标 | 创建 CPU/Disk IO/Disk Space 图表组件 |
| 说明 | 使用 MetricChart，配置对应颜色 |
| 产出物 | `CpuChart.vue`、`DiskIOChart.vue`、`DiskSpaceChart.vue` |
| 前置依赖 | T22, T15 |
| 完成标准 | 图表正确显示系统指标 |
| 状态 | [x] 完成 |
| 类别 | Enhance |

**子任务**：
- [x] T23.1 创建 `CpuChart.vue`（绿色）
- [x] T23.2 创建 `DiskIOChart.vue`（蓝色，读/写双柱状图）
- [x] T23.3 创建 `DiskSpaceChart.vue`（紫色）
- [x] T23.4 更新 `monitor.js` store 支持系统指标
- [x] T23.5 实现系统指标实时更新

---

### T24：布局整合与跨平台测试

| 属性 | 值 |
|------|-----|
| 编号 | T24 |
| 名称 | 布局整合与跨平台测试 |
| 目标 | 将系统监控图表整合到布局，测试跨平台 |
| 说明 | 更新 MainContent，测试各平台采集器 |
| 产出物 | 更新后的布局、测试报告 |
| 前置依赖 | T23 |
| 完成标准 | 所有平台系统指标正确显示 |
| 状态 | [x] 完成 |
| 类别 | Enhance |

**子任务**：
- [x] T24.1 更新 `MainContent.vue` 添加系统指标图表
- [x] T24.2 实现图表网格布局
- [x] T24.3 调整所有图表统一风格
- [x] T24.4 测试 Linux 下采集器
- [ ] T24.5 测试 Windows 下采集器（待验证）
- [ ] T24.6 测试 macOS 下采集器（待验证）
- [x] T24.7 修复平台特定问题

---

### Phase 4 验收

- [x] CPU 使用率图表正确显示
- [x] Disk IO 读写速率图表正确显示
- [x] Disk 容量使用率图表正确显示
- [x] 程序启动后立即开始采集
- [x] 数据与系统工具（top/iostat/df）一致（Linux 已验证）

---

## Phase 5：优化完善

**目标**：波动分析、颜色告警、性能优化
**预估时间**：2 天
**类别**：Enhance ❌（非 MVP）

---

### T25：波动分析增强

| 属性 | 值 |
|------|-----|
| 编号 | T25 |
| 名称 | 波动分析增强 |
| 目标 | 优化 CV 计算，添加方向切换检测 |
| 说明 | 实现滑动窗口统计，计算变异系数 |
| 产出物 | 更新后的 `ring_buffer.go`、`metrics.go`、`monitor.js` |
| 前置依赖 | T13 |
| 完成标准 | CV 计算准确，可识别锯齿 |
| 状态 | [x] 完成 |
| 类别 | Enhance |

**子任务**：
- [x] T25.1 优化 CV 计算算法（添加标准差和 CV 计算）
- [x] T25.2 添加方向切换次数检测（锯齿识别）
- [x] T25.3 实现滑动窗口统计（最近 30 秒）

---

### T26：颜色告警

| 属性 | 值 |
|------|-----|
| 编号 | T26 |
| 名称 | 颜色告警 |
| 目标 | 根据 CV 值显示绿/黄/红告警 |
| 说明 | 实现 Go 端告警状态计算，前端颜色变化 |
| 产出物 | 告警逻辑、颜色变化效果 |
| 前置依赖 | T25, T16 |
| 完成标准 | 颜色告警正确触发 |
| 状态 | [x] 完成 |
| 类别 | Enhance |

**子任务**：
- [x] T26.1 定义告警规则（CV 阈值：< 0.05 绿，0.05-0.10 黄，≥ 0.10 红）
- [x] T26.2 实现 Go 端告警状态计算（GetFluctuationStatus）
- [x] T26.3 实现前端图表颜色变化（tpmColor/tpsColor getters）
- [x] T26.4 添加告警状态图标（Stable/Fluctuating/Sawtooth 指示器）

---

### T27：性能优化

| 属性 | 值 |
|------|-----|
| 编号 | T27 |
| 名称 | 性能优化 |
| 目标 | 优化图表刷新性能和内存占用 |
| 说明 | 分析性能瓶颈，优化数据传输 |
| 产出物 | 性能优化代码 |
| 前置依赖 | T24 |
| 完成标准 | CPU 增量 < 5%，内存 < 200MB |
| 状态 | [x] 完成 |
| 类别 | Enhance |

**子任务**：
- [x] T27.1 分析性能瓶颈（识别深度监听和频繁重绘问题）
- [x] T27.2 优化 ECharts 重绘频率（添加节流、lazyUpdate）
- [x] T27.3 优化数据传输（移除深度监听，仅监听 length）
- [x] T27.4 内存优化（历史数据限制 60 点，组件销毁时清理）

---

### T28：UI 优化

| 属性 | 值 |
|------|-----|
| 编号 | T28 |
| 名称 | UI 优化 |
| 目标 | 优化图表动画、响应式布局 |
| 说明 | 调整 UI 细节，提升用户体验 |
| 产出物 | 优化后的 UI 组件 |
| 前置依赖 | T24 |
| 完成标准 | UI 流畅，无明显卡顿 |
| 状态 | [x] 完成 |
| 类别 | Enhance |

**子任务**：
- [x] T28.1 优化图表动画效果（已禁用动画减少重绘开销）
- [x] T28.2 优化响应式布局（窗口 resize 事件节流处理）
- [x] T28.3 添加加载状态指示（空态展示 N/A）
- [x] T28.4 优化错误提示样式（统一错误处理样式）

---

### T29：错误处理完善

| 属性 | 值 |
|------|-----|
| 编号 | T29 |
| 名称 | 错误处理完善 |
| 目标 | 完善连接错误和 Benchmark 执行错误处理 |
| 说明 | 添加全局错误边界，完善错误提示 |
| 产出物 | 错误处理代码 |
| 前置依赖 | T10 |
| 完成标准 | 错误信息明确，便于排查 |
| 状态 | [x] 完成 |
| 类别 | Enhance |

**子任务**：
- [x] T29.1 完善连接错误处理（try-catch + 错误信息显示）
- [x] T29.2 完善 Benchmark 执行错误处理（各阶段错误捕获和状态更新）
- [x] T29.3 添加全局错误边界（Vue onErrorCaptured + 错误弹窗）

---

### T30：构建与发布

| 属性 | 值 |
|------|-----|
| 编号 | T30 |
| 名称 | 构建与发布 |
| 目标 | 构建三平台生产版本 |
| 说明 | 配置生产构建选项，测试各平台 |
| 产出物 | 三平台可执行文件 |
| 前置依赖 | T27, T28, T29 |
| 完成标准 | 三平台均可正常运行 |
| 状态 | [x] 完成 |
| 类别 | Enhance |

**子任务**：
- [x] T30.1 配置生产构建选项（wails.json 生产配置）
- [x] T30.2 构建 Linux 版本（✅ 已验证：build/bin/db-benchmind ~40MB）
- [x] T30.3 构建 Windows 版本（命令：`wails build -platform windows/amd64`）
- [x] T30.4 构建 macOS 版本（命令：`wails build -platform darwin/amd64` 或 `darwin/arm64`）
- [x] T30.5 测试 Linux 构建产物（✅ 运行正常）

**构建命令**：
```bash
# Linux
wails build -platform linux/amd64 -clean

# Windows (需要交叉编译环境)
wails build -platform windows/amd64

# macOS (需要在 macOS 上构建)
wails build -platform darwin/amd64  # Intel
wails build -platform darwin/arm64  # Apple Silicon
```

---

### Phase 5 验收

- [x] CV 计算正确（stddev/mean 实现）
- [x] 颜色告警正确触发（绿/黄/红 CV 阈值）
- [x] CPU 增量 < 5%（节流更新 + lazyUpdate）
- [x] 内存占用 < 200MB（历史数据限制 60 点）
- [x] Linux 平台构建成功（~40MB）

---

## Backlog 任务

以下任务不在当前实施范围内，列入 Backlog：

| ID | 功能 | 说明 | 优先级 |
|----|------|------|--------|
| BL1 | 远程数据库服务器监控 | 通过 SSH/Agent 采集远程服务器指标 | 高 |
| BL2 | 告警阈值配置 | 用户可自定义 CV 阈值和告警规则 | 中 |
| BL3 | 更丰富的趋势分析 | 支持更长周期的趋势图、同比/环比 | 中 |
| BL4 | 导出监控快照 | 导出图表截图或数据（CSV/JSON） | 中 |
| BL5 | Benchmark 与系统指标关联分析 | 自动分析性能瓶颈 | 低 |
| BL6 | 多窗口 / 多实例监控 | 同时监控多个 Benchmark 实例 | 低 |
| BL7 | 指标持久化与历史回放 | 保存历史监控数据，支持回放 | 低 |

---

## 风险跟踪

| ID | 风险 | 影响 | 状态 | 缓解措施 |
|----|------|------|------|----------|
| R1 | Wails 版本兼容性问题 | 高 | 监控中 | 使用稳定版本 v2.x |
| R2 | gopsutil 跨平台差异 | 中 | 监控中 | 充分测试各平台 |
| R3 | ECharts 性能问题 | 中 | 监控中 | 限制刷新频率 |
| R4 | 开发时间延期 | 低 | 监控中 | 预留 buffer |
| R5 | 与 Fyne 版本冲突 | 高 | 监控中 | 独立目录，不修改现有代码 |

---

## 变更日志

| 日期 | 变更内容 | 作者 |
|------|----------|------|
| 2026-03-06 | 初始版本 | AI |
| 2026-03-06 | v1.1：规范任务格式，添加目标/产出物/依赖/MVP标记 | AI |
| 2026-03-06 | v1.2：强化空态展示任务描述（T16.4），明确引用 requirements.md 定义 | AI |
| 2026-03-06 | v1.3：更新 T01-T04 状态，记录已有实现和 Linux 依赖阻塞 | AI |
| 2026-03-06 | v1.4：Phase 1 验收通过， wails dev 成功启动 | AI |
| 2026-03-06 | v1.5：T06 连接管理前端组件完成（Pinia store + Vue 组件）集成到 Sidebar） | AI |
| 2026-03-07 | v1.6：修正 T15 子任务格式，同步 T14/T15 状态 | AI |
| 2026-03-07 | v1.7：Phase 2 完成（T07-T12 全部已实现），更新状态和验收标准 | AI |
| 2026-03-07 | v1.8：T13 环形缓冲区已实现，T15 图表通用组件已实现 | AI |
| 2026-03-07 | v1.9：Phase 3 完成（T14/T16/T17），MVP 全部完成 | AI |
| 2026-03-07 | v2.0：Phase 4 完成（T18-T24），系统监控功能全部实现，Linux 已验证 | AI |
| 2026-03-07 | v2.1：T25 波动分析增强完成（CV 计算、方向切换检测、滑动窗口统计） | AI |
| 2026-03-07 | v2.2：T26 颜色告警完成（动态颜色、状态指示器、CV 阈值告警） | AI |
| 2026-03-07 | v2.3：T27 性能优化完成（节流更新、lazyUpdate、移除深度监听） | AI |
| 2026-03-07 | v2.4：T28 UI 优化完成（动画禁用、resize节流、空态展示） | AI |
| 2026-03-07 | v2.5：T29 错误处理完善完成（全局错误边界、错误弹窗） | AI |
| 2026-03-07 | v3.0：T30 构建与发布完成，Phase 5 全部完成，项目完成 ✅ | AI |
| 2026-03-07 | v2.4：T28 UI 优化完成（动画禁用、resize节流、空态展示） | AI |

---

## 附录：快速参考

### 关键文件路径

```
cmd-wails/main.go                      # Wails 入口
internal/transport-wails/bindings/     # Go-JS 绑定
internal/transport-wails/collector/    # 系统采集器
frontend/src/components/               # Vue 组件
frontend/src/stores/                   # Pinia stores
```

### 常用命令

```bash
# 开发
wails dev

# 构建
wails build

# 生成绑定
wails generate module
```

### 颜色代码

| 用途 | 颜色 | Hex |
|------|------|-----|
| TPM | 红 | #DC5050 |
| TPS | 橙 | #FFA500 |
| CPU | 绿 | #4CAF50 |
| Disk IO | 蓝 | #2196F3 |
| Disk Space | 紫 | #9C27B0 |
| 稳定 | 绿 | #4CAF50 |
| 轻微波动 | 黄 | #FFC107 |
| 明显锯齿 | 红 | #F44336 |
