# DB-BenchMind Wails 多页面导航技术规格

> **文档类型**：技术设计规格
> **版本**：1.0
> **日期**：2026-03-07
> **状态**：待确认
> **关联文档**：`requirements.md`（需求规格）

---

## 目录

1. [技术选型](#技术选型)
2. [架构设计](#架构设计)
3. [目录结构](#目录结构)
4. [前端设计](#前端设计)
5. [后端设计](#后端设计)
6. [数据流设计](#数据流设计)
7. [组件设计](#组件设计)
8. [API 设计](#api-设计)
9. [状态管理设计](#状态管理设计)
10. [样式设计](#样式设计)

---

## 技术选型

### 前端技术栈

| 技术 | 版本 | 用途 | 选择理由 |
|------|------|------|----------|
| Vue 3 | 3.x | 前端框架 | 已在项目中使用 |
| Vue Router | 4.x | 路由管理 | Vue 3 官方路由 |
| Pinia | 2.x | 状态管理 | 已在项目中使用 |
| ECharts | 5.x | 图表库 | 已在项目中使用 |
| CSS Variables | - | 样式主题 | 轻量级主题方案 |

### 不使用的技术（及原因）

| 技术 | 不使用原因 |
|------|------------|
| TypeScript | 项目当前未使用，保持一致性 |
| UI 框架（Element Plus/Naive UI） | 自定义样式更灵活，减少依赖 |
| CSS 预处理器（SCSS/Less） | CSS Variables 足够，减少复杂度 |

---

## 架构设计

### 整体架构图

```
┌─────────────────────────────────────────────────────────────────────┐
│                        Frontend (Vue 3)                              │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────────────┐   │
│  │  Router      │    │  Stores      │    │  Components          │   │
│  │  (Vue Router)│    │  (Pinia)     │    │  (Pages + Shared)    │   │
│  └──────┬───────┘    └──────────────┘    └──────────────────────┘   │
│         │                                                            │
│         ▼                                                            │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │                    Layout Components                          │   │
│  │  ┌────────────┐  ┌────────────────────────────────────────┐  │   │
│  │  │ AppLayout  │  │            MainContent                  │  │   │
│  │  │ ┌────────┐ │  │  ┌────────────────────────────────────┐│  │   │
│  │  │ │Sidebar │ │  │  │                                    ││  │   │
│  │  │ │(Nav)   │ │  │  │  Pages (动态加载)                   ││  │   │
│  │  │ │        │ │  │  │  - ConnectionsPage                 ││  │   │
│  │  │ │        │ │  │  │  - TemplatesPage                   ││  │   │
│  │  │ │        │ │  │  │  - TasksMonitorPage                ││  │   │
│  │  │ │        │ │  │  │  - HistoryPage                     ││  │   │
│  │  │ │        │ │  │  │  - ComparisonPage                  ││  │   │
│  │  │ │        │ │  │  │  - ReportsPage                     ││  │   │
│  │  │ │        │ │  │  │  - SettingsPage                    ││  │   │
│  │  │ └────────┘ │  │  └────────────────────────────────────┘│  │   │
│  │  └────────────┘  └────────────────────────────────────────┘  │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
                                  │
                                  │ Wails Bindings
                                  ▼
┌─────────────────────────────────────────────────────────────────────┐
│                        Backend (Go)                                  │
├─────────────────────────────────────────────────────────────────────┤
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │                    Wails Bindings                             │   │
│  │  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐              │   │
│  │  │Connection   │ │Template     │ │Benchmark    │              │   │
│  │  │Binding      │ │Binding      │ │Binding      │              │   │
│  │  └─────────────┘ └─────────────┘ └─────────────┘              │   │
│  │  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐              │   │
│  │  │History      │ │Comparison   │ │Report       │  (新增)      │   │
│  │  │Binding      │ │Binding      │ │Binding      │              │   │
│  │  └─────────────┘ └─────────────┘ └─────────────┘              │   │
│  │  ┌─────────────┐                                              │   │
│  │  │Settings     │  (新增)                                      │   │
│  │  │Binding      │                                              │   │
│  │  └─────────────┘                                              │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                      │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │                    Use Cases (复用现有)                       │   │
│  │  ConnectionUC | TemplateUC | BenchmarkUC | HistoryUC |       │   │
│  │  ExportUC | ComparisonUC | ReportUC (新增)                   │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

### 页面路由结构

```
/                       → 重定向到 /tasks-monitor
/connections            → ConnectionsPage
/templates              → TemplatesPage
/tasks-monitor          → TasksMonitorPage
/history                → HistoryPage
/comparison             → ComparisonPage
/reports                → ReportsPage
/settings               → SettingsPage
```

---

## 目录结构

### 新增/修改的目录结构

```
frontend/src/
├── main.js                    # 修改：添加 Router
├── App.vue                    # 修改：使用 AppLayout
├── router/                    # 🆕 新增
│   └── index.js               # 路由配置
├── components/
│   ├── layout/                # 🆕 新增
│   │   ├── AppLayout.vue      # 整体布局组件
│   │   ├── Sidebar.vue        # 左侧导航菜单（重写）
│   │   └── PageHeader.vue     # 页面标题组件
│   ├── pages/                 # 🆕 新增：页面组件
│   │   ├── ConnectionsPage.vue
│   │   ├── TemplatesPage.vue
│   │   ├── TasksMonitorPage.vue
│   │   ├── HistoryPage.vue
│   │   ├── ComparisonPage.vue
│   │   ├── ReportsPage.vue
│   │   └── SettingsPage.vue
│   ├── connection/            # 已有：复用
│   │   ├── ConnectionList.vue
│   │   └── ConnectionForm.vue
│   ├── template/              # 已有：复用
│   │   └── TemplateList.vue
│   ├── benchmark/             # 已有：复用
│   │   └── LogPanel.vue
│   ├── charts/                # 已有：复用
│   │   ├── MetricChart.vue
│   │   ├── TpmChart.vue
│   │   ├── TpsChart.vue
│   │   ├── CpuChart.vue
│   │   ├── DiskIOChart.vue
│   │   └── DiskSpaceChart.vue
│   ├── history/               # 🆕 新增
│   │   ├── HistoryList.vue
│   │   └── HistoryDetail.vue
│   ├── comparison/            # 🆕 新增
│   │   ├── RecordSelector.vue
│   │   └── ComparisonResult.vue
│   ├── report/                # 🆕 新增
│   │   ├── ReportConfig.vue
│   │   └── ReportPreview.vue
│   └── settings/              # 🆕 新增
│       └── SettingsForm.vue
├── stores/                    # Pinia stores
│   ├── connection.js          # 已有：复用
│   ├── template.js            # 已有：复用
│   ├── benchmark.js           # 已有：复用
│   ├── monitor.js             # 已有：复用
│   ├── history.js             # 🆕 新增
│   ├── comparison.js          # 🆕 新增
│   └── settings.js            # 🆕 新增
└── styles/
    └── main.css               # 修改：添加导航相关样式

internal/transportwails/bindings/
├── connection.go              # 已有：复用
├── template.go                # 已有：复用
├── benchmark.go               # 已有：复用
├── monitor.go                 # 已有：复用
├── history.go                 # 🆕 新增
├── comparison.go              # 🆕 新增
├── report.go                  # 🆕 新增
└── settings.go                # 🆕 新增
```

---

## 前端设计

### 路由配置

```javascript
// router/index.js
import { createRouter, createWebHashHistory } from 'vue-router'

const routes = [
  {
    path: '/',
    redirect: '/tasks-monitor'
  },
  {
    path: '/connections',
    name: 'Connections',
    component: () => import('../components/pages/ConnectionsPage.vue'),
    meta: { icon: '🔌', title: 'Connections' }
  },
  {
    path: '/templates',
    name: 'Templates',
    component: () => import('../components/pages/TemplatesPage.vue'),
    meta: { icon: '📋', title: 'Templates' }
  },
  {
    path: '/tasks-monitor',
    name: 'TasksMonitor',
    component: () => import('../components/pages/TasksMonitorPage.vue'),
    meta: { icon: '⚡', title: 'Tasks & Monitor' }
  },
  {
    path: '/history',
    name: 'History',
    component: () => import('../components/pages/HistoryPage.vue'),
    meta: { icon: '📜', title: 'History' }
  },
  {
    path: '/comparison',
    name: 'Comparison',
    component: () => import('../components/pages/ComparisonPage.vue'),
    meta: { icon: '📊', title: 'Comparison' }
  },
  {
    path: '/reports',
    name: 'Reports',
    component: () => import('../components/pages/ReportsPage.vue'),
    meta: { icon: '📄', title: 'Reports' }
  },
  {
    path: '/settings',
    name: 'Settings',
    component: () => import('../components/pages/SettingsPage.vue'),
    meta: { icon: '⚙️', title: 'Settings' }
  }
]

const router = createRouter({
  history: createWebHashHistory(),
  routes
})

export default router
```

### App.vue 结构

```vue
<template>
  <AppLayout>
    <router-view />
  </AppLayout>
</template>
```

### AppLayout.vue 设计

```vue
<template>
  <div class="app-layout">
    <!-- 左侧导航 -->
    <Sidebar />

    <!-- 右侧主内容区 -->
    <main class="main-content">
      <slot />
    </main>
  </div>
</template>

<style scoped>
.app-layout {
  display: flex;
  height: 100vh;
  width: 100vw;
  overflow: hidden;
}

.main-content {
  flex: 1;
  overflow: auto;
  background-color: var(--bg-primary);
}
</style>
```

### Sidebar.vue 设计

```vue
<template>
  <nav class="sidebar">
    <!-- Logo -->
    <div class="logo">
      <span class="logo-icon">📊</span>
      <span class="logo-text">DB-BenchMind</span>
    </div>

    <!-- 导航菜单 -->
    <ul class="nav-menu">
      <li
        v-for="route in routes"
        :key="route.path"
        :class="['nav-item', { active: isActive(route) }]"
        @click="navigateTo(route)"
      >
        <span class="nav-icon">{{ route.meta?.icon }}</span>
        <span class="nav-text">{{ route.meta?.title }}</span>
      </li>
    </ul>
  </nav>
</template>
```

---

## 后端设计

### 新增 Bindings

#### HistoryBinding

```go
// internal/transportwails/bindings/history.go

type HistoryBinding struct {
    ctx       context.Context
    historyUC *usecase.HistoryUseCase
    exportUC  *usecase.ExportUseCase
}

func NewHistoryBinding(historyUC *usecase.HistoryUseCase, exportUC *usecase.ExportUseCase) *HistoryBinding

// GetRecords returns all history records
func (b *HistoryBinding) GetRecords() ([]*HistoryRecordDTO, error)

// GetRecordByID returns a single record by ID
func (b *HistoryBinding) GetRecordByID(id string) (*HistoryRecordDTO, error)

// DeleteRecord deletes a record by ID
func (b *HistoryBinding) DeleteRecord(id string) error

// DeleteAllRecords deletes all records
func (b *HistoryBinding) DeleteAllRecords() error

// ExportRecord exports a single record
func (b *HistoryBinding) ExportRecord(id string, format string) (string, error)

// ExportAllRecords exports all records
func (b *HistoryBinding) ExportAllRecords(format string) (int, string, error)
```

#### ComparisonBinding

```go
// internal/transportwails/bindings/comparison.go

type ComparisonBinding struct {
    ctx          context.Context
    comparisonUC *usecase.ComparisonUseCase
}

func NewComparisonBinding(comparisonUC *usecase.ComparisonUseCase) *ComparisonBinding

// GetRecordRefs returns all record references for selection
func (b *ComparisonBinding) GetRecordRefs() ([]*RecordRefDTO, error)

// CompareRecords compares selected records and returns comparison result
func (b *ComparisonBinding) CompareRecords(recordIDs []string) (*ComparisonResultDTO, error)

// ExportComparison exports comparison result
func (b *ComparisonBinding) ExportComparison(result *ComparisonResultDTO, format string) (string, error)
```

#### ReportBinding

```go
// internal/transportwails/bindings/report.go

type ReportBinding struct {
    ctx       context.Context
    reportUC  *usecase.ReportUseCase
}

func NewReportBinding(reportUC *usecase.ReportUseCase) *ReportBinding

// GetAvailableRuns returns list of runs available for report
func (b *ReportBinding) GetAvailableRuns() ([]*RunInfoDTO, error)

// GenerateReport generates a report
func (b *ReportBinding) GenerateReport(req *ReportRequest) (*ReportResult, error)

// PreviewReport previews a report
func (b *ReportBinding) PreviewReport(req *ReportRequest) (string, error)
```

#### SettingsBinding

```go
// internal/transportwails/bindings/settings.go

type SettingsBinding struct {
    ctx       context.Context
    settingsUC *usecase.SettingsUseCase
}

func NewSettingsBinding(settingsUC *usecase.SettingsUseCase) *SettingsBinding

// GetSettings returns current settings
func (b *SettingsBinding) GetSettings() (*SettingsDTO, error)

// SaveSettings saves settings
func (b *SettingsBinding) SaveSettings(settings *SettingsDTO) error

// DetectTools detects installed benchmark tools
func (b *SettingsBinding) DetectTools() (*DetectedToolsDTO, error)

// ResetToDefaults resets settings to defaults
func (b *SettingsBinding) ResetToDefaults() error
```

---

## 数据流设计

### 页面切换流程

```
用户点击导航菜单项
       │
       ▼
Router.push('/target-path')
       │
       ▼
路由守卫检查
       │
       ▼
加载目标页面组件 (懒加载)
       │
       ▼
组件 mounted 钩子执行
       │
       ▼
调用 Pinia Store 加载数据
       │
       ▼
Store 调用 Wails Binding
       │
       ▼
Go Binding 调用 UseCase
       │
       ▼
数据返回，更新 Store 状态
       │
       ▼
Vue 响应式更新页面
```

### CRUD 操作流程

```
用户操作 (创建/编辑/删除)
       │
       ▼
页面组件调用 Store action
       │
       ▼
Store 调用 Wails Binding 方法
       │
       ▼
Go Binding 调用 UseCase
       │
       ▼
UseCase 执行业务逻辑
       │
       ▼
Repository 持久化数据
       │
       ▼
返回结果给前端
       │
       ▼
Store 更新本地状态
       │
       ▼
页面自动刷新
```

---

## 组件设计

### 页面组件接口

每个页面组件应遵循以下接口约定：

```typescript
// 页面组件 Props
interface PageProps {
  // 可选的页面参数
}

// 页面组件 Emits
interface PageEmits {
  // 页面事件
}

// 页面生命周期
interface PageLifecycle {
  // mounted 时加载数据
  onMounted(): void

  // unmounted 时清理资源
  onUnmounted(): void

  // 路由进入时（可选）
  onBeforeRouteEnter?(): void

  // 路由离开时（可选）
  onBeforeRouteLeave?(): void
}
```

### 共享组件

| 组件名 | 用途 | 使用页面 |
|--------|------|----------|
| PageHeader | 页面标题和工具栏 | 所有页面 |
| DataTable | 通用数据表格 | Connections, History, Comparison |
| SearchInput | 搜索输入框 | Connections, Templates, History |
| EmptyState | 空状态占位 | 所有页面 |
| LoadingSpinner | 加载中状态 | 所有页面 |
| Modal | 模态框 | Connections, Reports, Settings |
| Toast | 操作提示 | 所有页面 |

---

## API 设计

### Wails Binding API 列表

#### ConnectionBinding（已有）

| 方法 | 参数 | 返回值 | 说明 |
|------|------|--------|------|
| ListConnections | - | []ConnectionDTO | 获取连接列表 |
| GetConnection | id: string | ConnectionDTO | 获取单个连接 |
| CreateConnection | conn: ConnectionDTO | ConnectionDTO | 创建连接 |
| UpdateConnection | conn: ConnectionDTO | ConnectionDTO | 更新连接 |
| DeleteConnection | id: string | error | 删除连接 |
| TestConnection | id: string | TestResultDTO | 测试连接 |

#### TemplateBinding（已有）

| 方法 | 参数 | 返回值 | 说明 |
|------|------|--------|------|
| ListTemplates | dbType: string | []TemplateDTO | 获取模板列表 |
| GetTemplate | id: string | TemplateDTO | 获取模板详情 |
| GetTemplateParams | id: string | []ParamDTO | 获取模板参数 |

#### BenchmarkBinding（已有）

| 方法 | 参数 | 返回值 | 说明 |
|------|------|--------|------|
| PrepareOnly | req: PrepareRequest | error | 执行 Prepare |
| RunBenchmark | req: RunRequest | error | 执行 Benchmark |
| CleanupOnly | req: CleanupRequest | error | 执行 Cleanup |
| StopBenchmark | force: bool | error | 停止执行 |
| GetStatus | - | BenchmarkStatusDTO | 获取状态 |

#### HistoryBinding（新增）

| 方法 | 参数 | 返回值 | 说明 |
|------|------|--------|------|
| GetRecords | - | []HistoryRecordDTO | 获取历史记录 |
| GetRecordByID | id: string | HistoryRecordDTO | 获取单条记录 |
| DeleteRecord | id: string | error | 删除记录 |
| DeleteAllRecords | - | error | 删除所有记录 |
| ExportRecord | id, format: string | filepath: string | 导出记录 |
| ExportAllRecords | format: string | count, filepath: string | 导出所有记录 |

#### ComparisonBinding（新增）

| 方法 | 参数 | 返回值 | 说明 |
|------|------|--------|------|
| GetRecordRefs | - | []RecordRefDTO | 获取可选记录列表 |
| CompareRecords | ids: []string | ComparisonResultDTO | 比较记录 |
| ExportComparison | result, format | filepath: string | 导出比较结果 |

#### ReportBinding（新增）

| 方法 | 参数 | 返回值 | 说明 |
|------|------|--------|------|
| GetAvailableRuns | - | []RunInfoDTO | 获取可用运行列表 |
| GenerateReport | req: ReportRequest | ReportResult | 生成报告 |
| PreviewReport | req: ReportRequest | content: string | 预览报告 |

#### SettingsBinding（新增）

| 方法 | 参数 | 返回值 | 说明 |
|------|------|--------|------|
| GetSettings | - | SettingsDTO | 获取设置 |
| SaveSettings | settings: SettingsDTO | error | 保存设置 |
| DetectTools | - | DetectedToolsDTO | 检测工具 |
| ResetToDefaults | - | error | 重置设置 |

---

## 状态管理设计

### Pinia Stores

#### history.js（新增）

```javascript
export const useHistoryStore = defineStore('history', {
  state: () => ({
    records: [],
    selectedRecord: null,
    loading: false,
    error: null
  }),

  getters: {
    recordCount: (state) => state.records.length,
    getRecordById: (state) => (id) => state.records.find(r => r.id === id)
  },

  actions: {
    async fetchRecords() { /* ... */ },
    async deleteRecord(id) { /* ... */ },
    async exportRecord(id, format) { /* ... */ }
  }
})
```

#### comparison.js（新增）

```javascript
export const useComparisonStore = defineStore('comparison', {
  state: () => ({
    recordRefs: [],
    selectedIds: [],
    comparisonResult: null,
    loading: false
  }),

  getters: {
    selectedRecords: (state) => state.recordRefs.filter(r => state.selectedIds.includes(r.id))
  },

  actions: {
    async fetchRecordRefs() { /* ... */ },
    async compareSelected() { /* ... */ },
    toggleSelection(id) { /* ... */ }
  }
})
```

#### settings.js（新增）

```javascript
export const useSettingsStore = defineStore('settings', {
  state: () => ({
    settings: {
      sysbenchPath: '',
      swingPath: '',
      hammerPath: '',
      javaPath: '',
      defaultTimeout: 10
    },
    loading: false
  }),

  actions: {
    async fetchSettings() { /* ... */ },
    async saveSettings(settings) { /* ... */ },
    async detectTools() { /* ... */ }
  }
})
```

---

## 样式设计

### CSS 变量定义

```css
/* styles/main.css */

:root {
  /* 导航相关 */
  --nav-width: 200px;
  --nav-bg: #1e2a3a;
  --nav-item-bg: #2a3a4a;
  --nav-item-hover-bg: #3a4a5a;
  --nav-item-active-bg: #4299e1;
  --nav-text-color: #a0aec0;
  --nav-text-active-color: #ffffff;

  /* 主内容区 */
  --content-bg: #1a202c;
  --content-padding: 20px;

  /* 页面相关 */
  --page-header-height: 60px;
  --card-bg: #2a3a4a;
  --card-radius: 8px;

  /* 表格相关 */
  --table-header-bg: #2a3a4a;
  --table-row-hover-bg: #3a4a5a;
  --table-border-color: #3a4a5a;
}
```

### 导航样式

```css
.sidebar {
  width: var(--nav-width);
  min-width: var(--nav-width);
  height: 100vh;
  background-color: var(--nav-bg);
  display: flex;
  flex-direction: column;
  border-right: 1px solid #3a4a5a;
}

.logo {
  height: 60px;
  display: flex;
  align-items: center;
  padding: 0 16px;
  border-bottom: 1px solid #3a4a5a;
}

.nav-menu {
  list-style: none;
  padding: 0;
  margin: 0;
}

.nav-item {
  height: 48px;
  display: flex;
  align-items: center;
  padding: 0 16px;
  cursor: pointer;
  color: var(--nav-text-color);
  transition: background-color 0.2s ease;
}

.nav-item:hover {
  background-color: var(--nav-item-hover-bg);
}

.nav-item.active {
  background-color: var(--nav-item-active-bg);
  color: var(--nav-text-active-color);
}

.nav-icon {
  width: 24px;
  margin-right: 12px;
  font-size: 16px;
}

.nav-text {
  font-size: 14px;
  font-weight: 500;
}
```

---

## 附录

### 文件变更清单

| 类型 | 文件 | 操作 |
|------|------|------|
| 新增 | `frontend/src/router/index.js` | 创建 |
| 新增 | `frontend/src/components/layout/AppLayout.vue` | 创建 |
| 修改 | `frontend/src/components/layout/Sidebar.vue` | 重写 |
| 新增 | `frontend/src/components/layout/PageHeader.vue` | 创建 |
| 新增 | `frontend/src/components/pages/*.vue` | 创建 7 个页面 |
| 新增 | `frontend/src/components/history/*.vue` | 创建 |
| 新增 | `frontend/src/components/comparison/*.vue` | 创建 |
| 新增 | `frontend/src/components/report/*.vue` | 创建 |
| 新增 | `frontend/src/components/settings/*.vue` | 创建 |
| 新增 | `frontend/src/stores/history.js` | 创建 |
| 新增 | `frontend/src/stores/comparison.js` | 创建 |
| 新增 | `frontend/src/stores/settings.js` | 创建 |
| 修改 | `frontend/src/main.js` | 添加 Router |
| 修改 | `frontend/src/App.vue` | 使用 AppLayout |
| 修改 | `frontend/src/styles/main.css` | 添加导航样式 |
| 新增 | `internal/transportwails/bindings/history.go` | 创建 |
| 新增 | `internal/transportwails/bindings/comparison.go` | 创建 |
| 新增 | `internal/transportwails/bindings/report.go` | 创建 |
| 新增 | `internal/transportwails/bindings/settings.go` | 创建 |
| 修改 | `main_wails.go` | 注册新 Bindings |

### 依赖变更

```json
// package.json
{
  "dependencies": {
    "vue-router": "^4.2.0"  // 新增
  }
}
```

### Wails 绑定生成

```bash
# 添加新绑定后运行
wails generate module
```
