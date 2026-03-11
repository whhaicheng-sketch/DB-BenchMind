# 框架收敛规范：Fyne 到 Wails

## 1. 背景与目标

### 问题陈述
当前项目同时存在 Fyne 和 Wails 两套桌面应用框架：
- **Fyne**: `cmd/db-benchmind/main.go` -> `internal/transport/ui/`
- **Wails**: `main_wails.go` -> `frontend/` + `internal/transportwails/`

这导致：
- 结构重复：两套 UI 实现并行
- 维护成本高：同一功能需要维护两份代码
- 实现分裂：新需求不知该走哪个框架
- 后续开发混乱：容易出现双份开发

### 收敛目标
- **目标**: 只保留 Wails 框架，移除所有 Fyne 实现
- **保留**: Wails (Vue 前端 + Go bindings)
- **移除**: Fyne (Go UI 层 + 入口)

### 风险点
1. Fyne 中可能存在 Wails 未覆盖的功能
2. 共享业务层依赖 Fyne UI 类型（如需迁移）
3. 历史数据迁移（无需，SQLite 数据库独立）

### 回滚点
- 若 Wails 功能缺失严重，可回退到 Fyne
- 共享业务层 (internal/app, internal/domain, internal/infra) 完全保留

## 2. 审计结果

### Fyne 相关（需删除）
| 类别 | 路径/文件 | 说明 |
|------|-----------|------|
| 入口 | `cmd/db-benchmind/main.go` | Fyne 主入口 |
| UI 层 | `internal/transport/ui/app.go` | Fyne 应用创建 |
| 页面 | `internal/transport/ui/pages/*.go` | 所有 Fyne 页面实现 |
| 依赖 | `fyne.io/fyne/v2` | go.mod 依赖 |
| 配置 | 无专用配置 | 通过代码硬编码 |

### Wails 相关（需保留）
| 类别 | 路径/文件 | 说明 |
|------|-----------|------|
| 入口 | `main_wails.go` | Wails 主入口 |
| 前端 | `frontend/` | Vue 3 前端 |
| Bindings | `internal/transportwails/` | Wails 绑定层 |
| 配置 | `wails.json` | Wails 配置 |

### 共享层（需保留）
| 类别 | 路径 | 说明 |
|------|------|------|
| Use Cases | `internal/app/usecase/` | 业务逻辑 |
| Domain | `internal/domain/` | 领域模型 |
| Infra | `internal/infra/` | 数据库、适配器等 |
| CLI | `cmd/db-benchmind-cli/` | CLI 入口（独立） |

## 3. 功能覆盖分析

### Fyne 页面 -> Wails 映射
| Fyne 页面 | Wails 实现 | 状态 |
|-----------|-----------|------|
| ConnectionPage | ConnectionsTab + ConnectionForm | ✅ 已实现 |
| TaskPage + TaskMonitorPage | TasksMonitorTab | ✅ 已实现 |
| TemplatePage | TemplatesTab | ✅ 已实现 |
| HistoryPage | HistoryTab | ✅ 已实现 |
| ComparisonPage | ComparisonTab | ✅ 已实现 |
| SettingsPage | SettingsTab | ✅ 已实现 |
| ReportPage | ReportsTab | ✅ 已实现 |
| MonitorPage | 内嵌 Charts | ✅ 已实现 |
| RealtimeChart | Charts 组件 | ✅ 已实现 |

**结论**: Wails 已完整覆盖 Fyne 所有功能，无需功能迁移。

## 4. 实施计划

### Phase 1: 准备
1. 确认 Wails 功能完整覆盖
2. 备份当前代码
3. 更新文档

### Phase 2: 删除 Fyne
1. 删除 `cmd/db-benchmind/` 入口
2. 删除 `internal/transport/ui/` 目录
3. 清理 go.mod 中的 Fyne 依赖
4. 删除 `main_wails.go` 并重命名为 `main.go`

### Phase 3: 统一入口
1. 创建新的 `cmd/db-benchmind/main.go` 指向 Wails
2. 更新 `build-app` 脚本
3. 更新 README 文档

### Phase 4: 验证
1. 构建项目
2. 运行应用
3. 测试关键流程

## 5. 验收标准

- [ ] 项目只保留 Wails 一个框架
- [ ] go.mod 中无 Fyne 依赖
- [ ] 无 `internal/transport/ui/` 目录
- [ ] 入口统一为 Wails
- [ ] 文档统一为 Wails
- [ ] 构建和运行正常
