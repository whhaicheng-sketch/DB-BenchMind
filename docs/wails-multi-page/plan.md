# DB-BenchMind Wails 多页面导航实施计划

> **文档类型**：实施路径依据
> **版本**：1.0
> **日期**：2026-03-07
> **状态**：待确认

---

## 目录

1. [实施原则](#实施原则)
2. [实施阶段](#实施阶段)
3. [里程碑](#里程碑)
4. [开发顺序](#开发顺序)
5. [测试策略](#测试策略)
6. [回滚策略](#回滚策略)
7. [验收标准](#验收标准)

---

## 实施原则

### 核心原则

| 原则 | 说明 |
|------|------|
| **渐进式迁移** | 不破坏现有功能，逐步迁移 |
| **保持 API 兼容** | 后端绑定保持兼容，前端调用不变 |
| **复用现有代码** | 尽量复用现有组件和 Store |
| **小步提交** | 每个 Phase 完成后提交验证 |

### 技术约束

| 约束 | 说明 |
|------|------|
| 不修改 Fyne 版本代码 | Wails 版本独立开发 |
| 复用现有 UseCase | 后端业务逻辑不变 |
| 保持数据兼容 | 使用相同的数据库和配置 |

---

## 实施阶段

### 阶段总览

```
Timeline
────────────────────────────────────────────────────────────────────▶

Day 1                    Day 2                    Day 3
├─────────────────────────┼─────────────────────────┼───────────────┤
│                         │                         │               │
│  ┌─────────────────┐    │  ┌─────────────────┐    │  ┌─────────┐  │
│  │ Phase 1         │    │  │ Phase 3         │    │  │Phase 5  │  │
│  │ 导航框架        │    │  │ Tasks&Monitor   │    │  │History  │  │
│  └────────┬────────┘    │  └────────┬────────┘    │  └────────┘  │
│           │             │           │             │               │
│  ┌────────▼────────┐    │  ┌────────▼────────┐    │  ┌─────────┐  │
│  │ Phase 2         │    │  │ Phase 4         │    │  │Phase 6  │  │
│  │ Connections     │    │  │ (后端绑定)      │    │  │Comparison│ │
│  │ Templates       │    │  │ History/Comp    │    │  │         │  │
│  └─────────────────┘    │  └─────────────────┘    │  └────────┘  │
│                         │                         │               │
├─────────────────────────┼─────────────────────────┼───────────────┤

Day 4                    Day 5                    Day 6
├─────────────────────────┼─────────────────────────┼───────────────┤
│                         │                         │               │
│  ┌─────────────────┐    │  ┌─────────────────┐    │  ┌─────────┐  │
│  │ Phase 7         │    │  │ Phase 8         │    │  │ 验收    │  │
│  │ Reports         │    │  │ Settings        │    │  │ 修复    │  │
│  └─────────────────┘    │  └─────────────────┘    │  └────────┘  │
│                         │                         │               │
├─────────────────────────┼─────────────────────────┼───────────────┤
│                         │                         │               │
│  M1: 核心页面      │  M2: 扩展页面      │  M3: 完整版   │
│  (Day 2)               │  (Day 4)              │  (Day 6)     │
│                         │                         │               │
└─────────────────────────┴─────────────────────────┴───────────────┘
```

### Phase 1：导航框架搭建

**时间**：Day 1（1 天）
**目标**：实现多页面导航框架
**优先级**：P0（核心）

**关键产出**：
- Vue Router 配置
- AppLayout 布局组件
- Sidebar 导航菜单
- 7 个页面占位组件

**验收标准**：
- [ ] 左侧导航正确显示 7 个菜单项
- [ ] 点击菜单项可切换页面
- [ ] 当前页面高亮显示
- [ ] `wails dev` 启动正常

**任务依赖**：
```
MP-01 (添加依赖)
    │
    ▼
MP-02 (创建路由) ──────┬──▶ MP-05 (创建占位组件)
                       │
    ┌──────────────────┘
    ▼
MP-03 (AppLayout)
    │
    ▼
MP-04 (重写 Sidebar)
    │
    ▼
MP-06 (集成到 App.vue)
    │
    ▼
MP-07 (更新样式)
    │
    ▼
MP-08 (验证)
```

---

### Phase 2：Connections + Templates 页面

**时间**：Day 1（0.5 天 + Phase 1 并行）
**目标**：实现 Connections 和 Templates 独立页面
**优先级**：P0（核心）

**关键产出**：
- ConnectionsPage 完整功能
- TemplatesPage 完整功能

**验收标准**：
- [ ] Connections 页面：列表、创建、编辑、删除、测试连接
- [ ] Templates 页面：列表、过滤、详情、参数配置

**复用策略**：
- 复用现有 `ConnectionList.vue` 组件
- 复用现有 `ConnectionForm.vue` 组件
- 复用现有 `TemplateList.vue` 组件
- 复用现有 `connection.js` 和 `template.js` store

---

### Phase 3：Tasks & Monitor 页面优化

**时间**：Day 2（1 天）
**目标**：优化 Tasks & Monitor 页面布局
**优先级**：P0（核心）

**关键产出**：
- 重构后的 TasksMonitorPage
- 保持所有现有功能

**验收标准**：
- [ ] 连接选择功能正常
- [ ] 模板选择和参数配置功能正常
- [ ] Prepare/Run/Stop/Cleanup 功能正常
- [ ] 图表实时更新正常
- [ ] 日志显示正常

**迁移策略**：
- 将现有 Sidebar 中的控制功能迁移到 TasksMonitorPage
- 保持现有的 benchmark.js 和 monitor.js store 不变
- 复用现有图表组件和 LogPanel 组件

---

### Phase 4：后端绑定新增

**时间**：Day 2-3（1 天）
**目标**：新增 History、Comparison、Report、Settings 后端绑定
**优先级**：P1（重要）

**关键产出**：
- HistoryBinding（history.go）
- ComparisonBinding（comparison.go）
- ReportBinding（report.go）
- SettingsBinding（settings.go）

**验收标准**：
- [ ] 所有绑定方法可从前端调用
- [ ] TypeScript 类型定义自动生成

**实现策略**：
- 复用现有 UseCase（HistoryUseCase、ComparisonUseCase 等）
- 保持与 Fyne 版本相同的业务逻辑

---

### Phase 5：History 页面

**时间**：Day 3（1 天）
**目标**：实现 History 历史记录页面
**优先级**：P1（重要）

**关键产出**：
- HistoryPage 完整功能
- history.js store

**验收标准**：
- [ ] 历史记录列表正确显示
- [ ] 可查看记录详情
- [ ] 可删除记录
- [ ] 可导出记录

---

### Phase 6：Comparison 页面

**时间**：Day 3-4（1 天）
**目标**：实现 Comparison 结果比对页面
**优先级**：P1（重要）

**关键产出**：
- ComparisonPage 完整功能
- comparison.js store

**验收标准**：
- [ ] 可多选历史记录
- [ ] 可按类型过滤
- [ ] 可搜索过滤
- [ ] 对比报告正确生成
- [ ] 可导出对比报告

---

### Phase 7：Reports 页面

**时间**：Day 4（0.5 天）
**目标**：实现 Reports 报告生成页面
**优先级**：P2（增强）

**关键产出**：
- ReportsPage 完整功能

**验收标准**：
- [ ] 可选择运行记录
- [ ] 可选择报告格式
- [ ] 可选择包含章节
- [ ] 可生成报告
- [ ] 可预览报告

---

### Phase 8：Settings 页面

**时间**：Day 4-5（0.5 天）
**目标**：实现 Settings 设置页面
**优先级**：P2（增强）

**关键产出**：
- SettingsPage 完整功能
- settings.js store

**验收标准**：
- [ ] 可配置工具路径
- [ ] 可配置默认超时
- [ ] 可检测已安装工具
- [ ] 可保存设置
- [ ] 可重置为默认值

---

## 里程碑

| ID | 里程碑 | 时间点 | 验收标准 |
|----|--------|--------|----------|
| M1 | 核心页面完成 | Day 2 | Phase 1-3 完成，Connections/Templates/TasksMonitor 功能正常 |
| M2 | 扩展页面完成 | Day 4 | Phase 4-6 完成，History/Comparison 功能正常 |
| M3 | 完整版本 | Day 6 | 所有 7 个页面功能完整，验收通过 |

---

## 开发顺序

### 推荐开发顺序

```
1. Phase 1: 导航框架
   ├── MP-01: 添加 Vue Router
   ├── MP-02: 创建路由配置
   ├── MP-05: 创建页面占位组件
   ├── MP-03: 创建 AppLayout
   ├── MP-04: 重写 Sidebar
   ├── MP-06: 集成到 App.vue
   ├── MP-07: 更新样式
   └── MP-08: 验证

2. Phase 2: Connections 页面
   ├── MP-09: 页面布局
   ├── MP-10: 连接列表
   ├── MP-11: 连接表单弹窗
   ├── MP-12: 测试连接
   ├── MP-13: 删除连接
   └── MP-14: 验证

3. Phase 2: Templates 页面
   ├── MP-15: 页面布局
   ├── MP-16: 模板列表
   ├── MP-17: 模板详情
   ├── MP-18: 参数配置
   └── MP-19: 验证

4. Phase 3: Tasks & Monitor 优化
   ├── MP-20: 页面布局
   ├── MP-21: 连接选择迁移
   ├── MP-22: 模板选择迁移
   ├── MP-23: 控制按钮迁移
   ├── MP-24: 图表日志迁移
   └── MP-25: 验证

5. Phase 4: 后端绑定
   ├── MP-26: HistoryBinding
   ├── MP-34: ComparisonBinding
   ├── MP-41: ReportBinding
   └── MP-44: SettingsBinding

6. Phase 5: History 页面
   ├── MP-27: history store
   ├── MP-28: 页面布局
   ├── MP-29: 记录列表
   ├── MP-30: 记录详情
   ├── MP-31: 删除功能
   ├── MP-32: 导出功能
   └── MP-33: 验证

7. Phase 6: Comparison 页面
   ├── MP-35: comparison store
   ├── MP-36: 页面布局
   ├── MP-37: 记录选择
   ├── MP-38: 对比功能
   ├── MP-39: 导出功能
   └── MP-40: 验证

8. Phase 7: Reports 页面
   ├── MP-42: 页面实现
   └── MP-43: 验证

9. Phase 8: Settings 页面
   ├── MP-45: 页面实现
   └── MP-46: 验证
```

---

## 测试策略

### 单元测试

| 测试范围 | 工具 | 频率 |
|----------|------|------|
| 后端绑定方法 | Go test | 每个 Phase |
| 前端 Store | Vitest | 每个 Phase |

### 集成测试

| 测试范围 | 工具 | 频率 |
|----------|------|------|
| 前后端通信 | 手动测试 | 每个 Phase |
| 页面功能 | 手动测试 | 每个 Phase |

### 验收测试

| 测试范围 | 时间点 |
|----------|--------|
| M1 验收 | Day 2 |
| M2 验收 | Day 4 |
| M3 验收 | Day 6 |

---

## 回滚策略

### 代码回滚

每个 Phase 完成后创建 Git 分支：

```
wails-multi-page/phase-1-nav
wails-multi-page/phase-2-conn-tmpl
wails-multi-page/phase-3-tasks-monitor
wails-multi-page/phase-4-bindings
wails-multi-page/phase-5-history
wails-multi-page/phase-6-comparison
wails-multi-page/phase-7-reports
wails-multi-page/phase-8-settings
```

### 功能回滚

如果某个 Phase 出现严重问题：

1. 切换到上一个稳定分支
2. 验证上一个版本功能正常
3. 修复问题后重新合并

---

## 验收标准

### M1 验收（Day 2）

| ID | 验收项 | 通过标准 |
|----|--------|----------|
| M1-01 | 导航框架 | 7 个菜单项正确显示，可切换 |
| M1-02 | Connections 页面 | 增删改查+测试连接功能正常 |
| M1-03 | Templates 页面 | 列表+过滤+详情+参数功能正常 |
| M1-04 | Tasks & Monitor 页面 | 执行+监控功能正常 |

### M2 验收（Day 4）

| ID | 验收项 | 通过标准 |
|----|--------|----------|
| M2-01 | 后端绑定 | 4 个新增绑定可调用 |
| M2-02 | History 页面 | 列表+详情+删除+导出功能正常 |
| M2-03 | Comparison 页面 | 多选+对比+导出功能正常 |

### M3 验收（Day 6）

| ID | 验收项 | 通过标准 |
|----|--------|----------|
| M3-01 | Reports 页面 | 生成+预览功能正常 |
| M3-02 | Settings 页面 | 配置+保存+检测功能正常 |
| M3-03 | 全功能测试 | 所有 7 个页面功能完整 |
| M3-04 | 性能测试 | 页面切换流畅，无卡顿 |
| M3-05 | Fyne 版本 | 不受影响，正常运行 |

---

## 附录

### 关键文件清单

**新增文件**：
```
frontend/src/router/index.js
frontend/src/components/layout/AppLayout.vue
frontend/src/components/pages/ConnectionsPage.vue
frontend/src/components/pages/TemplatesPage.vue
frontend/src/components/pages/TasksMonitorPage.vue
frontend/src/components/pages/HistoryPage.vue
frontend/src/components/pages/ComparisonPage.vue
frontend/src/components/pages/ReportsPage.vue
frontend/src/components/pages/SettingsPage.vue
frontend/src/stores/history.js
frontend/src/stores/comparison.js
frontend/src/stores/settings.js
internal/transportwails/bindings/history.go
internal/transportwails/bindings/comparison.go
internal/transportwails/bindings/report.go
internal/transportwails/bindings/settings.go
```

**修改文件**：
```
frontend/src/main.js
frontend/src/App.vue
frontend/src/components/layout/Sidebar.vue
frontend/src/styles/main.css
main_wails.go
package.json
```

### 常用命令

```bash
# 开发模式
wails dev

# 生成绑定
wails generate module

# 构建
wails build

# 运行 Fyne 版本（验证不影响）
./start-gui.sh
```
