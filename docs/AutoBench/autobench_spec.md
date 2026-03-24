# AutoBench Spec

## 1. 范围声明
AutoBench 是新页面、新模块。它只能调用现有能力，不能影响当前已有功能。

## 2. 功能需求
### FR-1 页面入口
系统应新增 AutoBench 页面入口，不替代现有 Performance Analysis。

### FR-2 连接选择
用户应能多选 connection，并按数据库类型查看候选。

### FR-3 测试范围选择
用户应能选择：
- test
- cpu_bound
- io_bound

### FR-4 执行计划预览
开始前，系统应展示将要执行的 suite items 列表、顺序、预计数量与执行策略。

### FR-5 编排执行
系统应将每个 connection × profile 组合生成为一个 suite item，并按执行策略依次调用现有单任务能力。

### FR-6 失败隔离
单个 suite item 失败不得破坏整个系统状态。默认策略下，应继续后续 connection。

### FR-7 报告输出
执行结束后，系统应生成：
- 用户可读报告（HTML，PDF 可后续补充）
- 机器可读 JSON 结果

### FR-8 结果追踪
系统应能在页面中查看 suite 总体状态、当前 item 状态、各 item 结果。

### FR-9 兼容性保证
新增 AutoBench 后，现有 Connections / Templates / Performance Analysis / History 的原行为不得变化。

## 3. 非功能需求
### NFR-1 向后兼容
新增能力必须向后兼容。

### NFR-2 最小侵入
优先新增页面和编排层，不得为实现 AutoBench 而大改现有链路。

### NFR-3 续跑能力
实施与交付过程中，必须维护 progress.md / progress.json。

### NFR-4 可测试性
每个模块都必须能独立测试，并有明确验收口径。

### NFR-5 可恢复性
中断后可根据进度文档恢复到 next_task。

## 4. 数据对象
### Suite
- id
- name
- status
- selected_connection_ids[]
- selected_profiles[]
- execution_policy
- created_at / started_at / ended_at

### SuiteItem
- id
- suite_id
- connection_id
- template_id
- profile_type
- status
- phase
- linked_task_id
- result_summary
- logs[]
- metrics_summary

### SuiteReport
- suite_id
- summary
- connection_rows[]
- failures[]
- recommendations[]
- artifact_paths

## 5. 状态机
### Suite 状态
- draft
- ready
- running
- partial_success
- success
- failed
- cancelled

### SuiteItem 状态
- pending
- validating
- preparing
- running
- cleaning
- success
- failed
- skipped

## 6. 执行规则
- 默认串行执行
- 默认顺序：test -> cpu_bound -> io_bound
- 默认 cleanup 开启
- 默认失败策略：continue_by_connection

## 7. 验收标准
### AC-1 新页面独立
AutoBench 作为新页面存在，现有 Performance Analysis 不变。

### AC-2 只调用现有能力
AutoBench 仅新增编排与报告逻辑，不重写单任务执行链路。

### AC-3 报告可用
完成后可产出完整 suite 级报告与 JSON 结果。

### AC-4 不回归
现有功能回归测试不受影响。
