# DB-BenchMind 快速启动

## 唯一受支持的入口

项目根目录只保留两个可执行脚本：

```bash
./build   # 构建 Wails 桌面程序到 ./bin/db-benchmind
./start   # 启动程序；若缺少二进制会自动先执行 ./build
```

后续任何改动都必须至少验证这两个命令可以正常执行。

## 快速开始

### 1. 进入项目根目录

```bash
cd /opt/project/DB-BenchMind
```

### 2. 构建程序

```bash
./build
```

### 3. 启动程序

```bash
./start
```

## 运行要求

- 必须从项目根目录执行
- 需要可用的桌面显示环境
- 首次运行会自动创建 `./data/` 下的数据库和运行数据

## 说明

- `./build` 会通过 Wails 构建桌面程序，并生成 `./bin/db-benchmind`
- `./start` 会优先复用 `./bin/db-benchmind`；若二进制不存在，会先自动构建
- 若启动时看到 `libEGL` 相关 warning，通常是图形环境权限告警，不等同于程序启动失败

## 相关文档

- [README.md](./README.md)
- [CLI_USAGE.md](./CLI_USAGE.md)
