# Templates 模块 Tasks v0.1

## 1. 文档信息

- 模块：Templates
- 文档类型：Tasks
- 版本：v0.1
- 对应 Spec：`Templates_spec_v0.1.md`
- 对应 Plan：`Templates_plan_v0.1.md`

---

## 2. 使用说明

- `[ ]` 未开始
- `[~]` 进行中
- `[x]` 已完成
- `BLOCKED` 阻塞项

---

## 3. 文档与需求

### T01 前置设计文档整理
- [x] 输出前置设计文档
- [x] 明确工具范围：Sysbench / Swingbench 2.7 / HammerDB
- [x] 明确连接无关原则
- [x] 明确阶段化建模原则

### T02 Spec 文档
- [x] 输出 Templates Spec v0.1
- [x] 明确功能范围与非目标
- [x] 明确数据模型
- [x] 明确页面信息架构
- [x] 明确状态流转与验收标准

### T03 Plan 文档
- [x] 输出 Templates Plan v0.1
- [x] 明确实施阶段
- [x] 明确组件拆分计划
- [x] 明确状态管理计划
- [x] 明确 mock 数据计划

### T04 Tasks 文档
- [x] 输出 Tasks 文档
- [x] 建立开发任务列表
- [x] 按模块划分任务与验收项

---

## 4. 调研与梳理

### T05 现有前端结构调研
- [x] 检查 Templates 当前页面文件位置
- [x] 检查 tabs / page 容器组织方式
- [x] 检查 store 管理方式
- [x] 检查样式体系与主题变量

### T06 参考页面调研
- [x] 分析 Connections 页面结构
- [x] 识别可复用的表单样式
- [x] 识别可复用的标签/卡片/空状态组件
- [x] 总结可复用交互模式

### T07 Templates 旧实现梳理
- [x] 识别当前 Templates 现有逻辑
- [x] 标记需保留/复用代码
- [x] 标记需要替换的旧 UI 入口

---

## 5. 数据模型与状态层

### T08 模板类型定义
- [x] 新建或更新 Template 类型定义
- [x] 定义通用字段
- [x] 定义阶段字段
- [x] 定义 runtime 字段
- [x] 定义 tool-specific 字段

### T09 capability matrix
- [x] 建立 tool -> workload -> field 能力矩阵
- [x] 限制 Swingbench 为 Oracle-only
- [x] 限制 Sysbench 为数据库模板
- [x] 区分 HammerDB TPROC-C / TPROC-H 字段

### T10 mock 数据
- [x] 设计首批 mock templates
- [x] 覆盖 built-in / user 两类
- [x] 覆盖三类工具
- [x] 覆盖多数据库与多 workload

### T11 状态管理
- [x] 建立 templates store / composable
- [x] 支持 selectedTemplateId
- [x] 支持 editing draft
- [x] 支持 filters / search
- [x] 支持 create / duplicate / delete / save placeholder

---

## 6. 页面骨架

### T12 Templates 页面主结构
- [x] 将单下拉框页面重构为双栏布局
- [x] 添加页面标题与说明文字
- [x] 添加顶部操作按钮
- [x] 建立左侧列表区与右侧详情区

### T13 筛选工具栏
- [x] 添加搜索框
- [x] 添加 Tool Filter
- [x] 添加 DB Filter
- [x] 添加 Scope Filter
- [x] 添加 Tag Filter
- [x] 添加 Reset Filters

### T14 空状态
- [x] 未选中模板空状态
- [x] 空列表状态
- [x] 搜索无结果状态

---

## 7. 列表区实现

### T15 模板列表
- [x] 展示模板名称
- [x] 展示描述摘要
- [x] 展示 Tool / DB / Scope 标签
- [x] 展示更新时间
- [x] 支持选中高亮

### T16 列表项行为
- [x] 点击选中
- [x] 复制模板入口
- [x] 删除模板入口
- [x] 限制 built-in 不可删除

---

## 8. 详情区实现

### T17 基础信息区
- [x] Template Name
- [x] Description
- [x] Tool
- [x] DB Family
- [x] Workload Family
- [x] Scope / Tags

### T18 Compatibility 区
- [ ] Supported Databases
- [ ] Supported Versions
- [ ] Compatibility Notes
- [ ] Constraint Notes

### T19 Phases 区
- [x] build
- [x] prepare
- [x] generate
- [x] warmup
- [x] run
- [x] verify
- [x] cleanup
- [x] delete

### T20 Runtime 区
- [x] concurrency.mode
- [x] concurrency.value
- [x] durationSeconds
- [x] warmupSeconds
- [x] rampUpSeconds
- [x] iterations / rate / report interval 等高频字段

### T21 Tool-specific 区
- [x] Sysbench 专属字段区
- [x] Swingbench 专属字段区
- [x] HammerDB 专属字段区
- [x] 按 capability matrix 动态显示

### T22 Preview Summary
- [x] 展示 tool / db / workload 摘要
- [x] 展示阶段摘要
- [x] 展示并发与时长摘要
- [x] 展示关键工具参数摘要

### T23 Footer Actions
- [x] Save
- [x] Save As
- [x] Create Task from Template
- [x] Placeholder 提示行为
- [x] 接入真实后端 Save / Update / Delete / Duplicate

---

## 9. 编辑模式与高级能力

### T24 编辑模式切换
- [ ] Standard 模式
- [ ] Advanced 模式
- [ ] Expert 模式占位

