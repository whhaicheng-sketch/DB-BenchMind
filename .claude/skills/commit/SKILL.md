---
name: commit
description: 在用户输入 /commit 时，检查改动、只暂存本次相关文件、生成规范提交信息，并推送到 main。
---

# /commit

当前项目默认只使用 `main` 分支。  
目标：检查改动 → 暂存相关文件 → 创建规范 commit → 推送到 `main`。

## 执行流程

用户输入：

```text
/commit
````

按顺序执行：

```bash
git status --short
git diff --stat HEAD
git branch --show-current
git log --oneline -5
```

必要时再执行：

```bash
git diff --name-only HEAD
git diff --cached --stat
```

## 规则

1. 默认优先使用：

```bash
git add <相关文件路径>
```

不要直接使用：

```bash
git add .
git add -A
```

只有在以下条件都满足时才允许：

* 改动主题单一
* 没有无关文件
* 没有敏感文件
* 没有日志、数据库、构建产物

2. 当前分支必须是：

```bash
main
```

如果不是 `main`，或者处于 detached HEAD，必须暂停，不允许继续提交和推送。

3. 暂存后必须复查：

```bash
git diff --cached --stat
```

如果发现混入以下内容，必须先清理：

* 无关文件
* 构建产物
* 日志文件
* 数据库文件
* 敏感文件

4. 默认推送到：

```bash
git push origin main
```

## 必须暂停的情况

遇到以下任一情况，暂停并询问用户：

* 检测到敏感文件或敏感内容
* 改动超过 50 个文件
* 无法判断哪些文件属于本次任务
* 当前不在 `main`
* 暂存内容混入明显无关文件

## 敏感文件示例

* `.env`
* `.env.*`
* `*.pem`
* `*.key`
* `id_rsa`
* `id_ed25519`
* `credentials`
* `secrets.*`

如果发现明文密码、Token、私钥、真实数据库连接凭据，也必须暂停。

## .gitignore 策略

只有在未跟踪文件中出现明显噪音时，才更新 `.gitignore`，例如：

* `node_modules/`
* `dist/`
* `*.log`
* `*.db`
* `tmp/`
* 临时测试脚本

不要为了完整而无条件修改 `.gitignore`。

## Commit 规范

格式：

```text
<type>(<scope>): <subject>

[optional body]

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>
```

常用 `type`：

* `feat`
* `fix`
* `refactor`
* `docs`
* `test`
* `chore`

常用 `scope`：

* `ui`
* `connection`
* `template`
* `tasks-monitor`
* `benchmark`

要求：

* subject 使用祈使句
* 不超过 72 个字符
* 不要写成 `update`、`misc`、`fix bug`

## 推荐模板

检查状态：

```bash
git status --short
git diff --stat HEAD
git branch --show-current
git log --oneline -5
```

只暂存相关文件：

```bash
git add <相关文件路径>
git diff --cached --stat
```

创建提交：

```bash
git commit -m "$(cat <<'EOF'
feat(tasks-monitor): refine monitor-first task panel

- simplify current task layout and highlight start stop actions
- keep elapsed fixed after completion
- give more space to throughput and system metrics

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>
EOF
)"
```

推送：

```bash
git push origin main
```

## 预期行为

如果没有改动，回复：

```text
Nothing to commit
```

如果存在风险，暂停并说明原因。

如果可以安全提交，则执行：

1. 检查改动
2. 只暂存本次相关文件
3. 复查暂存内容
4. 生成规范提交信息
5. 创建 commit
6. 推送到 `main`

```

