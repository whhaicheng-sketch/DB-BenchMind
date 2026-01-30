# GUI 开发规范

**Version:** 1.0
**Date:** 2026-01-28
**Status:** Active

---

## 1. 日志记录规范 (Logging Standards)

### 1.1 强制要求：所有 GUI 交互必须记录日志

**原则：** GUI 层的所有用户交互操作都必须记录日志，确保可追溯性和可调试性。

#### 1.1.1 必须记录的操作类型

所有以下操作都必须记录日志：

1. **按钮点击事件**
   - 所有工具栏按钮（Add, Delete, Edit, Test, Refresh 等）
   - 所有对话框按钮（Save, Cancel, OK, Confirm 等）
   - 所有功能按钮（Run, Stop, Set Default 等）

2. **列表/表格交互**
   - 单击选择操作
   - 双击打开详情
   - 多选操作

3. **对话框操作**
   - 对话框打开
   - 对话框关闭
   - 对话框中的字段变更

4. **表单操作**
   - 表单提交
   - 表单验证失败
   - 表单字段变更

5. **页面/标签切换**
   - 页面切换
   - 标签页切换

#### 1.1.2 日志级别使用

- **INFO**：正常的用户操作
  ```go
  slog.Info("Templates: Add Template button clicked")
  slog.Info("Templates: Template selected", "template", tmplName, "index", idx)
  ```

- **WARN**：用户操作可能存在问题
  ```go
  slog.Warn("Templates: Template name is empty")
  slog.Warn("Connection: No connection selected before delete")
  ```

- **ERROR**：操作失败或错误
  ```go
  slog.Error("Templates: Failed to save template", "error", err)
  ```

#### 1.1.3 日志格式规范

**基本格式：**
```go
slog.Level("<Page>: <Action>", <key-value pairs>...)
```

**示例：**
```go
// 页面级别操作
slog.Info("Templates: Add Template button clicked")
slog.Info("Connections: Test Connection button clicked", "connection", connName)

// 对象操作
slog.Info("Templates: Template selected", "template", tmplName, "index", idx)
slog.Info("Templates: Creating new template", "name", name)

// 操作结果
slog.Info("Templates: Template added successfully", "name", name, "total_templates", len(p.templates))
slog.Info("Connections: Connection tested successfully", "connection", connName, "latency_ms", result.LatencyMs)

// 错误情况
slog.Error("Templates: Failed to validate template", "error", err, "name", name)
```

#### 1.1.4 必须包含的关键信息

日志记录必须包含以下信息之一或多个：

1. **操作来源**：哪个页面/组件
2. **操作类型**：点击、选择、双击、提交等
3. **操作对象**：模板名称、连接名称、任务 ID 等
4. **操作结果**：成功、失败、部分成功
5. **关键参数**：索引、数量、时间戳等

### 1.2 日志记录实施检查清单

在代码审查时，必须检查以下内容：

- [ ] 所有按钮点击都有日志记录
- [ ] 所有列表选择都有日志记录
- [ ] 所有对话框操作都有日志记录
- [ ] 所有表单提交都有日志记录
- [ ] 所有错误情况都有日志记录
- [ ] 日志包含足够的上下文信息
- [ ] 日志级别使用正确

### 1.3 日志文件位置

**主日志文件：**
```
data/logs/db-benchmind-YYYY-MM-DD.log
```

**实时查看日志：**
```bash
tail -f data/logs/db-benchmind-$(date +%Y-%m-%d).log
```

**查看最新日志：**
```bash
tail -50 data/logs/db-benchmind-$(date +%Y-%m-%d).log
```

**搜索特定页面日志：**
```bash
grep "Templates:" data/logs/db-benchmind-$(date +%Y-%m-%d).log
grep "Connections:" data/logs/db-benchmind-$(date +%Y-%m-%d).log
```

---

## 2. GUI 开发最佳实践

### 2.1 布局规范

1. **使用 Border 布局填充空间**
   ```go
   content := container.NewBorder(
       topArea,                              // top - toolbar
       nil,                                   // bottom
       nil,                                   // left
       nil,                                   // right
       container.NewScroll(listContainer),    // center - fills available space
   )
   ```

2. **使用 VBox + Scroll 代替 widget.List**
   - widget.List 创建时固定项目数量，不适应数据变化
   - VBox + Scroll 支持动态列表重建

### 2.2 列表和选择

1. **单击选中，双击打开详情**
   ```go
   // 双击检测（500ms 内）
   if idx == p.lastClickIndex && now.Sub(p.lastClickTime) < 500*time.Millisecond {
       // 双击 - 打开详情
       p.showDetails(item)
   } else {
       // 单击 - 仅选中
       p.selected = idx
       p.updateSelectionVisual()
   }
   ```

2. **选中高亮显示**
   ```go
   func (p *Page) updateSelectionVisual() {
       for i, obj := range p.listContainer.Objects {
           if btn, ok := obj.(*widget.Button); ok {
               if i == p.selected {
                   btn.Importance = widget.HighImportance // 选中
               } else {
                   btn.Importance = widget.MediumImportance // 普通
               }
               btn.Refresh()
           }
       }
   }
   ```

