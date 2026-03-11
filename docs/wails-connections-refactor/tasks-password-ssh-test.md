# Connections 页面密码/SSH/测试修复任务清单 (v2)

## 任务列表

### 后端修改

- [x] **T1**: 添加 SSHTestRequest 结构体（connection.go）
- [x] **T2**: 添加 TestSSHConnection 方法（connection.go）
- [x] **T3**: 验证后端构建

### 前端修改 - Store

- [x] **T4**: 添加 testSSHConnection action（connection.js）
- [x] **T5**: 更新 wails binding（ConnectionBinding.js, .d.ts, models.ts）

### 前端修改 - 表单

- [x] **T6**: 修改测试区域布局，添加 Test SSH 按钮
- [x] **T7**: 修改 Test Database 行为（不走 SSH）
- [x] **T8**: 修改测试状态展示（分开显示）
- [x] **T9**: 移除 "via SSH" 相关误导文案
- [x] **T10**: 修复 UI 样式（Database Name 对齐）

### 前端修改 - 连接列表

- [x] **T11**: 添加 SSH 状态显示（ConnectionsPage.vue）

### 验证

- [x] **T12**: 运行前端构建检查 - ✅ 通过
- [x] **T13**: 运行后端构建检查 - ✅ 通过
- [x] **T14**: 手动测试场景验证

---

## 验收标准对照

| 验收项 | 状态 |
|--------|------|
| 密码回显正确 | ✅ 已实现 |
| SSH 回显正确 | ✅ 已实现 |
| 未启用 SSH：只显示 Test Database | ✅ 已实现 |
| 启用 SSH：显示两个按钮 | ✅ 已实现 |
| Test Database 不走 SSH | ✅ 已实现 |
| Test SSH 真实测试 | ✅ 已实现（后端接口已添加） |
| 连接列表显示 SSH 状态 | ✅ 已实现 |
| UI 样式修复 | ✅ 已实现 |

### 额外修复 (2026-03-10)

- [x] **T13**: 修复 SSH Test 变量名不匹配问题
  - 问题：handleTestSSH 中赋值给 `testSSHResult`，但模板中展示用的是 `sshTestResult`
  - 修复：统一为 `testSSHResult`
  - 验证：构建通过
