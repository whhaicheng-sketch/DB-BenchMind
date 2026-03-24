# AutoBench Tasks

## 执行纪律
- 每次只执行一个任务
- 不允许跨模块并行推进
- 每次完成后必须回写 progress.md / progress.json
- 任何范围外问题只能记录，不能顺手修

## M0 规则与进度基础设施
### T0.1
建立 Compatibility Guardrail 文档条款，写入 spec 与 plan。
**完成标志**：文档明确“新页面、只调用、不影响现有功能”。

### T0.2
建立 progress.md / progress.json 模板与字段定义。
**完成标志**：文档可支撑中断续跑。

## M1 领域模型与后端编排骨架
### T1.1
定义 Suite / SuiteItem / SuiteResult / SuiteReport 领域模型。

### T1.2
新增 AutoBenchSuiteUseCase 骨架，只保留 create / plan / status 接口占位。

### T1.3
补领域模型与 usecase 骨架测试。

## M2 前端页面骨架
### T2.1
新增 AutoBench 页面入口与路由。

### T2.2
新增 AutoBenchPage 基础布局：Wizard / Monitor / Report 三块占位。

### T2.3
补前端路由与页面基础渲染测试。

## M3 连接选择与测试范围配置
### T3.1
实现 connection 多选与数据库类型过滤展示。

### T3.2
实现 profile 选择：test / cpu_bound / io_bound。

### T3.3
实现执行计划预览。

### T3.4
补对应前端测试。

## M4 调用现有任务能力的编排链路
### T4.1
实现 Suite plan -> SuiteItem 生成逻辑。

### T4.2
实现串行调度器，只调用现有单任务能力。

### T4.3
实现失败策略 continue_by_connection。

### T4.4
补 usecase / runner 测试。

## M5 运行监控与结果聚合
### T5.1
展示 suite 总体进度与当前 item。

### T5.2
展示各 item 状态、阶段、日志入口、摘要结果。

### T5.3
补监控与聚合测试。

## M6 报告输出
### T6.1
定义 SuiteReport 结构与 JSON 输出。

### T6.2
实现 HTML 报告生成。

### T6.3
在页面展示报告摘要与导出入口。

### T6.4
补报告测试。

## M7 验收与兼容性回归
### T7.1
跑现有功能回归，证明未影响当前功能。

### T7.2
跑 AutoBench 模块回归。

### T7.3
更新最终 progress 与收尾文档。

## 固定 progress.json 字段
- current_module
- current_task
- next_task
- done_tasks[]
- blocked_tasks[]
- changed_files[]
- test_results[]
- known_risks[]