### 2.3 数据管理

1. **分离数据加载和 UI 刷新**
   ```go
   // loadTemplates() - 从数据源加载数据
   func (p *Page) loadTemplates() {
       p.templates = p.loadTemplatesFromSource()
       p.refreshTemplateList()
   }

   // refreshTemplateList() - 仅刷新 UI
   func (p *Page) refreshTemplateList() {
       // 重建列表 UI
   }
   ```

2. **避免数据丢失**
   - 添加/删除操作后调用 `refreshTemplateList()` 而不是 `loadTemplates()`
   - `loadTemplates()` 会从数据源重新加载，覆盖内存中的修改

### 2.4 对话框规范

1. **使用自定义对话框**
   ```go
   dlg := dialog.NewCustomWithoutButtons(title, content, win)
   dlg.Resize(fyne.NewSize(500, 700))
   dlg.Show()
   ```

2. **对话框按钮布局**
   ```go
   btnSave := widget.NewButton("Save", func() {
       d.onSave()
       d.dialog.Hide()
   })
   btnSave.Importance = widget.HighImportance

   btnCancel := widget.NewButton("Cancel", func() {
       dlg.Hide()
   })

   buttonContainer := container.NewHBox(btnSave, btnCancel)
   ```

### 2.5 图标使用

使用 emoji 图标增强可识别性：

- 📦 内置项
- 📄 自定义项
- ⭐ 默认项/选中项
- ➕ 添加
- 🗑️ 删除
- ✏️ 编辑
- 🔌 测试连接
- 🔄 刷新
- ⭐ 设置默认
- ▶️ 运行
- ■ 停止

---

## 3. 代码示例

### 3.1 完整的页面结构示例

```go
type ExamplePage struct {
    win             fyne.Window
    items           []Item
    selected        int
    listContainer   *fyne.Container
    lastClickTime   time.Time
    lastClickIndex  int
}

func NewExamplePage(win fyne.Window) fyne.CanvasObject {
    page := &ExamplePage{
        win:       win,
        selected:  -1,
        items:     []Item{},
    }

    page.listContainer = container.NewVBox()
    page.loadItems()

    // 工具栏
    toolbar := container.NewVBox(
        container.NewHBox(
            widget.NewButton("➕ Add", func() { page.onAdd() }),
            widget.NewButton("🗑️ Delete", func() { page.onDelete() }),
            widget.NewButton("🔄 Refresh", func() { page.loadItems() }),
        ),
    )

    topArea := container.NewVBox(
        toolbar,
        widget.NewSeparator(),
        widget.NewLabel("Items:"),
    )

    content := container.NewBorder(
        topArea,
        nil,
        nil,
        nil,
        container.NewScroll(page.listContainer),
    )

    return content
}

func (p *ExamplePage) loadItems() {
    slog.Info("Example: Loading items")
    p.items = p.loadItemsFromSource()
    p.refreshList()
}

func (p *ExamplePage) refreshList() {
    slog.Info("Example: Refreshing list", "count", len(p.items))
    p.listContainer.Objects = nil

    for i, item := range p.items {
        idx := i
        text := fmt.Sprintf("📄  %s", item.Name)
        btn := widget.NewButton(text, func() {
            now := time.Now()
            if idx == p.lastClickIndex && now.Sub(p.lastClickTime) < 500*time.Millisecond {
                // 双击
                slog.Info("Example: Double-click", "item", item.Name)
                p.showDetails(item)
            } else {
                // 单击
                slog.Info("Example: Selected", "item", item.Name, "index", idx)
                p.selected = idx
                p.updateSelectionVisual()
            }
            p.lastClickTime = now
            p.lastClickIndex = idx
        })
        btn.Importance = widget.MediumImportance
        p.listContainer.Add(btn)
    }

    p.listContainer.Refresh()
    p.updateSelectionVisual()
}

func (p *ExamplePage) onAdd() {
    slog.Info("Example: Add button clicked")
    // 显示添加对话框
    // ...
    slog.Info("Example: Item added", "name", name)
    p.refreshList()  // 仅刷新 UI
}
```

---

## 4. 测试清单

在提交 GUI 代码前，必须通过以下测试：

- [ ] 所有按钮都有日志记录
- [ ] 所有操作都可在日志中追溯
- [ ] 单击选中功能正常
- [ ] 双击打开详情功能正常
- [ ] 选中高亮显示正常
- [ ] 添加/删除操作后列表正确更新
- [ ] 对话框打开和关闭正常
- [ ] 表单验证有日志记录
- [ ] 错误情况有日志记录

---

## 5. 参考资料

- Fyne 文档: https://docs.fyne.io/
- Go slog 包: https://pkg.go.dev/log/slog
- 项目宪法: `/opt/project/DB-BenchMind/constitution.md`
- 项目协作规范: `/opt/project/DB-BenchMind/CLAUDE.md`
