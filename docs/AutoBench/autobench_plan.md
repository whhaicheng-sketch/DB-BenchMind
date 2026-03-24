# AutoBench Plan

## 总体策略
按模块执行，不能一次性全部做完。每轮只做一个任务；每轮结束必须更新 progress.md 与 progress.json。

## 模块划分
### M0 规则与进度基础设施
目标：建立 Compatibility Guardrail、progress 规范、任务执行纪律。

### M1 领域模型与后端编排骨架
目标：新增 Suite / SuiteItem / SuiteResult / SuiteReport 领域模型与 usecase 骨架。

### M2 前端页面骨架
目标：新增 AutoBench 页面入口、路由、基础布局，不影响现有页面。

### M3 连接选择与测试范围配置
目标：实现多选 connection、profile 选择、执行计划预览。

### M4 调用现有任务能力的编排链路
目标：通过新增编排层串行调用现有单任务执行能力。

### M5 运行监控与结果聚合
目标：展示 suite 进度、当前 item、分项结果、失败摘要。

### M6 报告输出
目标：生成 HTML + JSON 报告。

### M7 验收与兼容性回归
目标：验证不影响当前功能，完成文档与 progress 收尾。

## 模块顺序
M0 -> M1 -> M2 -> M3 -> M4 -> M5 -> M6 -> M7

## 每轮固定流程
1. 读取 progress.md / progress.json
2. 只执行 next_task
3. 先审计
4. 先补失败测试
5. 再做最小侵入实现
6. 跑测试
7. 回写 progress 文档

## 阶段里程碑
### 里程碑 A
完成 M0-M2：具备新页面与基础架构，但尚未真正执行任务。

### 里程碑 B
完成 M3-M5：可以从新页面编排并监控 suite。

### 里程碑 C
完成 M6-M7：可以输出报告并完成兼容性验收。
