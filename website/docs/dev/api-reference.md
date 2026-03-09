# API 参考

LTools 提供的 API 参考文档。

## 核心 API

### Plugin 接口

所有插件必须实现的接口：

```go
type Plugin interface {
    // Metadata 返回插件元数据
    Metadata() *PluginMetadata

    // Init 初始化插件
    Init(app *application.App) error

    // ServiceStartup 应用启动时调用
    ServiceStartup(app *application.App) error

    // ServiceShutdown 应用关闭时调用
    ServiceShutdown(app *application.App) error

    // Enabled 返回插件是否启用
    Enabled() bool

    // SetEnabled 设置插件启用状态
    SetEnabled(enabled bool) error
}
```

### ViewLifecycle 接口

视图生命周期接口（可选）：

```go
type ViewLifecycle interface {
    // OnViewEnter 进入插件视图时调用
    OnViewEnter() error

    // OnViewLeave 离开插件视图时调用
    OnViewLeave() error
}
```

## 插件元数据

### PluginMetadata

```go
type PluginMetadata struct {
    ID          string       // 唯一标识符，如 "datetime.builtin"
    Name        string       // 显示名称
    Version     string       // 语义版本
    Author      string       // 作者
    Description string       // 描述
    Type        PluginType   // 类型
    State       PluginState  // 状态
    Permissions []Permission // 所需权限
    Keywords    []string     // 搜索关键词
    ShowInMenu  *bool        // 是否在菜单显示
    HasPage     *bool        // 是否有独立页面
}
```

### PluginType

```go
const (
    PluginTypeBuiltIn PluginType = "builtin"  // 内置插件
    PluginTypeWeb     PluginType = "web"      // Web 插件
    PluginTypeNative  PluginType = "native"   // 原生插件
)
```

### PluginState

```go
const (
    PluginStateInstalled PluginState = "installed"  // 已安装
    PluginStateEnabled   PluginState = "enabled"    // 已启用
    PluginStateDisabled  PluginState = "disabled"   // 已禁用
    PluginStateError     PluginState = "error"      // 错误
)
```

## 权限系统

### Permission

```go
const (
    PermissionFileSystem   Permission = "filesystem"   // 文件系统
    PermissionNetwork      Permission = "network"      // 网络
    PermissionClipboard    Permission = "clipboard"    // 剪贴板
    PermissionNotification Permission = "notification" // 通知
    PermissionProcess      Permission = "process"      // 进程
)
```

## 插件管理器 API

### Manager

```go
// Register 注册插件
func (m *Manager) Register(plugin Plugin)

// Unregister 注销插件
func (m *Manager) Unregister(pluginID string)

// Get 获取插件
func (m *Manager) Get(id string) Plugin

// List 列出所有插件
func (m *Manager) List() []Plugin

// Enable 启用插件
func (m *Manager) Enable(id string) error

// Disable 禁用插件
func (m *Manager) Disable(id string) error

// Search 搜索插件
func (m *Manager) Search(keywords ...string) []Plugin
```

## 事件系统 API

### 注册事件

```go
// 注册事件类型
application.RegisterEvent[string]("myplugin:data")
application.RegisterEvent[int]("myplugin:count")
application.RegisterEvent[MyStruct]("myplugin:struct")
```

### 发送事件

```go
// 发送事件
app.Event.Emit("myplugin:data", "hello")
app.Event.Emit("myplugin:count", 42)
app.Event.Emit("myplugin:struct", MyStruct{...})
```

### 前端监听

```typescript
import { Events } from '@wailsio/runtime'

// 监听事件
const unsubscribe = Events.On('myplugin:data', (ev: { data: string }) => {
    console.log(ev.data)
})

// 取消监听
unsubscribe()
```

## 前端 API

### 服务调用

```typescript
import { PluginService } from './bindings'

// 获取插件列表
const plugins = await PluginService.List()

// 启用插件
await PluginService.Enable('datetime.builtin')

// 禁用插件
await PluginService.Disable('datetime.builtin')

// 搜索插件
const results = await PluginService.Search('calc')
```

### 快捷键服务

```typescript
import { ShortcutService } from './bindings'

// 获取所有快捷键
const shortcuts = await ShortcutService.GetAllShortcuts()

// 设置快捷键
await ShortcutService.SetShortcut('cmd+shift+c', 'clipboard.builtin')

// 移除快捷键
await ShortcutService.RemoveShortcut('cmd+shift+c')

// 格式化快捷键
const formatted = ShortcutService.FormatShortcut('cmd+shift+c')
// 返回: "⌘⇧C"
```

### 搜索窗口服务

```typescript
import { SearchWindowService } from './bindings'

// 显示搜索窗口
await SearchWindowService.Show()

// 显示并填充查询
await SearchWindowService.ShowWithQuery('calc')

// 隐藏搜索窗口
await SearchWindowService.Hide()

// 切换搜索窗口
await SearchWindowService.Toggle()

// 搜索
const results = await SearchWindowService.Search('json')
```

## 自定义 Hooks

### usePlugins

```typescript
// 获取插件列表
const { plugins, loading, error } = usePlugins()

// 启用/禁用
const { enablePlugin, disablePlugin } = usePlugins()
```

### usePlugin

```typescript
// 获取单个插件
const { plugin, loading } = usePlugin('datetime.builtin')
```

### useGlobalShortcuts

```typescript
// 监听全局快捷键
useGlobalShortcuts((event) => {
    console.log('快捷键触发:', event)
})
```

## 数据存储 API

### 配置存储

```go
// 保存配置
func (p *MyPlugin) SaveConfig(config *Config) error {
    data, _ := json.Marshal(config)
    return os.WriteFile(p.configPath, data, 0644)
}

// 加载配置
func (p *MyPlugin) LoadConfig() (*Config, error) {
    data, err := os.ReadFile(p.configPath)
    if err != nil {
        return nil, err
    }
    var config Config
    json.Unmarshal(data, &config)
    return &config, nil
}
```

### 数据目录

```go
// 获取数据目录
dataDir := filepath.Join(os.UserHomeDir(), ".ltools", "myplugin")

// 确保目录存在
os.MkdirAll(dataDir, 0755)
```

## 常用工具函数

### BasePlugin 方法

```go
// 获取元数据
metadata := plugin.Metadata()

// 检查是否启用
enabled := plugin.Enabled()

// 设置启用状态
plugin.SetEnabled(true)
```

### 错误处理

```go
import "fmt"

// 包装错误
if err != nil {
    return fmt.Errorf("操作失败: %w", err)
}
```

### 日志记录

```go
import "log"

// 记录日志
log.Println("[MyPlugin] 操作开始")
log.Printf("[MyPlugin] 处理数据: %v", data)
```

## 类型定义

### 前端类型

```typescript
// 插件元数据
interface PluginMetadata {
    id: string
    name: string
    version: string
    author: string
    description: string
    type: 'builtin' | 'web' | 'native'
    state: 'installed' | 'enabled' | 'disabled' | 'error'
    permissions: Permission[]
    keywords: string[]
    showInMenu?: boolean
    hasPage?: boolean
}

// 权限类型
type Permission = 'filesystem' | 'network' | 'clipboard' | 'notification' | 'process'

// 快捷键信息
interface ShortcutInfo {
    keyCombo: string
    pluginID: string
    enabled: boolean
}

// 搜索结果
interface SearchResult {
    type: 'plugin' | 'app' | 'file'
    id: string
    title: string
    description: string
    icon: string
}
```

## 示例代码

### 创建完整插件

参考 [插件开发指南](./plugin-development) 中的完整示例。
