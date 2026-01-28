# 日志管理指南 (Logging Guide)

本文档详细说明 DB-BenchMind 的日志系统。

---

## 目录

- [日志概述](#日志概述)
- [日志位置](#日志位置)
- [日志格式](#日志格式)
- [日志级别](#日志级别)
- [查看日志](#查看日志)
- [日志分析](#日志分析)
- [日志维护](#日志维护)

---

## 日志概述

DB-BenchMind 使用 Go 标准库 `log/slog` 实现结构化日志，所有操作都会被记录：

- ✅ 应用启动和关闭
- ✅ 数据库初始化和连接
- ✅ 用户操作（添加/删除连接、运行测试等）
- ✅ 错误和异常信息
- ✅ 性能指标（TPS、延迟等）

### 特性

- **双输出**：同时输出到控制台和日志文件
- **按日期归档**：每天自动创建新的日志文件
- **结构化格式**：易于解析和查询
- **高性能**：异步写入，不影响应用性能

---

## 日志位置

### 目录结构

```
./data/logs/
├── db-benchmind-2026-01-28.log       # GUI 应用日志
├── db-benchmind-2026-01-29.log       # 下一天的日志
├── db-benchmind-cli-2026-01-28.log   # CLI 工具日志
└── db-benchmind-cli-2026-01-29.log   # 下一天的日志
```

### 命名规则

- **GUI 日志**: `db-benchmind-YYYY-MM-DD.log`
- **CLI 日志**: `db-benchmind-cli-YYYY-MM-DD.log`

---

## 日志格式

### 结构化日志格式

```log
time=2026-01-28T07:40:08.438Z level=INFO msg="Starting DB-BenchMind" log_file=data/logs/db-benchmind-2026-01-28.log
```

### 字段说明

| 字段 | 说明 | 示例 |
|------|------|------|
| `time` | ISO 8601 时间戳（UTC） | `2026-01-28T07:40:08.438Z` |
| `level` | 日志级别 | `INFO`, `WARN`, `ERROR`, `DEBUG` |
| `msg` | 日志消息 | `Starting DB-BenchMind` |
| `key=value` | 附加属性（键值对） | `log_file=...`, `error=...` |

---

## 日志级别

### 级别定义

| 级别 | 用途 | 示例 |
|------|------|------|
| `DEBUG` | 调试信息（默认关闭） | 函数入口/出口、变量值 |
| `INFO` | 关键业务节点 | 应用启动、连接创建、测试开始 |
| `WARN` | 可恢复的异常 | 连接超时重试、工具未检测到 |
| `ERROR` | 不可恢复的错误 | 数据库连接失败、文件读写错误 |

### 日志级别示例

```log
# INFO 级别
time=2026-01-28T07:40:08.440Z level=INFO msg="Database initialized" path=./data/db-benchmind.db

# WARN 级别
time=2026-01-28T07:40:09.123Z level=WARN msg="Tool not found" tool=swingbench

# ERROR 级别
time=2026-01-28T07:40:10.456Z level=ERROR msg="Failed to connect" error="connection refused" host=localhost:3306
```

---

## 查看日志

### 方法 1：实时监控（推荐用于调试）

```bash
# 实时查看 GUI 日志
tail -f ./data/logs/db-benchmind-$(date +%Y-%m-%d).log

# 实时查看 CLI 日志
tail -f ./data/logs/db-benchmind-cli-$(date +%Y-%m-%d).log

# 同时监控两个日志文件
tail -f ./data/logs/*.log
```

### 方法 2：查看当日日志

```bash
# 查看 GUI 当日日志
cat ./data/logs/db-benchmind-$(date +%Y-%m-%d).log

# 查看 CLI 当日日志
cat ./data/logs/db-benchmind-cli-$(date +%Y-%m-%d).log
```

### 方法 3：查看最近操作

```bash
# 查看最近 20 行
tail -n 20 ./data/logs/db-benchmind-cli-2026-01-28.log

# 查看最近 50 行
tail -n 50 ./data/logs/db-benchmind-2026-01-28.log
```

### 方法 4：搜索特定内容

```bash
# 搜索所有错误
grep -i "error" ./data/logs/*.log

# 搜索连接相关操作
grep "connection" ./data/logs/*.log

# 搜索特定时间的日志
grep "2026-01-28T07:" ./data/logs/*.log

# 搜索包含特定关键词的行
grep "Adding connection\|Testing connection\|Deleting connection" ./data/logs/*.log
```

### 方法 5：查看日志统计

```bash
# 统计错误数量
grep -c "ERROR" ./data/logs/db-benchmind-2026-01-28.log

# 统计各级别日志数量
grep -c "INFO" ./data/logs/db-benchmind-2026-01-28.log
grep -c "WARN" ./data/logs/db-benchmind-2026-01-28.log
grep -c "ERROR" ./data/logs/db-benchmind-2026-01-28.log
```

---

## 日志分析

### 常见分析场景

#### 1. 查找应用启动问题

```bash
# 查看启动过程的所有日志
grep "Starting\|initialized\|failed" ./data/logs/db-benchmind-2026-01-28.log
```

#### 2. 分析连接失败原因

```bash
# 查找连接相关错误
grep -A 5 -B 5 "Failed to connect" ./data/logs/*.log
```

#### 3. 追踪压测任务执行

```bash
# 查看任务开始和结束
grep "task.*started\|task.*completed\|task.*failed" ./data/logs/*.log

# 查看特定任务 ID 的日志
grep "task_id=abc-123" ./data/logs/*.log
```

#### 4. 性能分析

```bash
# 查找性能相关日志
grep "TPS\|latency\|duration" ./data/logs/*.log
```

### 使用 jq 进行结构化分析

如果需要更复杂的分析，可以将日志转换为 JSON 格式：

```bash
# 提取特定字段（假设日志是 JSON 格式）
jq 'select(.level=="ERROR")' ./data/logs/*.log
```

---

## 日志维护

### 日志轮转

建议使用 `logrotate` 管理日志文件：

创建 `/etc/logrotate.d/db-benchmind`：

```
/opt/project/DB-BenchMind/data/logs/*.log {
    daily
    rotate 30
    compress
    delaycompress
    missingok
    notifempty
    create 0644 root root
    sharedscripts
    postrotate
        # 应用程序会自动创建新日志文件
    endscript
}
```

### 手动清理

```bash
# 清理 7 天前的日志
find ./data/logs -name "*.log" -mtime +7 -delete

# 清理 30 天前的压缩日志
find ./data/logs -name "*.log.gz" -mtime +30 -delete

# 压缩旧日志
gzip ./data/logs/db-benchmind-2026-01-20.log
```

### 日志大小限制

如果需要限制日志文件大小，可以在应用中添加日志轮转配置（未来版本）。

---

## 日志最佳实践

### 1. 日志查看工作流

```bash
# 1. 查看最近的错误
tail -n 100 ./data/logs/*.log | grep ERROR

# 2. 如果发现问题，查看上下文
grep -B 10 -A 10 "error message" ./data/logs/*.log

# 3. 实时监控日志以复现问题
tail -f ./data/logs/*.log
```

### 2. 日志保留策略

- **开发环境**: 保留 7 天
- **测试环境**: 保留 30 天
- **生产环境**: 保留 90 天或更久

### 3. 日志备份

```bash
# 定期备份日志到归档目录
cp -r ./data/logs /backup/logs-$(date +%Y%m%d)
```

### 4. 监控日志磁盘使用

```bash
# 检查日志目录大小
du -sh ./data/logs

# 查找大文件
find ./data/logs -type f -size +100M
```

---

## 故障排查

### 问题：日志文件不存在

**原因**: 应用未启动或日志目录权限问题

**解决方案**:
```bash
# 检查日志目录
ls -la ./data/logs

# 检查目录权限
chmod 755 ./data/logs

# 重启应用
./build/db-benchmind
```

### 问题：日志文件为空

**原因**: 应用刚启动或日志写入失败

**解决方案**:
```bash
# 检查应用是否正常运行
ps aux | grep db-benchmind

# 查看控制台输出是否有错误
./build/db-benchmind
```

### 问题：日志文件过大

**原因**: 长时间运行未清理

**解决方案**:
```bash
# 压缩旧日志
gzip ./data/logs/db-benchmind-*.log

# 清理旧日志
find ./data/logs -name "*.log" -mtime +30 -delete
```

---

## 附录：日志快速参考

| 操作 | 命令 |
|------|------|
| 查看当日日志 | `cat ./data/logs/db-benchmind-$(date +%Y-%m-%d).log` |
| 实时监控 | `tail -f ./data/logs/db-benchmind-$(date +%Y-%m-%d).log` |
| 搜索错误 | `grep -i error ./data/logs/*.log` |
| 统计错误数 | `grep -c ERROR ./data/logs/*.log` |
| 查看最近 20 行 | `tail -n 20 ./data/logs/*.log` |
| 列出所有日志 | `ls -lah ./data/logs/` |
| 清理 7 天前日志 | `find ./data/logs -name "*.log" -mtime +7 -delete` |

---

## 相关文档

- [用户手册](./USER_GUIDE.md)
- [开发者指南](./DEVELOPER_GUIDE.md)
- [API 参考](./API_REFERENCE.md)
- [测试文档](./TESTING.md)
