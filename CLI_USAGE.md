# DB-BenchMind CLI - 使用指南

## 📦 可执行程序

程序已编译至：`build/db-benchmind-cli`

## 🚀 快速启动

### 1. 查看版本
```bash
./build/db-benchmind-cli version
# 输出: DB-BenchMind CLI v1.0.0
```

### 2. 查看帮助
```bash
./build/db-benchmind-cli help
```

### 3. 列出数据库连接
```bash
./build/db-benchmind-cli list
```
输出示例：
```
No connections found.

To add a connection, use the database API or CLI:
  mysql - Add MySQL connection
  postgresql - Add PostgreSQL connection
  oracle - Add Oracle connection
  sqlserver - Add SQL Server connection
```

### 4. 检测基准测试工具
```bash
./build/db-benchmind-cli detect
```
输出示例：
```
Detecting benchmark tools...
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
✗ swingbench (not found)

✗ hammerdb (not found)

✓ sysbench
  Path:    /usr/bin/sysbench
  Version: 1.0.20

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Tip: To install tools:
  Sysbench:   apt-get install sysbench
  Swingbench: Download from https://www.swingbench.com
  HammerDB:   Download from https://www.hammerdb.com
```

## 📂 数据存储

程序使用 SQLite 数据存储，数据保存在：
```
./data/db-benchmind.db     # 主数据库
./data/*.key              # 加密密钥（文件降级方案）
```

## 🔧 功能说明

### 当前可用命令

| 命令 | 功能 | 说明 |
|------|------|------|
| `list` | 列出连接 | 显示所有数据库连接 |
| `detect` | 工具检测 | 检测 sysbench/swingbench/hammerdb |
| `version` | 版本信息 | 显示程序版本 |
| `help` | 帮助信息 | 显示使用说明 |

### 支持的数据库类型

- ✅ MySQL
- ✅ PostgreSQL
- ✅ Oracle
- ✅ SQL Server

### 支持的基准测试工具

- ✅ Sysbench (已检测)
- ⚠️ Swingbench (未安装)
- ⚠️ HammerDB (未安装)

## 💡 使用示例

### 示例 1: 检测系统环境
```bash
# 检测已安装的基准测试工具
./build/db-benchmind-cli detect

# 如果 sysbench 未安装：
sudo apt-get install sysbench

# 再次检测
./build/db-benchmind-cli detect
```

### 示例 2: 管理连接
```bash
# 注意：当前版本 CLI 仅支持查看连接
# 添加连接需要通过 API 或 GUI（待实现）

# 查看现有连接
./build/db-benchmind-cli list
```

## 🔮 后续开发

CLI 版本还在开发中，后续将支持：

- `add` - 添加数据库连接
- `test` - 测试数据库连接
- `bench` - 运行基准测试
- `results` - 查看测试结果
- `report` - 生成测试报告
- `compare` - 对比多次运行结果

## 📚 技术栈

- Go 1.22.2
- SQLite (modernc.org/sqlite)
- Clean Architecture + DDD 设计

## 🐛 故障排除

### 问题 1: "Error: Failed to initialize database"
**解决方法**: 确保当前目录有写权限
```bash
chmod +w .
./build/db-benchmind-cli list
```

### 问题 2: "No connections found"
**说明**: 这是正常的，数据库刚初始化，需要先添加连接

### 问题 3: "✗ sysbench (not found)"
**解决方法**: 安装 sysbench
```bash
# Ubuntu/Debian
sudo apt-get install sysbench

# macOS
brew install sysbench
```

## 📖 更多信息

- GitHub: https://github.com/whhaicheng/DB-BenchMind
- 文档: [./README.md](./README.md)
- 架构: [.specify/steering/architecture.md](.specify/steering/architecture.md)
