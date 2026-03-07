# DB-BenchMind Wails 迁移实施方案

> **文档类型**：实施路径依据
> **版本**：1.1
> **日期**：2026-03-06
> **状态**：已确认

---

## 目录

1. [技术选型理由](#技术选型理由)
2. [技术架构](#技术架构)
3. [目录结构](#目录结构)
4. [前后端职责边界](#前后端职责边界)
5. [指标采集与数据流转](#指标采集与数据流转)
6. [图表渲染思路](#图表渲染思路)
7. [状态管理思路](#状态管理思路)
8. [MVP 范围](#mvp-范围)
9. [实施阶段](#实施阶段)
10. [里程碑](#里程碑)
11. [Backlog 清单](#backlog-清单)
12. [风险与应对](#风险与应对)
13. [依赖项](#依赖项)
14. [验证与回归建议](#验证与回归建议)
15. [开发规范](#开发规范)

---

## 技术选型理由

### 为什么选择 Wails + Vue 3 + ECharts？

| 技术 | 选择理由 | 替代方案及拒绝原因 |
|------|----------|-------------------|
| **Wails** | 1. Go 原生支持，与现有后端无缝集成<br>2. 单一可执行文件部署<br>3. 使用系统 WebView，体积小<br>4. 成熟稳定，社区活跃 | Electron（太重，内存占用高）、Fyne（图表能力差） |
| **Vue 3** | 1. 学习曲线平缓<br>2. Composition API 适合复杂状态管理<br>3. 生态丰富，文档完善 | React（学习曲线较陡）、Svelte（生态较小） |
| **ECharts** | 1. 专业级图表，效果美观<br>2. 支持横向柱状图 + 迷你折线图组合<br>3. 动态数据更新性能好<br>4. 中文文档完善 | Chart.js（功能较弱）、D3（开发成本高） |
| **gopsutil** | 1. 跨平台系统监控<br>2. API 简洁，易于集成<br>3. 广泛使用，稳定可靠 | 自己实现（开发成本高，兼容性差） |

### 为什么当前阶段不做远程采集？

| 原因 | 说明 |
|------|------|
| **复杂度高** | 远程采集需要 SSH/Agent/WMI 等多种协议，涉及凭证管理、连接池、错误重试等 |
| **安全性** | 需要管理远程服务器凭证，增加安全风险 |
| **MVP 原则** | 先保证本机监控稳定可用，验证架构可行性后再扩展 |
| **依赖不确定性** | 不同操作系统、不同环境的远程采集方案差异大 |

**扩展路径**：当前预留扩展接口，后续可通过 `internal/transport-wails/collector/remote/` 实现。

### 为什么采用并行开发而不是直接替换？

| 原因 | 说明 |
|------|------|
| **风险控制** | 避免 Wails 版本开发过程中影响现有功能 |
| **可回退** | Fyne 版本始终可用作备份方案 |
| **渐进式迁移** | 可以逐步验证 Wails 版本的稳定性和性能 |
| **用户选择** | 最终用户可以选择使用哪个版本 |

**切换条件**：Wails 版本通过所有验收标准（AC1-AC11）后，再评估是否切换为默认版本。

---

## 技术架构

### 整体架构图

```
┌─────────────────────────────────────────────────────────────────────┐
│                        DB-BenchMind (Wails)                          │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │                    Frontend (WebView)                         │   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐   │   │
│  │  │  Vue 3 App  │  │  Pinia      │  │  ECharts Charts     │   │   │
│  │  └──────┬──────┘  └─────────────┘  └─────────────────────┘   │   │
│  │         │                                                     │   │
│  │  ┌──────▼──────────────────────────────────────────────────┐ │   │
│  │  │                    Components                             │ │   │
│  │  │  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────────────┐│ │   │
│  │  │  │Sidebar  │ │Charts   │ │LogPanel │ │ConnectionForm   ││ │   │
│  │  │  │(左侧)   │ │(右侧)   │ │         │ │TemplateForm     ││ │   │
│  │  │  └─────────┘ └─────────┘ └─────────┘ └─────────────────┘│ │   │
│  │  └──────────────────────────────────────────────────────────┘ │   │
│  └──────────────────────────────────────────────────────────────────┘   │
│                              │                                        │
│                              │ Wails Bindings (Go <-> JS)            │
│                              ▼                                        │
│  ┌──────────────────────────────────────────────────────────────────┐ │
│  │                    Backend (Go)                                   │ │
│  │  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐   │ │
│  │  │  Wails App      │  │  Benchmark UC   │  │  Collector      │   │ │
│  │  │  (Bindings)     │  │  (复用现有)      │  │  (新增)         │   │ │
│  │  └─────────────────┘  └─────────────────┘  └─────────────────┘   │ │
│  │                              │                                    │ │
│  │  ┌─────────────────────────────────────────────────────────────┐│ │
│  │  │                    Domain Layer (复用现有)                   ││ │
│  │  │  Connection | Template | Execution | History                ││ │
│  │  └─────────────────────────────────────────────────────────────┘│ │
│  │                              │                                    │ │
│  │  ┌─────────────────────────────────────────────────────────────┐│ │
│  │  │                  Infrastructure Layer                        ││ │
│  │  │  Adapters (Sysbench/Swingbench/HammerDB) | Repository      ││ │
│  │  └─────────────────────────────────────────────────────────────┘│ │
│  └──────────────────────────────────────────────────────────────────┘ │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

### Wails 版本与现有 Fyne 版本的关系

```
┌─────────────────────────────────────────────────────────────────────┐
│                        DB-BenchMind 项目                             │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  ┌─────────────────────────────────────────────────────────────┐    │
│  │                    共享业务逻辑层（复用）                     │    │
│  │  internal/app/     - UseCase                                │    │
│  │  internal/domain/  - 领域模型                               │    │
│  │  internal/infra/   - 基础设施（Adapter/Repository）          │    │
│  └─────────────────────────────────────────────────────────────┘    │
│                              │                                       │
│              ┌───────────────┴───────────────┐                      │
│              ▼                               ▼                       │
│  ┌───────────────────────┐      ┌───────────────────────┐          │
│  │  Fyne 版本（保留）     │      │  Wails 版本（新增）   │          │
│  │  cmd/db-benchmind/    │      │  cmd-wails/           │          │
│  │  internal/transport/  │      │  internal/transport-  │          │
│  │    ui/                │      │    wails/             │          │
│  │                       │      │  frontend/            │          │
│  └───────────────────────┘      └───────────────────────┘          │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

---

## 目录结构

### 新增目录

```
DB-BenchMind/
├── cmd/
│   └── db-benchmind/          # 现有 Fyne 入口（保留，不修改）
│
├── cmd-wails/                 # 🆕 Wails 版本入口
│   └── main.go                # Wails 应用入口
│
├── internal/
│   ├── app/                   # 复用：业务用例
│   ├── domain/                # 复用：领域模型
│   ├── infra/                 # 复用：基础设施
│   │   └── adapter/
│   ├── transport/
│   │   └── ui/                # 现有 Fyne UI（保留，不修改）
│   │
│   └── transport-wails/       # 🆕 Wails 传输层
│       ├── app.go             # Wails App 绑定
│       ├── bindings/          # Go-JS 绑定
│       │   ├── connection.go  # 连接管理绑定
│       │   ├── template.go    # 模板管理绑定
│       │   ├── benchmark.go   # Benchmark 执行绑定
│       │   └── monitor.go     # 监控数据绑定
│       └── collector/         # 系统指标采集
│           ├── collector.go   # 统一采集入口
│           ├── cpu.go         # CPU 采集
│           ├── disk_io.go     # Disk IO 采集
│           └── disk_space.go  # Disk 容量采集
│
├── frontend/                  # 🆕 Vue 3 前端
│   ├── src/
│   │   ├── main.js            # Vue 入口
│   │   ├── App.vue            # 根组件
│   │   ├── components/        # 组件
│   │   │   ├── layout/
│   │   │   │   ├── Sidebar.vue        # 左侧控制面板
│   │   │   │   └── MainContent.vue    # 右侧主内容区
│   │   │   ├── connection/
│   │   │   │   ├── ConnectionList.vue
│   │   │   │   └── ConnectionForm.vue
│   │   │   ├── template/
│   │   │   │   ├── TemplateList.vue
│   │   │   │   └── TemplateForm.vue
│   │   │   ├── benchmark/
│   │   │   │   ├── ControlPanel.vue   # Run/Stop 按钮
│   │   │   │   ├── ProgressBar.vue
│   │   │   │   └── LogPanel.vue       # 日志输出
│   │   │   └── charts/
│   │   │       ├── MetricChart.vue    # 通用图表组件
│   │   │       ├── TpmChart.vue       # TPM 图表
│   │   │       ├── TpsChart.vue       # TPS 图表
│   │   │       ├── CpuChart.vue       # CPU 图表
│   │   │       ├── DiskIOChart.vue    # Disk IO 图表
│   │   │       └── DiskSpaceChart.vue # Disk 容量图表
│   │   ├── stores/            # Pinia 状态管理
│   │   │   ├── connection.js
│   │   │   ├── template.js
│   │   │   ├── benchmark.js
│   │   │   └── monitor.js
│   │   ├── services/          # Wails 调用封装
│   │   │   └── wails.js
│   │   └── styles/            # 样式
│   │       └── main.css
│   ├── index.html
│   ├── package.json
│   └── vite.config.js
│
├── wails.json                 # 🆕 Wails 配置
└── build/                     # 🆕 构建输出
    └── wails/
```

### 与现有代码的关系

| 目录 | 处理方式 | 说明 |
|------|----------|------|
| `cmd/db-benchmind/` | 保留，不修改 | Fyne 版本入口 |
| `internal/app/` | 复用，不修改 | 业务用例 |
| `internal/domain/` | 复用，不修改 | 领域模型 |
| `internal/infra/` | 复用，不修改 | 基础设施 |
| `internal/transport/ui/` | 保留，不修改 | Fyne UI 代码 |
| `cmd-wails/` | 新增 | Wails 入口 |
| `internal/transport-wails/` | 新增 | Wails 绑定层 |
| `frontend/` | 新增 | Vue 3 前端 |

---

## 前后端职责边界

### Go 后端职责

| 职责 | 说明 |
|------|------|
| 业务逻辑 | 连接管理、模板管理、Benchmark 执行 |
| 数据持久化 | SQLite 数据库操作 |
| 系统监控 | CPU/Disk IO/Disk Space 采集（使用 gopsutil） |
| Benchmark 适配 | Sysbench/Swingbench/HammerDB 适配器 |
| 事件推送 | 通过 Wails Events 推送实时数据 |

### Vue 前端职责

| 职责 | 说明 |
|------|------|
| UI 渲染 | 左右分栏布局、组件渲染 |
| 用户交互 | 表单输入、按钮点击、状态反馈 |
| 图表渲染 | ECharts 柱状图 + 折线图 |
| 状态管理 | Pinia Store 管理前端状态 |
| 事件监听 | 监听 Wails Events 更新 UI |

### 数据通信

| 方向 | 方式 | 用途 |
|------|------|------|
| 前端 → 后端 | Wails Bindings（方法调用） | CRUD 操作、Benchmark 执行 |
| 后端 → 前端 | Wails Events（事件推送） | 实时数据、进度更新、日志 |

---

## 指标采集与数据流转

### 采集流程

```
┌──────────────────────────────────────────────────────────────────┐
│                         指标采集流程                              │
└──────────────────────────────────────────────────────────────────┘

┌─────────────────┐                    ┌─────────────────┐
│  System Collector│ 1s interval       │ Benchmark Adapter│
│  (CPU/Disk)      │───────────────────▶│ (TPM/TPS)       │
│  gopsutil        │                    │ realtime callback│
└────────┬────────┘                    └────────┬────────┘
         │                                      │
         ▼                                      ▼
┌─────────────────────────────────────────────────────────┐
│                    Ring Buffer (60 points)               │
│  - CPU: [cpu1, cpu2, ..., cpu60]                        │
│  - Disk IO: [io1, io2, ..., io60]                       │
│  - TPM: [tpm1, tpm2, ..., tpm60]                        │
│  - TPS: [tps1, tps2, ..., tps60]                        │
└────────────────────────────┬────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────┐
│                    Stats Calculator                      │
│  - avg, max, min, stddev                                │
│  - CV (Coefficient of Variation)                        │
└────────────────────────────┬────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────┐
│                    Wails Event Emitter                   │
│  EventsEmit("metrics:update", data)                     │
└────────────────────────────┬────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────┐
│                    Vue Frontend (Pinia Store)            │
│  - 监听 "metrics:update" 事件                            │
│  - 更新 Store 状态                                       │
│  - 触发 ECharts 重绘                                     │
└─────────────────────────────────────────────────────────┘
```

### 数据结构

```go
// Go 端数据结构
type MetricsData struct {
    Timestamp   int64   `json:"timestamp"`

    // Benchmark 指标
    Tpm         float64 `json:"tpm"`
    Tps         float64 `json:"tps"`
    TpmAvg      float64 `json:"tpm_avg"`
    TpsAvg      float64 `json:"tps_avg"`
    TpmMax      float64 `json:"tpm_max"`
    TpsMax      float64 `json:"tps_max"`

    // 系统指标
    CpuPercent  float64 `json:"cpu_percent"`
    DiskReadBps float64 `json:"disk_read_bps"`
    DiskWriteBps float64 `json:"disk_write_bps"`
    DiskUsedPercent float64 `json:"disk_used_percent"`

    // 波动分析
    TpmCv       float64 `json:"tpm_cv"`
    TpsCv       float64 `json:"tps_cv"`

    // 状态
    BenchmarkRunning bool `json:"benchmark_running"`
}
```

---

## 图表渲染思路

### 单个图表行结构

```
┌────────────────────────────────────────────────────────────────┐
│  TPM  ████████████████░░░░░░░░  12,450    avg: 10,200  max: 15,000  │
│  ↑     ↑                          ↑          ↑           ↑          │
│ 标签   横向柱状图（当前值）        当前值      平均值       最大值    │
│                                                                    │
│  ┌────────────────────────────────────────────────────────────┐  │
│  │ ▁▂▃▄▅▆▇█▇▆▅▄▃▂▁▂▃▄▅▆▇█▇▆▅▄▃▂▁▂▃▄▅▆▇█▇▆▅▄▃▂▁          │  │
│  │  ↑ 迷你折线图（60秒趋势）                                    │  │
│  └────────────────────────────────────────────────────────────┘  │
└────────────────────────────────────────────────────────────────┘
```

### ECharts 配置要点

| 要素 | 配置 |
|------|------|
| 柱状图 | `type: 'bar'`，横向，宽度 20px |
| 折线图 | `type: 'line'`，smooth，无坐标轴 |
| 数值标签 | 柱状图右侧，formatNumber |
| 颜色 | 动态计算（根据 CV 值） |
| 刷新 | `setOption({ ... }, { notMerge: false })` |

---

## 状态管理思路

### Pinia Store 结构

```javascript
// stores/monitor.js
export const useMonitorStore = defineStore('monitor', {
  state: () => ({
    // 当前值
    currentTpm: null,
    currentTps: null,
    currentCpu: null,
    currentDiskRead: null,
    currentDiskWrite: null,
    currentDiskSpace: null,

    // 统计值
    tpmStats: { avg: 0, max: 0, cv: 0 },
    tpsStats: { avg: 0, max: 0, cv: 0 },

    // 历史数据（用于折线图）
    tpmHistory: [],  // 最近 60 个点
    tpsHistory: [],
    cpuHistory: [],

    // 状态
    benchmarkRunning: false,
  }),

  actions: {
    updateMetrics(data) {
      // 更新状态
    },
    clearBenchmarkData() {
      // Benchmark 结束时清空
    },
  },
})
```

---

## MVP 范围

### MVP 包含（Phase 1-3）

| 功能 | MVP | 说明 |
|------|-----|------|
| Wails 项目骨架 | ✅ | 可启动的空壳 |
| 左右分栏布局 | ✅ | 左侧控制面板，右侧主要展示区 |
| 连接管理 | ✅ | CRUD + 测试连接 |
| 模板管理 | ✅ | 列表 + 选择 + 参数表单 |
| Benchmark 执行 | ✅ | Prepare/Run/Cleanup/Stop |
| 实时日志 | ✅ | 滚动日志 |
| TPM 图表 | ✅ | 柱状图 + 折线图 |
| TPS 图表 | ✅ | 柱状图 + 折线图 |

### MVP 不包含（Phase 4-5）

| 功能 | MVP | 说明 |
|------|-----|------|
| 系统监控图表 | ❌ | CPU/Disk IO/Disk Space |
| 波动分析 | ❌ | CV 计算 |
| 颜色告警 | ❌ | 绿/黄/红 |
| 性能优化 | ❌ | 专项优化 |

---

## 实施阶段

### Phase 1：项目骨架搭建

**目标**：搭建可运行的 Wails + Vue 3 空壳项目
**预估**：2 天
**MVP**：✅

**任务清单**：
- 环境准备（Node.js、Wails CLI）
- 项目初始化（Wails + Vue 3 模板）
- 基础布局（Sidebar + MainContent）
- 基础绑定结构

**验收标准**：
- `wails dev` 可启动应用
- 左侧 300px + 右侧自适应布局正确

---

### Phase 2：核心功能迁移

**目标**：迁移连接管理、模板管理、Benchmark 执行功能
**预估**：5 天
**MVP**：✅

**任务清单**：
- 连接管理（Go 绑定 + Vue 组件）
- 模板管理（Go 绑定 + Vue 组件）
- Benchmark 执行（Go 绑定 + Vue 组件）
- 实时日志

**验收标准**：
- 可管理连接、模板
- 可执行 Benchmark
- 日志实时显示

---

### Phase 3：TPM/TPS 图表

**目标**：实现 TPM/TPS 实时监控图表
**预估**：3 天
**MVP**：✅

**任务清单**：
- 监控服务（Ring Buffer + Stats）
- 图表组件（ECharts 柱状图 + 折线图）
- 实时数据更新（Wails Events）

**验收标准**：
- 图表每秒更新
- Benchmark 未运行时显示 N/A

---

### Phase 4：系统监控图表

**目标**：实现 CPU/Disk IO/Disk Space 监控图表
**预估**：3 天
**MVP**：❌

**任务清单**：
- gopsutil 集成
- CPU/Disk IO/Disk Space 采集器
- 图表组件
- 跨平台测试

**验收标准**：
- 程序启动后立即显示系统指标
- 数值与系统工具一致

---

### Phase 5：优化完善

**目标**：波动分析、颜色告警、性能优化
**预估**：2 天
**MVP**：❌

**任务清单**：
- CV 计算优化
- 颜色告警
- 性能优化
- 构建与测试

**验收标准**：
- CV 计算正确
- 颜色告警生效
- CPU < 5%，内存 < 200MB

---

## 里程碑

```
Timeline
────────────────────────────────────────────────────────────────────▶

Week 1                    Week 2                    Week 3
├─────────────────────────┼─────────────────────────┼───────────────┤
│                         │                         │               │
│  ┌─────────────────┐    │  ┌─────────────────┐    │  ┌─────────┐  │
│  │ Phase 1         │    │  │ Phase 3         │    │  │Phase 5  │  │
│  │ 项目骨架        │    │  │ TPM/TPS 图表    │    │  │优化完善 │  │
│  └────────┬────────┘    │  └────────┬────────┘    │  └────────┘  │
│           │             │           │             │               │
│  ┌────────▼────────┐    │  ┌────────▼────────┐    │               │
│  │ Phase 2         │    │  │ Phase 4         │    │               │
│  │ 核心功能迁移    │    │  │ 系统监控        │    │               │
│  └─────────────────┘    │  └─────────────────┘    │               │
│                         │                         │               │
├─────────────────────────┼─────────────────────────┼───────────────┤
│                         │                         │               │
│  M1: MVP 骨架          │  M2: MVP 完成          │  M3: 完整版   │
│  (Day 2)               │  (Day 10)              │  (Day 15)     │
│                         │                         │               │
└─────────────────────────┴─────────────────────────┴───────────────┘
```

### 里程碑定义

| ID | 里程碑 | 时间点 | MVP | 验收标准 |
|----|--------|--------|-----|----------|
| M1 | MVP 骨架 | Phase 1 结束 | ✅ | Wails 应用可启动，左右分栏布局 |
| M2 | MVP 完成 | Phase 3 结束 | ✅ | 可执行 Benchmark，TPM/TPS 图表可见 |
| M3 | 系统监控可用 | Phase 4 结束 | ❌ | CPU/Disk 图表正常显示 |
| M4 | 完整版本 | Phase 5 结束 | ❌ | 所有功能完成，性能达标 |

---

## Backlog 清单

以下内容不在当前阶段范围内，列入 Backlog：

| ID | 功能 | 说明 | 优先级 | 预计阶段 |
|----|------|------|--------|----------|
| BL1 | 远程数据库服务器监控 | 通过 SSH/Agent 采集远程服务器指标 | 高 | Phase 6 |
| BL2 | 告警阈值配置 | 用户可自定义 CV 阈值和告警规则 | 中 | Phase 6 |
| BL3 | 更丰富的趋势分析 | 支持更长周期的趋势图、同比/环比 | 中 | Phase 7 |
| BL4 | 导出监控快照 | 导出图表截图或数据（CSV/JSON） | 中 | Phase 6 |
| BL5 | Benchmark 与系统指标关联分析 | 自动分析性能瓶颈（如 IO 等待导致 TPS 下降） | 低 | Phase 7 |
| BL6 | 多窗口 / 多实例监控 | 同时监控多个 Benchmark 实例 | 低 | Phase 8 |
| BL7 | 指标持久化与历史回放 | 保存历史监控数据，支持回放 | 低 | Phase 8 |

---

## 风险与应对

| ID | 风险 | 可能性 | 影响 | 应对措施 |
|----|------|--------|------|----------|
| R1 | Wails 版本兼容性问题 | 低 | 高 | 使用稳定版本 v2.x，充分测试 |
| R2 | gopsutil 跨平台差异 | 中 | 中 | 各平台独立测试，预留 platform-specific 代码 |
| R3 | ECharts 高频刷新性能问题 | 中 | 中 | 限制刷新频率 1s，使用数据聚合 |
| R4 | 学习曲线导致开发延期 | 中 | 低 | 预留 buffer，参考官方示例 |
| R5 | 与 Fyne 版本冲突 | 低 | 高 | 独立目录，不修改现有代码 |

---

## 依赖项

### 外部依赖

| 依赖 | 版本 | 用途 | 风险等级 |
|------|------|------|----------|
| Wails | v2.x | 桌面框架 | 低（成熟稳定） |
| Vue | 3.x | 前端框架 | 低（生态丰富） |
| Pinia | 2.x | 状态管理 | 低（官方推荐） |
| Vite | 5.x | 构建工具 | 低（快速高效） |
| ECharts | 5.x | 图表库 | 低（文档完善） |
| gopsutil | v3 | 系统监控 | 低（广泛使用） |
| Node.js | 18+ | 前端构建 | 低 |

### 内部依赖

| 依赖 | 说明 |
|------|------|
| `internal/app/usecase/` | 业务用例，复用 |
| `internal/domain/` | 领域模型，复用 |
| `internal/infra/adapter/` | Benchmark 适配器，复用 |

---

## 验证与回归建议

### 每个 Phase 完成后的验证

1. **单元测试**：新增代码的单元测试
2. **集成测试**：前后端集成测试
3. **手动测试**：按验收标准逐一验证
4. **回归测试**：确保 Fyne 版本仍可正常运行

### MVP 完成后的验证

1. 所有 AC1-AC7 验收标准通过
2. `go test ./...` 通过
3. Fyne 版本 `./start-gui.sh` 正常运行
4. Wails 版本 `wails dev` 正常运行

### 完整版本完成后的验证

1. 所有 AC1-AC11 验收标准通过
2. 三平台（Windows/macOS/Linux）测试通过
3. 性能指标达标（CPU < 5%，内存 < 200MB）

---

## 开发规范

### 前端规范

#### 组件命名
- 使用 PascalCase：`TpmChart.vue`
- 页面组件后缀 `Page`：`MonitorPage.vue`
- 通用组件无后缀：`MetricChart.vue`

#### 状态管理
- 使用 Pinia Composition API 风格
- Store 文件与模块对应：`monitor.js` 管理 Monitor 状态

#### 样式
- 使用 CSS 变量定义主题色
- 避免内联样式
- 使用 Flex/Grid 布局

### 后端规范

#### 绑定层
- 每个模块一个 Binding 文件
- 方法首字母大写（Wails 要求导出）
- 错误使用 `error` 类型返回

#### 事件命名
- 使用 `模块:动作` 格式
- 示例：`metrics:update`、`benchmark:progress`、`log:append`

### Git 规范

#### 分支命名
```
wails/phase-1-skeleton
wails/phase-2-connection
wails/phase-3-charts
```

#### Commit 格式
```
feat(wails): add connection management bindings
fix(wails): correct chart refresh timing
refactor(wails): extract common chart component
```

---

## 附录

### Wails 常用命令

```bash
# 开发模式（热重载）
wails dev

# 构建生产版本
wails build

# 构建特定平台
wails build -platform windows/amd64
wails build -platform darwin/amd64
wails build -platform darwin/arm64
wails build -platform linux/amd64

# 生成前端绑定
wails generate module
```

### 参考资源

- [Wails 官方文档](https://wails.io/docs/introduction)
- [Vue 3 文档](https://vuejs.org/guide/introduction.html)
- [Pinia 文档](https://pinia.vuejs.org/)
- [ECharts 配置项手册](https://echarts.apache.org/zh/option.html)
- [gopsutil 示例](https://github.com/shirou/gopsutil/tree/master/_examples)
