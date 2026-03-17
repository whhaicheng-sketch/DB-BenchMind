# DB-BenchMind 自动化回归门禁任务清单

> 文档类型：Tasks
> 状态：强制执行

## P0：必须完成

- [x] 建立统一发布门禁入口 `scripts/gates/release_gate.sh`
- [x] 建立脚本目录结构门禁 `scripts/gates/test_script_layout.sh`
- [x] 建立构建门禁脚本 `scripts/gates/test_build_gate.sh`
- [x] 建立后端门禁脚本 `scripts/gates/test_backend_gate.sh`
- [x] 建立前端门禁脚本 `scripts/gates/test_frontend_gate.sh`
- [x] 建立启动 smoke test `scripts/gates/test_smoke_gate.sh`
- [x] 修正启动 cleanup 契约测试 `scripts/gates/test_start_cleanup.sh`
- [x] 把根目录 `build/start` 作为唯一受支持入口写入文档
- [x] 在 Makefile 中提供 `make regression` / `make release-gate`
- [x] 在前端 `package.json` 中提供统一测试入口
- [x] 增加前端关键状态测试
- [x] 修正后端启动清理命令与实际脚本路径不一致的问题
- [x] 新增 CI workflow 复用统一门禁脚本
- [x] 在 spec / plan / tasks 中声明“未通过不得发布”

## P1：下一阶段必须推进

- [ ] 修复 `internal/transportwails/bindings` 现有失效测试，纳入正式门禁
- [ ] 修复 `test/integration` 与当前配置规则不一致的问题，纳入正式门禁
- [ ] 修复 `internal/domain/connection` 全量测试/检查失败问题，纳入正式门禁
- [ ] 为关键任务创建/执行主流程增加更深 smoke test
- [ ] 为核心配置保存/读取增加 smoke test

## P2：增强项

- [ ] 为回归门禁输出结构化测试报告
- [ ] 增加发布前专用回归清单归档
- [ ] 增加关键平台矩阵构建验证
- [ ] 增加更完整的前端交互测试
