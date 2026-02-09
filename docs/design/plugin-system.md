# LTools 插件系统架构设计

## 概述

设计一个类似 uTools 的 PC 工具箱软件，每个功能都是独立插件。

## 架构设计

### 1. 插件类型

```
┌─────────────────────────────────────────────────────────────┐
│                    LTools 主应用                              │
├─────────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │  内置插件     │  │  Go 插件      │  │  Web 插件     │      │
│  │  (Built-in)  │  │  (Native)    │  │  (Dynamic)   │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
│         │                 │                 │               │
│         └─────────────────┴─────────────────┘               │
│                           │                                 │
│                  ┌────────▼────────┐                        │
│                  │  插件管理器       │                        │
│                  │  PluginManager  │                        │
│                  └────────┬────────┘                        │
│                           │                                 │
└───────────────────────────┼─────────────────────────────────┘
                            │
                    ┌───────▼────────┐
                    │  插件注册表      │
                    │  PluginRegistry│
                    └────────────────┘
```

### 2. 插件元数据结构

```go
// PluginMetadata 插件元数据
type PluginMetadata struct {
    ID          string   `json:"id"`           // 唯一标识
    Name        string   `json:"name"`         // 显示名称
    Version     string   `json:"version"`      // 版本号
    Description string   `json:"description"`  // 描述
    Author      string   `json:"author"`       // 作者
    Type        PluginType `json:"type"`       // 插件类型
    Entry       string   `json:"entry"`        // 入口文件（Web插件）
    Icon        string   `json:"icon"`         // 图标
    Keywords    []string `json:"keywords"`     // 搜索关键词
    Permissions []string `json:"permissions"`  // 权限声明
}

type PluginType string
const (
    PluginTypeBuiltIn PluginType = "builtin"  // 内置插件
    PluginTypeNative  PluginType = "native"   // Go 编译插件
    PluginTypeWeb     PluginType = "web"      // Web 动态插件
)
```

### 3. 插件生命周期

```go
// Plugin 插件接口
type Plugin interface {
    // 获取插件元数据
    Metadata() PluginMetadata

    // 生命周期钩子
    OnLoad(ctx context.Context) error    // 加载时
    OnEnable(ctx context.Context) error  // 启用时
    OnDisable(ctx context.Context) error // 禁用时
    OnUnload(ctx context.Context) error  // 卸载时

    // 获取插件服务（如果有）
    Services() []application.Service
}

// WebPlugin Web 插件特有接口
type WebPlugin interface {
    Plugin
    // 渲染插件 UI
    Render(props map[string]interface{}) ReactElement
    // 处理插件消息
    HandleMessage(msg Message) error
}
```

### 4. 插件管理器

```go
// PluginManager 插件管理器服务
type PluginManager struct {
    registry  *PluginRegistry
    loader    PluginLoader
    plugins   map[string]Plugin
    events    *EventBus
}

// 功能：
// - 插件发现与加载
// - 插件启用/禁用
// - 插件搜索（支持关键词）
// - 插件依赖管理
// - 插件沙箱隔离
```

### 5. 插件目录结构

```
~/.ltools/
├── plugins/              # 插件安装目录
│   ├── builtin/          # 内置插件（编译进主程序）
│   ├── installed/        # 用户安装的插件
│   │   ├── calculator/   # 计算器插件示例
│   │   │   ├── plugin.json
│   │   │   ├── index.html
│   │   │   └── assets/
│   │   └── weather/
│   └── cache/            # 插件缓存
├── registry.json         # 插件注册表
└── settings.json         # 全局设置
```

### 6. 插件配置示例 (plugin.json)

```json
{
  "id": "com.ltools.plugin.calculator",
  "name": "计算器",
  "version": "1.0.0",
  "description": "简单的计算器工具",
  "author": "Your Name",
  "type": "web",
  "entry": "index.html",
  "icon": "icon.png",
  "keywords": ["计算", "计算器", "math", "calc"],
  "permissions": ["clipboard"],
  "features": {
    "searchable": true,
    "shortcut": "calc"
  }
}
```

### 7. 插件通信机制

```
┌──────────────┐         Event Bus          ┌──────────────┐
│   Plugin A   │ ────────────────────────▶ │   Plugin B   │
└──────────────┘                           └──────────────┘
       │                                          │
       │                                          │
       ▼                                          ▼
┌──────────────┐                         ┌──────────────┐
│ Plugin API   │                         │ Plugin API   │
│ (权限控制)    │                         │ (权限控制)    │
└──────────────┘                         └──────────────┘
```

### 8. 安全模型

```go
// 插件权限系统
type Permission struct {
    Name        string
    Description string
    Required    bool
}

var AvailablePermissions = []Permission{
    {Name: "filesystem", Description: "访问文件系统"},
    {Name: "network", Description: "网络访问"},
    {Name: "clipboard", Description: "读写剪贴板"},
    {Name: "shell", Description: "执行 Shell 命令"},
    {Name: "database", Description: "本地数据库"},
}

// 沙箱隔离
// - Web 插件运行在 iframe 中
// - Go 插件限制 API 调用
// - 权限运行时检查
```

## 实现阶段

### Phase 1: 基础框架
1. 实现 PluginMetadata 和核心接口
2. 实现插件管理器服务
3. 实现插件注册表
4. 实现事件总线

### Phase 2: 插件加载
1. 实现内置插件加载
2. 实现 Web 插件动态加载
3. 实现插件生命周期管理
4. 实现 API 权限控制

### Phase 3: UI 集成
1. 实现插件市场 UI
2. 实现插件搜索/安装/卸载
3. 实现插件设置页面
4. 实现快捷调用界面

### Phase 4: 高级特性
1. 插件依赖管理
2. 插件自动更新
3. 插件热重载（开发模式）
4. 插件间通信 API

## 示例插件

### 内置插件示例：时间日期

```go
// datetime/plugin.go
package datetime

type DateTimeService struct {}
type DateTimePlugin struct {
    service *DateTimeService
}

func (p *DateTimePlugin) Metadata() PluginMetadata {
    return PluginMetadata{
        ID:      "builtin.datetime",
        Name:    "当前时间",
        Type:    PluginTypeBuiltIn,
        Keywords: []string{"时间", "date", "time", "时钟"},
    }
}

func (p *DateTimePlugin) Services() []application.Service {
    return []application.Service{
        application.NewService(p.service),
    }
}
```

### Web 插件示例：计算器

```javascript
// calculator/index.html
<div id="calculator-app">
  <input type="text" id="display" readonly>
  <div class="buttons">...</div>
</div>

<script>
// 使用 LTools 提供的 API
window.ltools.on('enter', () => {
  console.log('计算器插件激活');
});

function calculate() {
  // 计算逻辑
}
</script>
```

## 开发工具

```bash
# 创建新插件
ltools plugin create my-plugin --type=web

# 构建插件
ltools plugin build ./my-plugin

# 安装插件
ltools plugin install ./my-plugin.ltp

# 列出已安装插件
ltools plugin list
```

## 技术栈总结

- **后端**: Go 1.25+ + Wails v3
- **前端**: React 18 + TypeScript 5
- **插件加载**: 动态 Import / iframe 沙箱
- **配置**: JSON
- **事件**: 自定义 Event Bus
