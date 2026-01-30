# DB-BenchMind UI 布局模式与实践指南

**Version:** 2.0
**Date:** 2026-01-30
**Status:** Active
**Based On:** 实际开发 5 个 GUI 页面的经验总结

---

## 目录

1. [Fyne 布局系统深度解析](#1-fyne-布局系统深度解析)
2. [常见布局问题与解决方案](#2-常见布局问题与解决方案)
3. [已实现页面的布局模式](#3-已实现页面的布局模式)
4. [UI 组件尺寸控制规范](#4-ui-组件尺寸控制规范)
5. [性能优化最佳实践](#5-性能优化最佳实践)
6. [检查清单](#6-检查清单)

---

## 1. Fyne 布局系统深度解析

### 1.1 核心布局容器特性

#### Border 布局 ⭐ **最常用，最重要**

**特性：**
- `Top`/`Bottom`：固定高度，垂直布局
- `Left`/`Right`：固定宽度，水平布局
- `Center`：**自动扩展填充所有剩余空间** ⭐

**关键点：**
```go
content := container.NewBorder(
    topObj,    // Top - 固定高度
    bottomObj, // Bottom - 固定高度
    leftObj,   // Left - 固定宽度
    rightObj,  // Right - 固定宽度
    centerObj, // Center - 自动扩展 ⭐
)
```

**使用场景：**
- ✅ 需要某个组件自动扩展填充空间
- ✅ 上下分区，上部固定，下部自适应
- ✅ 左右分区，中间自适应

**⚠️ 常见错误：**
```go
// ❌ 错误：Center 被 VBox 包裹，无法扩展
content := container.NewBorder(
    topObj,
    bottomObj,
    nil, nil,
    container.NewVBox(centerObj), // ❌ VBox 限制了扩展
)

// ✅ 正确：Center 直接放置可扩展组件
content := container.NewBorder(
    topObj,
    bottomObj,
    nil, nil,
    centerScroll, // ✅ 直接放 Scroll 或其他可扩展组件
)
```

#### Grid 布局

**特性：**
- `NewGridWithRows(n)`：创建 n 行，**每行等高**（平均分配）
- `NewGridWithColumns(n)`：创建 n 列，**每列等宽**（平均分配）
- `NewGridLayout(rows, cols)`：创建网格，每个单元格等大

**使用场景：**
- ✅ 需要上下或左右等分空间
- ✅ 多个并列的等宽/等高区域

**实际应用：**
```go
// Comparison 页面：上下各 50%
content := container.NewGridWithRows(2,
    selectionArea,  // 上半部分：50%
    resultsArea,    // 下半部分：50%
)
```

#### VBox / HBox 布局

**特性：**
- `VBox`：垂直排列，子组件高度由内容决定
- `HBox`：水平排列，子组件宽度由内容决定
- **不会自动扩展子组件**

**使用场景：**
- ✅ 工具栏（按钮横向排列）
- ✅ 表单字段（垂直排列）
- ✅ 固定大小的组件组合

**⚠️ 注意：**
```go
// VBox 不会让子组件扩展
vbox := container.NewVBox(
    widget1,
    widget2,
)
// widget1 和 widget2 保持各自的内容尺寸，不会扩展
```

#### Scroll 容器

**特性：**
- 包装其他组件，提供滚动功能
- **需要设置内容的最小尺寸**才能正确显示滚动条

**使用场景：**
- ✅ 列表内容超出可见区域
- ✅ 长文本显示
- ✅ 任何需要滚动的内容

**⚠️ 常见问题：**
```go
// ❌ 问题：Scroll 包裹在 VBox 中，可能不显示滚动条
container.NewVBox(
    container.NewScroll(content), // ❌ 可能无法正确扩展
)

// ✅ 正确：Scroll 直接放在可扩展位置
container.NewBorder(
    top,
    nil, nil, nil,
    container.NewScroll(content), // ✅ 作为 Center，可扩展
)
```

---

## 2. 常见布局问题与解决方案

### 2.1 问题：组件只显示一行/一列

**症状：** List、Entry、TextArea 只显示一条记录

**原因：**
1. 组件被包裹在不能扩展的容器中（VBox、HBox）
2. 没有设置最小尺寸
3. 父容器没有给子组件分配空间

**解决方案：**

#### 方案 1：使用 Border + Center（推荐）⭐
```go
// ✅ 最佳实践
listScroll := container.NewScroll(page.list)

content := container.NewBorder(
    filterForm,    // Top
    nil,           // Bottom
    nil,           // Left
    nil,           // Right
    listScroll,    // Center - 自动扩展
)
```

#### 方案 2：使用 Grid 等分空间
```go
// ✅ 上下各 50%
content := container.NewGridWithRows(2,
    topArea,    // 50%
    bottomArea, // 50%
)
```

#### 方案 3：设置组件最小尺寸
```go
// Entry/TextArea
entry.SetMinRowsVisible(20) // 设置最少显示 20 行

// List
list.Resize(fyne.NewSize(width, height)) // ⚠️ 有限制
```

### 2.2 问题：Resize() 不生效

**症状：** 调用 `Resize()` 后组件尺寸没有变化

**原因：**
- Fyne 中 `Resize()` 只对**顶层容器**或**不在布局管理器中的组件**有效
- 子容器的尺寸由父布局管理器决定

**解决方案：**

#### ❌ 错误做法
```go
// ❌ Resize 对 VBox 的子组件无效
vbox := container.NewVBox(child)
child.Resize(fyne.NewSize(800, 600)) // 不会生效
```

#### ✅ 正确做法
```go
// ✅ 使用布局管理器控制尺寸
content := container.NewBorder(
    top,
    nil, nil, nil,
    child, // 作为 Center，自动扩展
)

// ✅ 或者使用 SetMinRowsVisible (Entry/TextArea)
entry.SetMinRowsVisible(20)

// ✅ 或者对最外层容器 Resize
content.Resize(fyne.NewSize(1024, 768))
```

### 2.3 问题：控件挤成一团/分布不均

**症状：** 多个控件挤在一起或空间分配不合理

**原因：**
- 使用 VBox/HBox 无法自动分配空间
- 没有使用 Grid 或 Border 的 Center 扩展特性

**解决方案：**

```go
// ✅ 使用 Grid 等分空间
grid := container.NewGridWithColumns(3,
    widget1,
    widget2,
    widget3,
) // 每列等宽

// ✅ 使用 Border + Spacer
border := container.NewBorder(
    nil, nil,
    widget1,        // Left
    layout.NewSpacer(), // Right - 推到右边
    centerWidget,
)

// ✅ 使用 Grid 指定比例
grid := container.NewGridWithColumns(2,
    container.NewGridWithRows(2, widget1, widget2), // 左侧 2 行
    widget3, // 右侧 1 行
)
```

---

## 3. 已实现页面的布局模式

### 3.1 Connections 页面

**布局结构：**
```
Border
├─ Top: Toolbar (Add, Delete, Test, Refresh, Set Default)
├─ Left: Groups (MySQL, Oracle, PostgreSQL, SQL Server)
└─ Center: Scroll(List of connections)
```

**关键代码：**
```go
content := container.NewBorder(
    toolbar,               // Top
    nil,                   // Bottom
    groupList,             // Left
    nil,                   // Right
    container.NewScroll(connectionList), // Center - 自动扩展
)
```

**特点：**
- 左侧分组列表固定宽度
- 右侧连接列表自动扩展
- 单击选中，双击编辑

### 3.2 Templates 页面

**布局结构：**
```
Border
├─ Top: Toolbar (Add, Delete, Refresh, Set Default)
├─ Left: DB Type Selector (MySQL, Oracle, PostgreSQL, SQL Server)
└─ Center: Scroll(List of templates)
```

**关键代码：**
```go
content := container.NewBorder(
    toolbar,               // Top
    nil,                   // Bottom
    dbTypeSelect,          // Left
    nil,                   // Right
    container.NewScroll(templateList), // Center - 自动扩展
)
```

**特点：**
- 左侧数据库类型选择器
- 右侧模板列表自动扩展
- 支持内置模板和自定义模板

### 3.3 Tasks & Monitor 页面 ⭐ **最复杂**

**布局结构：**
```
VBox
├─ Card: Task Configuration (Form)
├─ Card: Monitor Metrics
│  ├─ Status + Progress Bar
│  └─ Metrics Grid (TPS, QPS, Latency, Threads, Errors)
└─ VBox: Real-time Log (MultiLineEntry)
```

**关键改进：**
```go
// 执行前连接测试
if p.connUC != nil {
    connName := p.connSelect.Selected
    conn, ok := p.connections[connName]

    // 静默测试连接（不弹窗）
    testResult, err := p.connUC.TestConnection(context.Background(), conn.GetID())
    if err != nil || !testResult.Success {
        // 失败时弹窗并终止
        dialog.ShowError(...)
        return
    }
    // 成功时只记录日志
    slog.Info("Tasks: Connection test successful")
}
```

**特点：**
- 集成任务配置和监控到一个页面
- 实时显示 TPS、QPS、延迟、错误率
- 执行前自动测试连接（失败弹窗，成功静默）
- 实时日志输出

### 3.4 History 页面

**布局结构：**
```
Border
├─ Top: Toolbar (Refresh, Export, View Details)
├─ Center: Scroll(List of history records)
└─ Bottom: Detail View (Card with metrics)
```

**关键代码：**
```go
// Tab 切换时自动刷新
tabs.OnSelected = func(tab *container.TabItem) {
    if tab.Text == "History" {
        historyPage.Refresh()
    }
}
```

**特点：**
- 单击选中记录
- 双击查看详情
- Tab 切换时自动刷新数据
- 支持多格式导出

### 3.5 Comparison 页面 ⭐ **最终布局**

**布局结构（最终版本）：**
```
GridWithRows(2) - 上下各 50%
├─ Row 1: Selection Area (Border)
│  ├─ Top: filterForm (Search + Group By)
│  └─ Center: listScroll (自动扩展)
└─ Row 2: Results Area (Border)
   ├─ Top: VBox(toolbar, Separator, Label)
   └─ Center: resultsScroll (自动扩展)
```

**关键代码：**
```go
// 上半部分：选择区域
selectionArea := container.NewBorder(
    filterForm,    // Top
    nil,           // Bottom
    nil,           // Left
    nil,           // Right
    listScroll,    // Center - 自动扩展
)

// 下半部分：结果区域
resultsArea := container.NewBorder(
    container.NewVBox(toolbar, widget.NewSeparator(), resultsLabel), // Top
    nil,           // Bottom
    nil,           // Left
    nil,           // Right
    resultsScroll, // Center - 自动扩展
)

// 整体：2行Grid，上下各50%
content := container.NewGridWithRows(2,
    selectionArea,
    resultsArea,
)
```

**特点：**
- Record List 向下扩展，显示多条记录
- Comparison Results 向下扩展，显示完整报告
- `resultsText.SetMinRowsVisible(30)` 确保有足够空间
- Tab 切换时自动刷新数据

---

## 4. UI 组件尺寸控制规范

### 4.1 TextEntry / MultiLineEntry

**设置最小行数：**
```go
entry := widget.NewMultiLineEntry()
entry.SetMinRowsVisible(20) // 至少显示 20 行
```

**常用行数设置：**
- 小型文本框：5-8 行
- 中型文本框：10-15 行
- 大型文本框（结果显示）：20-30 行
- 实时日志：30-60 行

### 4.2 List / Select

**List 尺寸：**
```go
// ❌ 不要直接 Resize list
list.Resize(...) // 无效

// ✅ 使用 Scroll 包装，放在可扩展位置
listScroll := container.NewScroll(list)
content := container.NewBorder(
    top, nil, nil, nil,
    listScroll, // Center 自动扩展
)
```

**Select 选项：**
```go
select := widget.NewSelect(options, onSelected)
select.SetSelected(option) // 设置默认选中
```

### 4.3 Card

**Card 自动适应内容：**
```go
card := widget.NewCard("Title", "Subtitle", content)
// Card 尺寸由 content 决定

// 让 Card 填充空间
content := container.NewBorder(
    nil, nil, nil, nil,
    card, // 作为 Center，自动扩展
)
```

---

## 5. 性能优化最佳实践

### 5.1 避免在 Goroutine 中直接操作 UI

**❌ 错误：**
```go
go func() {
    result := doHeavyWork()
    page.resultsText.SetText(result) // ❌ 在 goroutine 中操作 UI
}()
```

**✅ 正确：使用 Channel 传递结果**
```go
resultChan := make(chan *Result, 1)
errorChan := make(chan error, 1)

go func() {
    result, err := doHeavyWork()
    if err != nil {
        errorChan <- err
        return
    }
    resultChan <- result
}()

go func() {
    select {
    case result := <-resultChan:
        page.displayResults(result) // ✅ 在主 goroutine 中更新
    case err := <-errorChan:
        dialog.ShowError(err, page.win)
    }
}()
```

### 5.2 使用 Channel 模式避免阻塞

**标准模式：**
```go
// 创建缓冲 channel
resultChan := make(chan *Result, 1)
errorChan := make(chan error, 1)

// 在 goroutine 中执行
go func() {
    result, err := p.usecase.DoWork(ctx)
    if err != nil {
        errorChan <- err
        return
    }
    resultChan <- result
}()

// 在后台监听并更新UI
go func() {
    select {
    case result := <-resultChan:
        p.displayResults(result)
    case err := <-errorChan:
        dialog.ShowError(err, p.win)
    }
}()
```

### 5.3 减少不必要的 UI 刷新

**❌ 频繁刷新：**
```go
for i := 0; i < 1000; i++ {
    updateUI(i)
    time.Sleep(10 * time.Millisecond) // ❌ 太频繁
}
```

**✅ 批量更新：**
```go
ticker := time.NewTicker(500 * time.Millisecond) // 每 500ms 更新一次
defer ticker.Stop()

for {
    select {
    case <-dataChan:
        // 收集数据
    case <-ticker.C:
        // 定期刷新 UI
        page.updateMetrics()
    }
}
```

---

## 6. 检查清单

### 6.1 布局检查清单

在提交 UI 代码前，必须检查：

**基本布局：**
- [ ] 使用了合适的布局容器（Border/Grid/VBox）
- [ ] 需要扩展的组件放在了 Border 的 Center 或使用了 Grid
- [ ] Scroll 容器没有被 VBox/HBox 限制
- [ ] Entry/TextArea 设置了 SetMinRowsVisible()

**空间分配：**
- [ ] 上下分区使用了 `NewGridWithRows(2, ...)`
- [ ] 左右分区使用了 `NewGridWithColumns(2, ...)`
- [ ] Toolbar 使用了 VBox/HBox
- [ ] 没有过度嵌套容器（≤ 3 层）

**响应式设计：**
- [ ] 组件能够随窗口大小自动调整
- [ ] 列表/文本区域可以滚动
- [ ] 关键信息不会被遮挡

### 6.2 性能检查清单

- [ ] 没有在 goroutine 中直接操作 UI
- [ ] 使用了 channel 模式传递结果
- [ ] UI 更新频率合理（≤ 2 次/秒）
- [ ] 没有频繁的 Refresh() 调用
- [ ] 长时间操作在 goroutine 中执行

### 6.3 功能检查清单

- [ ] 所有按钮都有日志记录
- [ ] 错误情况有对话框提示
- [ ] 表单验证有日志记录
- [ ] 操作结果有反馈（成功/失败）
- [ ] 连接测试在执行前进行（Task Monitor）

---

## 7. 快速参考

### 7.1 布局选择决策树

```
需要等分空间？
├─ 是 → 使用 Grid (GridWithRows/GridWithColumns)
└─ 否 → 需要某组件自动扩展？
    ├─ 是 → 使用 Border (组件放在 Center)
    └─ 否 → 使用 VBox/HBox (固定排列)
```

### 7.2 常见模式速查

| 场景 | 布局 | 关键代码 |
|------|------|----------|
| 上下分区，下部自适应 | Border | `NewBorder(top, nil, nil, nil, scroll)` |
| 上下等分 | Grid | `NewGridWithRows(2, top, bottom)` |
| 工具栏 | HBox | `NewHBox(btn1, btn2, btn3)` |
| 表单 | VBox/Form | `NewForm(item1, item2)` |
| 列表 | Border+Scroll | `NewBorder(toolbar, nil, nil, nil, NewScroll(list))` |
| 实时日志 | Border+Scroll | `NewBorder(nil, nil, nil, nil, NewScroll(logEntry))` |

### 7.3 尺寸设置速查

| 组件 | 设置方法 | 示例 |
|------|----------|------|
| Entry/TextArea | SetMinRowsVisible() | `entry.SetMinRowsVisible(20)` |
| Select | SetSelected() | `select.SetSelected("选项")` |
| 窗口 | Resize() | `win.Resize(fyne.NewSize(1024, 768))` |
| 自定义对话框 | Resize() | `dlg.Resize(fyne.NewSize(500, 700))` |

---

## 8. 附录：完整示例

### 8.1 Comparison 页面完整布局（最终版本）

```go
func NewResultComparisonPage(win fyne.Window, comparisonUC *usecase.ComparisonUseCase) (*ResultComparisonPage, fyne.CanvasObject) {
    page := &ResultComparisonPage{
        win:          win,
        comparisonUC: comparisonUC,
        selectedMap:  make(map[string]bool),
        ctx:          context.Background(),
    }

    page.loadRecords()

    // Group By selector
    page.groupBySelect = widget.NewSelect([]string{
        "Threads", "Database Type", "Template Name", "Date",
    }, func(selected string) {
        page.onGroupByChange(selected)
    })
    page.groupBySelect.SetSelected("Threads")

    // Toolbar
    btnRefresh := widget.NewButton("🔄 Refresh", func() {
        page.loadRecords()
    })
    btnCompare := widget.NewButton("📊 Compare Selected", func() {
        page.onCompare()
    })
    btnExport := widget.NewButton("💾 Export Report", func() {
        page.onExportReport()
    })
    btnClear := widget.NewButton("🗑️ Clear", func() {
        page.resultsText.SetText("")
    })
    toolbar := container.NewHBox(btnRefresh, btnCompare, btnExport, btnClear)

    // Search and Group By
    searchEntry := widget.NewEntry()
    searchEntry.SetPlaceHolder("Search: MySQL, 8 threads, oltp...")
    searchEntry.OnChanged = func(text string) {
        page.filterRecords(text)
    }

    filterForm := widget.NewForm(
        widget.NewFormItem("Search Records", searchEntry),
        widget.NewFormItem("Group By", page.groupBySelect),
    )

    // Record list with checkboxes
    page.list = widget.NewList(...)

    // ⭐ 关键：使用 Border 布局让内容自动扩展
    listScroll := container.NewScroll(page.list)

    // ⭐ 上半部分：使用 Border 让 list 自动扩展
    selectionArea := container.NewBorder(
        filterForm,    // Top
        nil,           // Bottom
        nil,           // Left
        nil,           // Right
        listScroll,    // Center - 自动扩展填充空间
    )

    // Results text area
    page.resultsText = widget.NewMultiLineEntry()
    page.resultsText.SetText("Select 2 or more records...")
    page.resultsText.SetMinRowsVisible(30) // ⭐ 设置最小行数

    // ⭐ 下半部分：让 resultsScroll 直接作为 Center 扩展
    resultsLabel := widget.NewLabel("Comparison Results:")
    resultsScroll := container.NewScroll(page.resultsText)

    resultsArea := container.NewBorder(
        container.NewVBox(toolbar, widget.NewSeparator(), resultsLabel), // Top
        nil,           // Bottom
        nil,           // Left
        nil,           // Right
        resultsScroll, // Center - 直接让 scroll 自动扩展
    )

    // ⭐ 使用 2 行 Grid 布局，上下各占 50% 空间
    content := container.NewGridWithRows(2,
        selectionArea,
        resultsArea,
    )

    finalContent := widget.NewCard("Record Selection", "", content)

    return page, finalContent
}
```

---

**文档维护：**
本文档基于实际开发经验总结，如有新的布局模式或问题解决方案，请及时更新本文档。

**相关文档：**
- `docs/gui-development-guide.md` - GUI 开发基础规范
- `docs/USER_GUIDE.md` - 用户使用指南
- `CLAUDE.md` - AI 协作规范
- `constitution.md` - 项目宪法
