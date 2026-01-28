# DB-BenchMind

**数据库性能压测桌面工作台** - 一键化数据库压测、监控、对比、报告导出

[![Go Version](https://img.shields.io/badge/Go-1.22.2-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

---

## 项目简介

DB-BenchMind 是一款面向数据库工程师与性能测试工程师的桌面压测工作台，通过统一的 GUI 界面编排与运行外部压测工具（Sysbench、Swingbench、HammerDB），对数据库进行性能压测、监控、结果归档与报告导出。

### 核心特性

- **统一入口**：一个应用管理所有数据库和压测工具
- **支持多种数据库**：MySQL、Oracle、SQL Server、PostgreSQL
- **内置场景模板**：7 个常用压测场景（OLTP、TPC-C、TPC-B 等）
- **实时监控**：压测过程中实时显示 TPS、延迟、错误率等关键指标
- **结果对比**：支持 A/B 对比、基线对比、趋势分析
- **多格式报告**：支持 Markdown、HTML、JSON、PDF 四种格式导出
- **可复现性**：每次压测保存完整配置快照，支持精确复跑
- **完整日志记录**：所有操作自动记录到日志文件，便于审计和问题排查

---

## 快速开始

### 环境要求

- **操作系统**: Ubuntu 24（带桌面 GUI）
- **Go 版本**: 1.22.2+
- **压测工具**（需自行安装）:
  - Sysbench >= 1.0
  - Swingbench（最新版）
  - HammerDB（最新版）

### 安装

#### 1. 克隆仓库

```bash
git clone https://github.com/whhaicheng/DB-BenchMind.git
cd DB-BenchMind
```

#### 2. 安装依赖

```bash
make deps
```

#### 3. 构建应用

```bash
make build
```

构建完成后，可执行文件位于 `./build/db-benchmind`

#### 4. 运行应用

```bash
make run
# 或直接运行
./build/db-benchmind
```

---

## 使用指南

### 首次使用流程

1. **启动应用并设置工具路径**
   - 打开"设置"页面
   - 配置 Sysbench、Swingbench、HammerDB 的可执行文件路径
   - 点击"检测版本"确认工具可用

2. **添加数据库连接**
   - 打开"连接管理"页面
   - 点击"新增连接"，选择数据库类型
   - 填写连接信息（主机、端口、用户名、密码）
   - 点击"测试连接"验证配置
   - 保存连接

3. **选择场景模板**
   - 打开"场景模板"页面
   - 查看内置模板（如 Sysbench OLTP Read-Write）
   - 或导入自定义模板（JSON 格式）

4. **配置压测任务**
   - 打开"任务配置"页面
   - 选择数据库连接
   - 选择压测工具和场景模板
   - 配置参数（线程数、时长等）

5. **启动压测**
   - 点击"启动"按钮
   - 切换到"运行监控"页面
   - 实时查看 TPS、延迟、错误率、日志

6. **查看结果与导出报告**
   - 测试完成后，在"历史记录"页面查看运行记录
   - 点击某次运行查看详情
   - 在"报告导出"页面选择格式并导出

---

## 日志管理

### 日志位置

所有操作日志自动记录在 `./data/logs/` 目录：

```
./data/logs/
├── db-benchmind-2026-01-28.log       # GUI 应用日志
└── db-benchmind-cli-2026-01-28.log   # CLI 工具日志
```

### 日志特性

- ✅ **双输出**：同时输出到**控制台**和**日志文件**
- ✅ **按日期归档**：每天自动创建新的日志文件（文件名包含日期）
- ✅ **结构化日志**：使用 `log/slog` 格式，便于解析和查询
- ✅ **完整记录**：记录所有启动、操作、错误信息

### 查看日志

#### 实时监控日志
```bash
# 实时查看 GUI 日志
tail -f ./data/logs/db-benchmind-$(date +%Y-%m-%d).log

# 实时查看 CLI 日志
tail -f ./data/logs/db-benchmind-cli-$(date +%Y-%m-%d).log
```

#### 查看历史日志
```bash
# 查看特定日期的日志
cat ./data/logs/db-benchmind-2026-01-28.log

# 查看最近 20 行
tail -n 20 ./data/logs/db-benchmind-cli-2026-01-28.log

# 搜索错误
grep -i error ./data/logs/*.log

# 搜索特定操作
grep "Adding connection" ./data/logs/*.log
```

#### 列出所有日志文件
```bash
ls -lah ./data/logs/
```

### 日志内容示例

```log
time=2026-01-28T07:40:08.438Z level=INFO msg="Starting DB-BenchMind" log_file=data/logs/db-benchmind-2026-01-28.log
time=2026-01-28T07:40:08.440Z level=INFO msg="Database initialized" path=./data/db-benchmind.db
time=2026-01-28T07:40:08.440Z level=INFO msg="Repositories initialized"
time=2026-01-28T07:40:08.443Z level=INFO msg="Keyring initialized"
time=2026-01-28T07:40:08.443Z level=INFO msg="Use cases initialized"
time=2026-01-28T07:40:08.443Z level=INFO msg="Starting GUI"
```

### 日志维护

建议定期清理旧日志以节省磁盘空间：

```bash
# 清理 7 天前的日志
find ./data/logs -name "*.log" -mtime +7 -delete

# 或者使用 logrotate (推荐)
```

---

## 项目结构

```
DB-BenchMind/
├── cmd/db-benchmind/         # GUI 入口
├── internal/
│   ├── app/usecase/           # 应用层（用例编排）
│   ├── domain/                # 领域层（核心业务逻辑）
│   │   ├── connection/        # 连接管理领域
│   │   ├── template/          # 模板管理领域
│   │   ├── execution/         # 执行编排领域
│   │   └── metric/            # 指标采集领域
│   ├── infra/                 # 基础设施层
│   │   ├── adapter/           # 工具适配器
│   │   ├── database/          # 数据库访问
│   │   ├── keyring/           # 密钥管理
│   │   ├── report/            # 报告生成
│   │   └── chart/             # 图表生成
│   └── transport/ui/          # GUI 页面
├── pkg/benchmark/             # 对外可复用库
├── contracts/                 # 契约定义
│   ├── templates/             # 内置模板
│   ├── schemas/               # Schema 定义
│   └── reports/               # 报告模板
├── configs/                   # 配置样例
├── test/                      # 测试资源
├── docs/                      # 文档
├── specs/                     # 需求与计划文档
└── results/                   # 结果目录
```

---

## 开发指南

### 构建目标

```bash
make build        # 构建应用
make test         # 运行所有测试
make lint         # 运行 linter
make check        # 运行所有检查（格式、测试、lint）
make clean        # 清理构建产物
```

### 代码规范

本项目遵循以下规范：

- **错误处理**: 跨边界必须 wrap 错误并补上下文
- **日志**: 使用 `log/slog` 结构化日志
- **并发**: 保证 `go test -race` 通过
- **测试**: 优先表格驱动测试，TDD 流程
- **提交**: 遵循 Conventional Commits

详见 [CLAUDE.md](./CLAUDE.md) 和 [constitution.md](./constitution.md)

---

## 技术栈

| 技术领域 | 技术选型 | 版本 | 用途 |
|---------|---------|------|------|
| 编程语言 | Go | 1.22.2 | 主要开发语言 |
| GUI 框架 | Fyne | v2.x | 跨平台桌面 GUI |
| 数据库 | SQLite | modernc.org/sqlite | 结果与配置存储 |
| 密钥管理 | go-keyring | latest | 密码安全存储 |
| 日志 | log/slog | 标准库 | 结构化日志 |
| 图表 | fynesimplechart | latest | 实时图表 |

---

## 文档

### 用户文档

- [用户手册 (USER_GUIDE.md)](./docs/USER_GUIDE.md) - 详细的使用指南
- [CLI 快速开始 (QUICK_START.md)](./QUICK_START.md) - CLI 工具快速上手
- [CLI 使用指南 (CLI_USAGE.md)](./CLI_USAGE.md) - CLI 命令参考

### 开发文档

- [API 参考文档 (API_REFERENCE.md)](./docs/API_REFERENCE.md) - 完整的 API 文档
- [开发者指南 (DEVELOPER_GUIDE.md)](./docs/DEVELOPER_GUIDE.md) - 开发环境和规范
- [测试文档 (TESTING.md)](./docs/TESTING.md) - 测试策略和指南
- [日志管理指南 (LOGGING.md)](./docs/LOGGING.md) - 日志系统使用说明

### 项目文档

- [产品需求文档 (spec.md)](./specs/spec.md)
- [技术实现计划 (plan.md)](./specs/plan.md)
- [任务分解 (tasks.md)](./specs/tasks.md)
- [开发规范 (CLAUDE.md)](./CLAUDE.md)
- [项目宪法 (constitution.md)](./constitution.md)
- [架构决策 (architecture.md)](./.specify/steering/architecture.md)

---

## 路线图

### MVP (v1.0) - 12 周

- [x] Phase 1: 项目初始化与基础设施
- [x] Phase 2: 连接管理（4 种数据库）
- [x] Phase 3: 模板系统与任务配置
- [x] Phase 4: 工具适配器与执行编排（3 个工具）
- [x] Phase 5: 结果存储与历史记录
- [x] Phase 6: 报告生成与导出（4 种格式）
- [x] Phase 7: 结果对比功能
- [x] Phase 8: 设置页面与文档完善

### 测试覆盖率

- **总体覆盖率**: ~65%
- **Domain 层**: ~82%
- **Execution 层**: 97%
- **Report 层**: 89%
- **Tool 检测**: 91%

**说明**: 后端核心功能已完成，GUI 部分需要系统库支持（OpenGL/GLFW）。CLI 版本完全可用。

### 未来版本 (v2.0+)

- 分布式压测支持
- 自定义图表编辑器
- Web 界面
- 多语言支持
- 自动参数调优

---

## 许可证

本项目采用 Apache License 2.0 许可证。详见 [LICENSE](./LICENSE) 文件。

---

## 贡献

欢迎贡献！请遵循以下流程：

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feat/AmazingFeature`)
3. 提交更改 (`git commit -m 'feat: add AmazingFeature'`)
4. 推送到分支 (`git push origin feat/AmazingFeature`)
5. 创建 Pull Request

---

## 联系方式

- **作者**: whhaicheng
- **仓库**: https://github.com/whhaicheng/DB-BenchMind
- **问题反馈**: https://github.com/whhaicheng/DB-BenchMind/issues

---

## 致谢

感谢以下开源项目：

- [Fyne](https://fyne.io/) - 跨平台 GUI 框架
- [modernc.org/sqlite](https://gitlab.com/cznic/sqlite) - 纯 Go SQLite 实现
- [go-keyring](https://github.com/zalando/go-keyring) - 密钥管理库
- [Sysbench](https://github.com/akopytov/sysbench) - 数据库压测工具
- [Swingbench](https://www.dominicgiles.com/swingbench/) - Oracle 压测工具
- [HammerDB](https://www.hammerdb.com/) - 多数据库压测工具

---

**DB-BenchMind** - 让数据库性能测试更简单、更高效、更专业。
# DB-BenchMind
# DB-BenchMind
