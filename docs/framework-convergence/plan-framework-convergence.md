# 框架收敛计划：Fyne 到 Wails

## 概述
将项目从双框架（Fyne + Wails）收敛为单一框架（Wails）。

## 阶段 1: 审计阶段

### 1.1 完成审计
- [x] 列出所有 Fyne 相关文件
- [x] 列出所有 Wails 相关文件
- [x] 确认功能覆盖情况
- [x] 确认共享业务层

### 1.2 审计结论
- Wails 已完整覆盖 Fyne 所有功能
- 无需功能迁移
- 可直接删除 Fyne

## 阶段 2: 删除 Fyne 阶段

### 2.1 删除 Fyne UI 入口
- **文件**: `cmd/db-benchmind/main.go`
- **动作**: 删除整个 `cmd/db-benchmind/` 目录
- **原因**: Fyne 主入口，不再需要
- **验证**: 构建后无此入口

### 2.2 删除 Fyne UI 层
- **目录**: `internal/transport/ui/`
- **动作**: 删除整个目录
- **原因**: 所有页面已迁移到 Vue
- **验证**: 无 fyne 引用

### 2.3 清理 Fyne 依赖
- **文件**: `go.mod`, `go.sum`
- **动作**: 删除 fyne.io/fyne/v2 及相关依赖
- **原因**: 不再使用
- **验证**: `go mod tidy` 后无 fyne

## 阶段 3: 统一入口阶段

### 3.1 重命名 Wails 入口
- **当前**: `main_wails.go`
- **目标**: `main.go`
- **动作**: 将 Wails 入口作为唯一主入口
- **验证**: `go build ./cmd/db-benchmind/` 成功

### 3.2 更新构建脚本
- **文件**: `build-app`
- **动作**: 确保只构建 Wails 版本
- **验证**: 脚本执行成功

### 3.3 更新文档
- **文件**: `README.md`
- **动作**: 移除 Fyne 相关说明
- **验证**: 文档只提及 Wails

## 阶段 4: 验证阶段

### 4.1 构建验证
```bash
go build -tags prod -o bin/db-benchmind ./cmd/db-benchmind
```

### 4.2 运行验证
```bash
./bin/db-benchmind
```

### 4.3 依赖验证
```bash
go mod tidy
grep fyne go.mod  # 应无结果
```

## 执行顺序

1. 删除 `cmd/db-benchmind/`
2. 删除 `internal/transport/ui/`
3. 将 `main_wails.go` 内容移到 `cmd/db-benchmind/main.go`
4. 清理 go.mod 中的 Fyne 依赖
5. 更新 build-app 脚本
6. 更新 README.md
7. 构建验证
8. 运行验证

## 风险控制

- 若构建失败，立即回滚
- 共享业务层绝对不删
- 先在临时分支测试