### T25 高级参数入口
- [ ] Sysbench extra args 入口
- [ ] Swingbench XML override 入口
- [ ] HammerDB 高级参数入口

---

## 10. 交互与校验

### T26 新建模板
- [x] 创建默认 user template draft
- [x] 自动选中并进入编辑态
- [x] 自动生成初始名称

### T27 保存流程
- [x] dirty 状态标记
- [x] 后端保存逻辑

---

## 11. Test 模板修复任务

### T31 审计真实支持矩阵
- [ ] 审计 Templates 前端筛选逻辑与统计口径
- [ ] 审计 template domain / repository / usecase / binding
- [ ] 审计 task 创建与参数解析链路
- [ ] 确认数据库与 benchmark tool 支持矩阵以代码实现为准
- [ ] 为每个关键结论记录代码定位点

### T32 补齐默认 Test 模板 seed
- [ ] 为每个真实支持数据库补至少 1 个 Test 模板
- [ ] 命名统一为数据库 + 工具 + Test
- [ ] 统一 `scope=test`
- [ ] 统一 `tags` 包含 `test`
- [ ] 统一 `dbFamily` / `database_types` / `tool`

### T33 对齐最小可运行参数
- [ ] Sysbench Test 模板使用最小 `tables` / `table_size` / `threads` / `time`
- [ ] Swingbench Test 模板使用最小 `virtual_users` / `time` / `scale` 相关参数
- [ ] HammerDB Test 模板使用最小 `virtual_users` / `warehouses` / `duration` / `rampup`
- [ ] 参数键必须与任务解析与 adapter 实际消费键一致

### T34 修复筛选与展示一致性
- [ ] 核对前端 `scope` / `tag` / `dbFamily` / `tool` 过滤条件
- [ ] 核对 `visible / total` 统计口径
- [ ] 修复 seed 与前端过滤断裂点
- [ ] 确认仅选择 `Test` 时可见全部 Test 模板
- [ ] 确认 `Database + Tool + Test` 组合筛选可见

### T35 补充自动化测试
- [ ] 补后端 seed / usecase / repository 测试
- [ ] 补 task 参数映射测试
- [ ] 补前端筛选自动化或脚本验证
- [ ] 若前端无单测框架，补页面脚本级验证并记录执行方法

### T36 创建任务与最小链路验收
- [ ] 验证 Templates 页面可选中 Test 模板
- [ ] 验证可从 Templates 进入 Tasks & Monitor
- [ ] 验证可创建任务预览与正式任务
- [ ] 验证 prepare / run / cleanup / full pipeline 按模板能力可执行
- [ ] 验证基础监控 / 结果数据可见

### T37 最终交付约束检查
- [ ] 支持矩阵结论仅引用代码，不引用文档作为事实来源
- [ ] 不能只修显示，必须保证模板可真实创建任务
- [ ] 最终汇报中的关键判断全部附代码定位点
- [x] 保存成功提示

### T28 复制流程
- [x] 复制内置模板为用户模板
- [x] 自动重命名为 Copy
- [x] 复制后自动选中
- [x] 后端 Duplicate API

### T29 删除流程
- [x] 删除确认
- [x] 仅允许 user template 删除
- [x] 删除后选中状态回退
- [x] 后端删除与只读 scope 拒绝

### T30 表单校验
- [x] 通用必填校验
- [x] 按工具动态校验
- [x] 非法组合提示

---

## 11. 样式与体验

### T31 视觉统一
- [x] 对齐 Connections 页面视觉风格
- [x] 统一卡片、输入框、标签、按钮样式
- [x] 修复空白区利用率问题

### T32 细节体验
- [x] 选中态明显
- [x] 分组标题层级清晰
- [x] 表单对齐整齐
- [x] 滚动区处理自然

### T33 响应与稳健性
- [x] 长文本处理
- [x] 多标签换行处理
- [x] 小屏宽度下布局退化策略

---

## 12. 测试与验收

### T34 功能自测
- [x] 列表选择流程
- [x] 新建 / 保存 / 复制 / 删除真实后端流程
- [x] 搜索 / 筛选流程
- [x] 空状态 / 无结果状态
- [x] 前端构建通过

### T35 工具差异验证
- [x] Sysbench 字段显示正确
- [x] Swingbench 字段显示正确
- [x] HammerDB 字段显示正确
- [x] TPROC-C / TPROC-H 区分正确

### T36 验收汇报
- [x] 汇总文档改动
- [x] 汇总页面改动文件
- [x] 对照 spec 验收清单逐项说明
- [x] 说明本轮未实现项与后续建议

---

## 13. 后续阶段预留

### T37 后端接入预留
- [x] 落地 load/create/update/delete 接口
- [x] 预留 import/export 接口边界
- [x] 预留 createTaskFromTemplate 接口边界

### T39 Templates 后端基础能力
- [x] 建立后端 canonical template model
- [x] 复用现有 SQLite repository 持久化模板
- [x] 接入 Wails binding CRUD / Duplicate
- [x] 只读 scope 后端保护
- [x] 后端基础校验

### T40 前后端联调
- [x] Templates 页面初始化改为后端加载
- [x] New / Save / Edit / Delete / Copy 默认走后端
- [x] 保留 mock fallback 作为兜底
- [x] 后端失败提示前端可见

### T38 未来增强项
- [ ] 模板版本管理
- [ ] 模板导入导出
- [ ] 命令预览生成
- [ ] 真正的兼容性检查
- [ ] 模板到任务的完整跳转
