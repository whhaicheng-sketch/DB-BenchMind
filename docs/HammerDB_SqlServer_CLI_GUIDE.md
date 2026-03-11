# HammerDB SQL Server CLI 命令参考手册

> **重要说明**：
> - SSH 隧道不参与任何压测功能（宪法 Article X）
> - 数据库连接使用直连：IP:端口 + 用户名 + 密码
> - 本文档基于您提供的 GUI 截图流程编写
> - 默认 Test 模板已更新为 10 warehouses（快速测试）

---

## 目录

1. [进入 CLI](#0-进入-cli)
2. [测试快速测试模板 10 Warehouses](#测试快速测试模板-10-warehouses)
3. [TPROC-C 完整测试模板 100 Warehouses](#tproc-c-完整测试模板-100-warehouses)
4. [当前实现问题分析](#当前实现问题分析)
5. [常见参数说明](#常见参数说明)

---

## 0）进入 CLI

```bash
# 方式1：进入交互模式
hammerdbcli

# 方式2：查看版本
hammerdbcli v
```

进入后您会看到提示符 `hammerdb>`，后续所有命令都在这里输入。

---

## 1）测试快速测试模板（10 Warehouses）

这个模板适用于快速测试 SQL Server，使用 10 个 warehouses。

### 1.1 选择数据库类型和基准

```tcl
dbset db mssqls
```

**作用**：选择 SQL Server 作为数据库，TPROC-C 作为基准测试类型

---

### 1.2 配置数据库连接信息

根据截图，您的配置是：
- Host: 10.5.54.87
- Port: 1433
- Username: sa
- Password: SqlServer@2019
- ODBC Driver: ODBC Driver 17 for SQL Server

```tcl
diset connection mssqls_server {10.5.54.87}
diset connection mssqls_port 1433
diset connection mssqls_odbc_driver {ODBC Driver 17 for SQL Server}
diset connection mssqls_authentication {SQL Server Authentication}
diset connection mssqls_uid {sa}
diset connection mssqls_pass {SqlServer@2019}
```

**命令说明**：

| 命令 | 参数说明 |
|------|---------|
| `dbset db mssqls` | 设置数据库类型为 Microsoft SQL Server |
| `diset connection mssqls_server {IP}` | 设置 SQL Server 主机/IP 地址 |
| `diset connection mssqls_port 1433` | 设置 SQL Server 端口（默认 1433）|
| `diset connection mssqls_odbc_driver` | 设置 ODBC 驱动名 |
| `diset connection mssqls_authentication` | 设置认证方式为 SQL Server 认证 |
| `diset connection mssqls_uid {user}` | 设置 SQL Server 登录用户 |
| `diset connection mssqls_pass {pwd}` | 设置 SQL Server 登录密码 |

> **注意**：密码建议用花括号 `{密码}` 包住，避免特殊字符解析问题。

---

### 1.3 配置数据库名和仓库数

```tcl
diset tpcc mssqls_dbase {tpcc}
diset tpcc mssqls_count_ware 10
```

**命令说明**：

| 命令 | 参数说明 |
|------|---------|
| `diset tpcc mssqls_dbase tpcc` | 设置 TPROC-C 的目标数据库名（数据库必须存在） |
| `diset tpcc mssqls_count_ware 10` | 设置仓库数（10 个） |

---

### 1.4 配置构建并发用户数

```tcl
diset tpcc mssqls_num_vu 10
```

**命令说明**：
- `num_vu` = 并发构建用户数（更多线程=建表越快）

---

### 1.5 验证当前配置

```tcl
print dict
```

**作用**：打印当前所有 `diset` 设置，确认参数正确。

输出示例：
```
dict
...
connection:mssqls_server: 10.5.54.87
connection:mssqls_port: 1433
connection:mssqls_odbc_driver: ODBC Driver 17 for SQL Server
connection:mssqls_authentication: SQL Server Authentication
connection:mssqls_uid: sa
connection:mssqls_pass: SqlServer@2019
...
```

> **重要**：如果发现参数值不正确，重新输入对应命令修改。

---

### 1.6 执行 Schema Build（建库建表）

```tcl
buildschema
```

**作用**：开始执行建库建表和初始化数据流程。

过程：
1. 连接 SQL Server
2. 创建 TPROC-C 表结构（warehouse, stock, district, customer, orders, history, item, new_order）
3. 按照配置的仓库数生成测试数据
4. 创建索引
5. 初始化数据

---

### 1.7 等待建表完成

```tcl
waittocomplete
```

**作用**：等待 `buildschema` 命令执行完成。

> **注意**：某些版本可能叫 `vucomplete` 或 `vudestroy`，如果 `waittocomplete` 不存在，用 `help` 查找。

---

## 2）TPROC-C 完整测试模板（100 Warehouses）

这个模板适用于完整压测，使用 100 个 warehouses。

### 2.1 选择数据库类型和基准

```tcl
dbset db mssqls
dbset bm TPROC-C
```

**作用**：选择 SQL Server，TPROC-C 为基准类型

---

### 2.2 配置数据库连接信息

```tcl
diset connection mssqls_server {10.5.54.87}
diset connection mssqls_port 1433
diset connection mssqls_odbc_driver {ODBC Driver 17 for SQL Server}
diset connection mssqls_authentication {SQL Server Authentication}
diset connection mssqls_uid {sa}
diset connection mssqls_pass {SqlServer@2019}
```

> **重要**：连接参数与测试模板相同。

---

### 2.3 配置数据库名和仓库数

```tcl
diset tpcc mssqls_dbase {tpcc}
diset tpcc mssqls_count_ware 100
```

**参数说明**：
- `mssqls_count_ware 100`：100 个仓库（大规模压测）

---

### 2.4 配置压测驱动和运行参数

```tcl
diset tpcc mssqls_driver timed
diset tpcc mssqls_rampup 2
diset tpcc mssqls_duration 10
```

**命令说明**：

| 命令 | 参数说明 |
|------|---------|
| `diset tpcc mssqls_driver timed` | 使用时间驱动模式（运行固定时长） |
| `diset tpcc mssqls_rampup 2` | 预热时长 2 分钟 |
| `diset tpcc mssqls_duration 10` | 正式压测运行 10 分钟 |

> **说明**：如果要用固定事务数模式，把 `timed` 改为 `user`，然后设置 `mssqls_total_iterations`。

---

### 2.5 配置压测虚拟用户数

```tcl
vuset vu 10
```

**命令说明**：
- 设置压测并发虚拟用户数为 10

---

### 2.6 配置 VU 参数（可选但建议）

```tcl
vuset delay 500
vuset repeatdelay 500
vuset iterations 1
```

**命令说明**：

| 命令 | 参数说明 |
|------|---------|
| `vuset delay 500` | 每个事务之间延迟 500ms（模拟用户思考时间） |
| `vuset repeatdelay 500` | 每轮迭代之间延迟 500ms |
| `vuset iterations 1` | 每个用户执行脚本 1 次 |

---

### 2.7 验证压测配置

```tcl
print vuconf
```

**作用**：打印当前 VU 配置，确认所有参数已正确设置。

---

### 2.8 加载压测脚本

```tcl
loadscript
```

**作用**：加载 TPROC-C 的压测脚本（包含事务逻辑）。

---

### 2.9 创建虚拟用户

```tcl
vucreate
```

**作用**：创建虚拟用户线程（按 `vuset vu 10` 的数量）。

---

### 2.10 启动事务计数器（实时监控用）

```tcl
tcset refreshrate 10
tcstart
```

**命令说明**：
- `tcset refreshrate 10`：每 10 秒刷新一次计数器
- `tcstart`：启动事务计数

---

### 2.11 执行压测

```tcl
vurun
```

**作用**：启动压测，所有 VU 开始执行脚本。

---

### 2.12 等待压测完成（固定时长模式）

```tcl
runtimer 600
```

**作用**：让 CLI 主线程等待 600 秒（10 分钟），压测自动终止。

> **说明**：`runtimer` 是固定时长模式的必备命令，否则脚本一跑完就退出。

---

### 2.13 停止计数器

```tcl
tcstop
```

**作用**：停止事务计数器。

---

### 2.14 清理虚拟用户

```tcl
vudestroy
```

**作用**：销毁虚拟用户线程，释放资源。

---

### 2.15 退出 CLI

```tcl
quit
```

**作用**：退出 HammerDB CLI。

---

## 当前实现问题分析

### 问题 1：HammerDB 连接字符串显示 localhost

**现象**：
```
Error in Virtual User 1: Connection to DRIVER=ODBC Driver 18 for SQL Server;SERVER=localhost;...
```

**原因**：
- 当前代码已正确设置 `diset connection mssqls_server 192.168.134.129`
- 但 ODBC 连接字符串仍使用 localhost
- 可能是 HammerDB 有缓存的默认配置

**状态**：已尝试清除连接设置，但问题可能仍存在

### 问题 2：Test 模板默认值更新

**已修复**：✅ Test 模板的 warehouses 默认值从 1 改为 10

---

## 常见参数说明

### 数据库连接参数

| 参数 | 用途 | 示例值 |
|------|------|--------|
| `mssqls_server` | SQL Server 主机/IP | `10.5.54.87` |
| `mssqls_port` | SQL Server 端口 | `1433` |
| `mssqls_odbc_driver` | ODBC 驱动名 | `ODBC Driver 17 for SQL Server` |
| `mssqls_authentication` | 认证方式 | `SQL Server Authentication` |
| `mssqls_uid` | 登录用户 | `sa` |
| `mssqls_pass` | 登录密码 | `{密码}` |

### TPROC-C 参数

| 参数 | 用途 | 示例值 |
|------|------|--------|
| `mssqls_dbase` | 目标数据库名 | `tpcc` |
| `mssqls_count_ware` | 仓库数 | `10` (测试) / `100` (完整压测） |
| `mssqls_num_vu` | 压测并发用户数 | `10` |
| `mssqls_driver` | 驱动类型 | `timed` (时间驱动） / `user` (事务驱动） |
| `mssqls_rampup` | 预热时长（分钟） | `2` |
| `mssqls_duration` | 压测时长（分钟） | `10` |

### VU 配置参数

| 参数 | 用途 | 建议值 |
|------|------|--------|
| `vu` | 虚拟用户数 | `10` (按 warehouses 数） |
| `delay` | 事务延迟（毫秒） | `500` |
| `repeatdelay` | 迭代延迟（毫秒） | `500` |
| `iterations` | 每用户执行次数 | `1` |

---

## 快速参考：完整压测脚本

如果您想一次性运行完整压测，可以使用以下脚本：

### 快速测试（10 Warehouses, 10 分钟）

```tcl
dbset db mssqls
dbset bm TPROC-C
diset connection mssqls_server {10.5.54.87}
diset connection mssqls_port 1433
diset connection mssqls_odbc_driver {ODBC Driver 17 for SQL Server}
diset connection mssqls_authentication {SQL Server Authentication}
diset connection mssqls_uid {sa}
diset connection mssqls_pass {SqlServer@2019}
diset tpcc mssqls_dbase {tpcc}
diset tpcc mssqls_count_ware 10
diset tpcc mssqls_driver timed
diset tpcc mssqls_rampup 2
diset tpcc mssqls_duration 10
vuset vu 10
vuset delay 500
vuset repeatdelay 500
vuset iterations 1
loadscript
vucreate
tcset refreshrate 10
tcstart
vurun
runtimer 600
tcstop
vudestroy
quit
```

### 完整压测（100 Warehouses, 10 分钟）

```tcl
dbset db mssqls
dbset bm TPROC-C
diset connection mssqls_server {10.5.54.87}
diset connection mssqls_port 1433
diset connection mssqls_odbc_driver {ODBC Driver 17 for SQL Server}
diset connection mssqls_authentication {SQL Server Authentication}
diset connection mssqls_uid {sa}
diset connection mssqls_pass {SqlServer@2019}
diset tpcc mssqls_dbase {tpcc}
diset tpcc mssqls_count_ware 100
diset tpcc mssqls_driver timed
diset tpcc mssqls_rampup 2
diset tpcc mssqls_duration 10
vuset vu 10
vuset delay 500
vuset repeatdelay 500
vuset iterations 1
loadscript
vucreate
tcset refreshrate 10
tcstart
vurun
runtimer 600
tcstop
vudestroy
quit
```

---

## 调试技巧

### 查看可用字典和键

```tcl
print dict
```

### 查看当前连接配置

```tcl
print connection
```

### 查看 VU 配置

```tcl
print vuconf
```

### 查看当前数据库和基准

```tcl
print db
print bm
```

---

## 参考资料

- HammerDB 官方文档：https://www.hammerdb.com/docs/
- SQL Server TPROC-C 规范：https://www.tpc.org/tpc_spec_default.html
