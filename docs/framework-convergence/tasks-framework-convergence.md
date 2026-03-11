# 框架收敛任务清单：Fyne 到 Wails

## 任务总览

| 序号 | 任务 | 文件/目录 | 动作 | 验证 |
|------|------|-----------|------|------|
| 1 | 删除 Fyne 入口 | `cmd/db-benchmind/` | 删除 | 无此目录 |
| 2 | 删除 Fyne UI 层 | `internal/transport/ui/` | 删除 | 无此目录 |
| 3 | 清理 Fyne 依赖 | `go.mod`, `go.sum` | 更新 | 无 fyne 引用 |
| 4 | 统一入口 | `main_wails.go` -> `cmd/db-benchmind/main.go` | 迁移 | 构建成功 |
| 5 | 更新构建脚本 | `build-app` | 更新 | 脚本正确 |
| 6 | 更新文档 | `README.md` | 更新 | 无 Fyne 说明 |
| 7 | 构建验证 | - | 验证 | 构建成功 |
| 8 | 运行验证 | - | 验证 | 运行正常 |

## 详细任务

### Task 1: 删除 Fyne 入口

- **路径**: `cmd/db-benchmind/`
- **动作**: 删除整个目录
- **原因**: Fyne 主入口，已被 Wails 替代
- **执行**: `rm -rf cmd/db-benchmind/`
- **验证**: `ls cmd/db-benchmind/` 应报错

### Task 2: 删除 Fyne UI 层

- **路径**: `internal/transport/ui/`
- **动作**: 删除整个目录
- **原因**: 所有页面已迁移到 Vue 前端
- **执行**: `rm -rf internal/transport/ui/`
- **验证**: `ls internal/transport/ui/` 应报错

### Task 3: 清理 Fyne 依赖

- **文件**: `go.mod`, `go.sum`
- **动作**: 删除 fyne 相关依赖
- **依赖列表**:
  - fyne.io/fyne/v2
  - fyne.io/systray
  - github.com/fyne-io/gl-js
  - github.com/fyne-io/glfw-js
  - github.com/fyne-io/image
  - github.com/fyne-io/oksvg
- **执行**: `go mod tidy`
- **验证**: `grep fyne go.mod` 无结果

### Task 4: 统一入口

- **当前**: `main_wails.go` 在项目根目录
- **目标**: `cmd/db-benchmind/main.go`
- **动作**:
  1. 将 `main_wails.go` 内容复制到 `cmd/db-benchmind/main.go`
  2. 删除根目录的 `main_wails.go`
- **注意**: 需要创建 `cmd/db-benchmind/` 目录
- **验证**: `go build -tags prod -o bin/db-benchmind ./cmd/db-benchmind` 成功

### Task 5: 更新构建脚本

- **文件**: `build-app`
- **动作**: 确保只构建 Wails 版本
- **当前内容**: 使用 `wails build`
- **更新**: 无需修改，已使用 wails
- **验证**: 脚本执行成功

### Task 6: 更新文档

- **文件**: `README.md`
- **动作**: 移除 Fyne 相关说明
- **需要移除**:
  - Fyne 框架说明
  - Fyne 启动方式
- **验证**: `grep -i fyne README.md` 无结果

### Task 7: 构建验证

- **命令**: `go build -tags prod -o bin/db-benchmind ./cmd/db-benchmind`
- **预期**: 构建成功
- **输出**: `bin/db-benchmind` 可执行文件

### Task 8: 运行验证

- **命令**: `./bin/db-benchmind`
- **预期**: 应用启动，GUI 正常显示
- **检查**:
  - 无 Fyne 错误
  - Wails 正常加载前端

## 执行顺序

```
Task 1 -> Task 2 -> Task 3 -> Task 4 -> Task 5 -> Task 6 -> Task 7 -> Task 8
```

## 保留清单

以下目录/文件必须保留，不得删除：

- `internal/app/usecase/` - 业务逻辑
- `internal/domain/` - 领域模型
- `internal/infra/` - 基础设施
- `cmd/db-benchmind-cli/` - CLI 入口
- `frontend/` - Vue 前端
- `internal/transportwails/` - Wails bindings
- `wails.json` - Wails 配置
- `Makefile` - 构建配置
