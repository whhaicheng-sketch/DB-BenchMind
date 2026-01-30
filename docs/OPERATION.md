# 操作规范 (Operation Guidelines)

本文档定义 DB-BenchMind 的操作规范，确保使用一致性和可维护性。

---

## 1. 工作目录规范 (Working Directory)

### 1.1 强制要求

**DB-BenchMind 必须从项目根目录启动！**

```bash
# ✅ 正确：在项目根目录启动
cd /opt/project/DB-Benchmind  # 或你的项目路径
./bin/db-benchmind gui

# ❌ 错误：从其他目录启动
cd /tmp
/opt/project/DB-Benchmind/bin/db-benchmind gui  # 会找不到数据文件！
```

### 1.2 检查当前目录

启动前，确保你在正确的目录：

```bash
# 检查是否存在以下文件/目录
ls -la bin/db-benchmind  # 可执行文件
ls -la data/             # 数据目录（会在首次运行时创建）
ls -la contracts/        # 模板契约目录
ls -la Makefile          # Makefile
```

### 1.3 为什么必须从项目根目录启动？

1. **数据库文件**：`./data/db-benchmind.db`（相对路径）
2. **日志文件**：`./data/logs/db-benchmind-YYYY-MM-DD.log`（相对路径）
3. **模板文件**：`contracts/templates/`（相对路径）
4. **临时文件**：`/tmp/db-benchmind-<uuid>`（绝对路径，sysbench 工作目录）

### 1.4 使用 Makefile（推荐）

```bash
make run
```

Makefile 会自动检查工作目录，如果不在项目根目录会报错并退出。

---

## 2. 启动和停止

### 2.1 启动 GUI

```bash
# 方式1：Makefile（推荐）
make run

# 方式2：直接运行
./bin/db-benchmind gui
```

### 2.2 停止应用

- **GUI 方式**：点击窗口关闭按钮（会自动保存数据）
- **命令行方式**：`Ctrl+C` 或 `kill <pid>`

### 2.3 后台运行（不推荐，仅用于调试）

```bash
# 使用 nohup（仅用于调试，日志会写入 data/logs/）
nohup ./bin/db-benchmind gui > /dev/null 2>&1 &

# 查看日志
tail -f data/logs/db-benchmind-$(date +%Y-%m-%d).log
```

---

## 3. 日志管理

### 3.1 日志位置

所有日志统一写入：`data/logs/db-benchmind-YYYY-MM-DD.log`

### 3.2 查看实时日志

```bash
tail -f data/logs/db-benchmind-$(date +%Y-%m-%d).log
```

### 3.3 日志级别

- `DEBUG`：开发调试信息
- `INFO`：关键业务节点
- `WARN`：可恢复异常/降级
- `ERROR`：不可恢复/影响请求成功

### 3.4 日志轮转

- 每天一个日志文件（按日期分割）
- 旧日志不会自动删除，需手动清理
- 建议定期清理超过 30 天的日志

---

## 4. 数据管理

### 4.1 数据库文件

- **路径**：`data/db-benchmind.db`（SQLite）
- **内容**：连接配置、测试历史、用户设置
- **备份**：定期复制整个 `data/` 目录

```bash
# 备份数据
cp -r data/ data-backup-$(date +%Y%m%d)/
```

### 4.2 测试结果

- **临时文件**：`/tmp/db-benchmind-<uuid>/`（每次测试后自动清理）
- **结果存储**：存储在 `data/db-benchmind.db` 中

### 4.3 清理和重置

```bash
# 停止应用
killall db-benchmind

# 删除数据库（会丢失所有配置和历史！）
rm data/db-benchmind.db

# 删除日志
rm data/logs/*.log

# 重新启动（会创建新的数据库）
./bin/db-benchmind gui
```

---

## 5. 环境变量

### 5.1 支持的环境变量

- `LANG`：语言设置（建议设置 `en_US.UTF-8` 避免 Fyne 警告）
- `HOME`：用户主目录（用于 keyring 存储）

### 5.2 Fyne 语言警告修复

如果看到以下警告：
```
Fyne error: Error parsing user locale C
```

修复方法（已在 `main.go` 中自动处理）：

```bash
export LANG=en_US.UTF-8
./bin/db-benchmind gui
```

---

## 6. 常见问题

### 6.1 "database is locked"

**原因**：多个进程同时访问数据库

**解决**：
```bash
# 查找所有 db-benchmind 进程
ps aux | grep db-benchmind

# 杀死所有进程
killall db-benchmind

# 重新启动
./bin/db-benchmind gui
```

### 6.2 "cannot find data/db-benchmind.db"

**原因**：不在项目根目录启动

**解决**：
```bash
cd /path/to/DB-Benchmind  # 进入项目根目录
./bin/db-benchmind gui     # 重新启动
```

### 6.3 GUI 窗口无法显示

**原因**：显示环境变量未设置

**解决**：
```bash
export DISPLAY=:0
./bin/db-benchmind gui
```

---

## 7. 开发和调试

### 7.1 构建调试版本

```bash
go build -o bin/db-benchmind cmd/db-benchmind/main.go
```

### 7.2 运行测试

```bash
# 单元测试
make test-unit

# 集成测试
make test-integration

# 所有测试
make test
```

### 7.3 代码检查

```bash
# 格式化
make format

# Lint
make lint

# 完整检查
make check
```

---

## 8. 版本升级

### 8.1 备份数据

```bash
# 停止应用
killall db-benchmind

# 备份数据
cp -r data/ data-backup-$(date +%Y%m%d)/
```

### 8.2 更新代码

```bash
git pull origin main
```

### 8.3 重新构建

```bash
make build
```

### 8.4 启动新版本

```bash
./bin/db-benchmind gui
```

---

## 9. 安全注意事项

### 9.1 密码存储

- 数据库密码存储在系统 keyring 中（Linux: secretstorage）
- 密码不会以明文形式写入日志文件
- 环境变量 `MYSQL_PWD` 和 `PGPASSWORD` 仅在进程内部使用

### 9.2 日志敏感信息

- 日志中不会记录完整密码
- 连接信息可能包含主机、端口、用户名（但不包含密码）
- 如需分享日志，请检查是否包含敏感信息

### 9.3 文件权限

```bash
# 建议设置数据目录权限
chmod 700 data/
chmod 600 data/db-benchmind.db
```

---

## 10. 性能优化建议

### 10.1 数据库大小

- 定期清理测试历史（可在 GUI 中操作）
- 超过 10000 条记录会影响性能

### 10.2 日志大小

- 定期清理旧日志文件
- 建议使用 logrotate（需自行配置）

### 10.3 临时文件

- 每次测试的临时文件在测试结束后自动删除
- 如果异常退出，可能残留 `/tmp/db-benchmind-*` 目录
- 可定期清理：`rm -rf /tmp/db-benchmind-*`

---

**版本**：1.0.0
**最后更新**：2026-01-29
**维护者**：DB-BenchMind Team
