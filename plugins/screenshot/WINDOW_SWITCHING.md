# 截图功能 - 窗口切换实现

## ✅ 已实现的功能

### 1. 点击截图按钮隐藏主窗口
当用户在截图工具页面点击"开始截图"按钮时：
- 主窗口隐藏
- 截图编辑器窗口显示
- 截图立即开始渲染

### 2. 截图编辑器显示
截图编辑器以独立窗口形式显示：
- 全屏显示（覆盖整个屏幕区域）
- 半透明遮罩效果
- 区域选择功能
- 标注工具

### 3. ESC 取消截图
在截图编辑器中按 ESC 键：
- 截图编辑器窗口关闭
- 主窗口重新显示
- 取消当前截图操作

## 🔄 完整交互流程

```
用户点击"开始截图"按钮
    ↓
ScreenshotService.Trigger()
    ↓
ScreenshotWindowService.StartCapture()
    ├─→ mainWindow.Hide()        # 隐藏主窗口
    ├─→ CaptureDisplay(0)        # 截取主显示器
    ├─→ showEditorWindow()       # 显示截图编辑器
    └─→ emit("started")          # 发送事件
    ↓
用户在截图编辑器中选择区域
    ↓
用户按 ESC 键
    ↓
ScreenshotSelector.onCancel()
    ↓
ScreenshotEditor.handleCancel()
    ↓
ScreenshotService.CancelCapture()
    ↓
ScreenshotWindowService.CloseEditor()
    ├─→ editorWindow.Close()      # 关闭编辑器
    └─→ mainWindow.Show()        # 显示主窗口
```

## 🎯 关键代码变更

### 后端 (Go)

#### 1. window_service.go
添加主窗口引用和显示/隐藏功能：

```go
type ScreenshotWindowService struct {
    app             *application.App
    plugin          *ScreenshotPlugin
    editorWindow    *application.WebviewWindow
    mainWindow      *application.WebviewWindow  // 新增
    // ...
}

// 新增方法
func (s *ScreenshotWindowService) SetMainWindow(window *application.WebviewWindow) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.mainWindow = window
}
```

#### 2. StartCapture 方法
在开始截图时隐藏主窗口：

```go
func (s *ScreenshotWindowService) StartCapture() (string, error) {
    // Hide main window before starting capture
    if s.mainWindow != nil {
        s.mainWindow.Hide()
    }
    // ... 截图逻辑 ...
}
```

#### 3. CloseEditor 方法
在关闭编辑器时显示主窗口：

```go
func (s *ScreenshotWindowService) CloseEditor() error {
    // ... 关闭编辑器 ...

    // Show main window again
    if s.mainWindow != nil {
        s.mainWindow.Show()
    }
    return nil
}
```

#### 4. service.go
修改 CancelCapture 使用窗口服务：

```go
func (s *ScreenshotService) CancelCapture() {
    if s.windowService != nil {
        s.windowService.CloseEditor()  // 自动显示主窗口
    } else {
        s.plugin.CloseEditorWindow()
    }
    // ...
}
```

#### 5. main.go
设置主窗口引用：

```go
// 创建主窗口后
mainWindow = app.Window.NewWithOptions(...)

// 设置主窗口引用
screenshotWindowService.SetMainWindow(mainWindow)
```

### 前端 (TypeScript/React)

#### 1. ScreenshotWidget.tsx
点击截图按钮触发功能：

```tsx
const handleStartCapture = useCallback(async () => {
    setError('');
    setIsCapturing(true);

    try {
        await ScreenshotService.Trigger();
        // 主窗口自动隐藏，编辑器自动显示
    } catch (err) {
        setError(String(err));
        setIsCapturing(false);
    }
}, []);
```

#### 2. ScreenshotEditor.tsx
ESC 键处理：

```tsx
// 键盘事件处理
const handleKeyPress = (e: KeyboardEvent) => {
    if (e.key === 'Escape') {
        handleCancel();  // 取消截图，显示主窗口
    }
};

const handleCancel = async () => {
    try {
        await ScreenshotService.CancelCapture();
        handleClose();
    } catch (error) {
        console.error('取消失败:', error);
    }
};
```

#### 3. ScreenshotSelector.tsx
区域选择阶段的 ESC 处理：

```tsx
const handleKeyDown = (e: KeyboardEvent) => {
    if (e.key === 'Escape') {
        e.preventDefault();
        onCancel();  // 调用父组件的取消处理
    }
};
```

## 🧪 测试方法

### 测试步骤 1: 点击按钮截图
1. 运行应用：`task dev`
2. 导航到截图工具页面
3. 点击"开始截图"按钮
4. **预期结果**：
   - 主窗口立即隐藏
   - 截图编辑器窗口显示
   - 看到屏幕截图和半透明遮罩

### 测试步骤 2: ESC 取消
1. 在截图编辑器窗口中
2. 按 ESC 键
3. **预期结果**：
   - 截图编辑器窗口关闭
   - 主窗口重新显示
   - 回到截图工具页面

### 测试步骤 3: 全局快捷键
1. 在主窗口可见时
2. 按 `Cmd+Shift+6` (macOS) 或 `Ctrl+Shift+6` (Windows/Linux)
3. **预期结果**：
   - 主窗口隐藏
   - 截图编辑器显示
4. 按 ESC 取消
5. **预期结果**：
   - 编辑器关闭
   - 主窗口显示

## ⚠️ 注意事项

1. **窗口引用时机**: 主窗口引用必须在窗口创建后设置
2. **错误处理**: 截图失败时会自动显示主窗口
3. **并发安全**: 使用互斥锁保护窗口操作
4. **事件监听**: 前端正确监听键盘事件

## 🔧 调试信息

查看后端日志：
```
[ScreenshotWindowService] Starting capture...
[ScreenshotWindowService] Hiding main window...
[ScreenshotWindowService] Showing editor window...
[ScreenshotWindowService] Closing editor window...
[ScreenshotWindowService] Showing main window...
```

查看前端控制台：
```
取消操作: CancelCapture called
关闭编辑器: handleClose called
```

## 📋 相关文件

### 后端文件
- `plugins/screenshot/window_service.go` - 窗口管理服务
- `plugins/screenshot/service.go` - 前端接口服务
- `main.go` - 主窗口设置

### 前端文件
- `frontend/src/components/ScreenshotWidget.tsx` - 截图按钮组件
- `frontend/src/components/ScreenshotEditor.tsx` - 编辑器主组件
- `frontend/src/components/ScreenshotSelector.tsx` - 区域选择组件
- `frontend/bindings/ltools/plugins/screenshot/screenshotservice.ts` - 自动生成的绑定

## ✨ 功能特性

- ✅ 点击按钮隐藏主窗口
- ✅ 显示截图编辑器
- ✅ 实时渲染截图
- ✅ 半透明遮罩效果
- ✅ 区域选择功能
- ✅ ESC 取消截图
- ✅ 取消后显示主窗口
- ✅ 全局快捷键支持
- ✅ 错误处理和恢复
