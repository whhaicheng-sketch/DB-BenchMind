# DB-BenchMind 自动化回归门禁实施计划

> 文档类型：Plan
> 状态：执行中

## 1. 总体策略

采用分阶段建设策略，但从第一阶段开始就强制执行自动化门禁。原则是“先能阻断，再逐步扩面”，不接受“门禁以后再补”的状态。

## 2. 分阶段目标

### Phase 1：最小可落地强制门禁

目标：

- 建立统一入口
- 固化构建/启动/核心测试为阻断项
- 把规则写入规范文档
- 接入基础 CI

范围：

- `./build`
- `./scripts/entry/build`
- 前端核心 Node 测试
- 后端稳定核心测试集
- 启动 cleanup 契约测试
- Wails 启动 smoke test
- `./scripts/entry/regression`
- `make regression` / `make release-gate`

### Phase 2：补齐失效测试并扩大后端门禁覆盖

目标：

- 修复 `internal/transportwails/bindings` 历史失效测试
- 收敛 `test/integration`
- 恢复 `internal/domain/connection` 到可纳入门禁状态
- 将更多核心包从“推荐”升级为“强制”

### Phase 3：发布前全量回归

目标：

- 建立分层矩阵：开发态 / PR / 发布前
- 引入更完整的任务流 smoke test
- 输出门禁执行报告
- 对关键历史问题建立回归清单

## 3. 当前阶段优先级

- P0：统一入口脚本
- P0：构建门禁
- P0：启动 smoke test
- P0：前端核心回归
- P0：后端稳定核心回归
- P0：README / 开发文档 / 测试文档同步
- P1：CI 工作流
- P1：修复历史失效测试并扩大覆盖
- P2：增加更深的 UI 自动化和发布报告归档

## 4. 风险接受策略

当前阶段可接受：

- 先对“稳定核心子集”强制门禁，而非 `go test ./...` 一步到位
- smoke test 使用受控超时方式验证启动主流程

当前阶段不可接受：

- 无统一门禁入口
- 规则只写在单一文档中
- 变更完成但没有构建/测试证据
- 构建脚本、启动链路变更后不跑自动化验证
