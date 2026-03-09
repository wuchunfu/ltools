# 代码规范

本文档定义了 LTools 项目的代码规范。

## Go 代码规范

### 命名约定

**包名**：
- 全小写，单个单词
- 不使用下划线或驼峰
- 例：`plugins`、`clipboard`、`datetime`

**文件名**：
- 小写，使用下划线分隔
- 例：`plugin_manager.go`、`search_window.go`

**类型名**：
- 导出类型：PascalCase（如 `PluginManager`）
- 私有类型：camelCase（如 `pluginRegistry`）

**函数/方法名**：
- 导出函数：PascalCase（如 `NewManager()`）
- 私有函数：camelCase（如 `loadConfig()`）
- 接口方法：动词或动词短语（如 `Get()`、`Set()`）

**常量**：
- 导出常量：PascalCase
- 私有常量：camelCase 或全大写
- 例：`PluginTypeBuiltIn`、`maxRetries`

### 代码组织

**包结构**：
```
myplugin/
├── myplugin.go      # 插件主文件
├── service.go       # 服务定义
├── types.go         # 类型定义
├── config.go        # 配置管理
└── myplugin_test.go # 测试文件
```

**导入顺序**：
```go
import (
    // 标准库
    "fmt"
    "os"

    // 第三方库
    "github.com/wailsapp/wails/v3/pkg/application"

    // 项目内部包
    "ltools/internal/plugins"
)
```

### 注释规范

**包注释**：
```go
// Package myplugin 提供某某功能。
// 详细描述...
package myplugin
```

**函数注释**：
```go
// DoSomething 执行某个操作
// 参数：
//   - input: 输入描述
// 返回：结果描述
func DoSomething(input string) (string, error) {
    // ...
}
```

**行注释**：
```go
// 创建新的管理器实例
manager := NewManager()

// TODO: 需要优化性能
// FIXME: 这个逻辑有问题
// NOTE: 重要说明
```

### 错误处理

**返回错误**：
```go
func (s *Service) DoWork() error {
    if err := s.validate(); err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }
    return nil
}
```

**检查错误**：
```go
result, err := someFunction()
if err != nil {
    return fmt.Errorf("operation failed: %w", err)
}
```

**不要忽略错误**：
```go
// ❌ 不好
result, _ := someFunction()

// ✅ 好
result, err := someFunction()
if err != nil {
    // 处理错误
}
```

### 并发编程

**Goroutine**：
```go
// 启动 goroutine
go func() {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("Recovered: %v", r)
        }
    }()
    // 执行任务
}()
```

**Channel**：
```go
// 有缓冲 channel
ch := make(chan int, 10)

// 关闭 channel
close(ch)

// 检查 channel 是否关闭
val, ok := <-ch
```

**Context**：
```go
func (s *Service) DoWork(ctx context.Context) error {
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
        // 执行工作
    }
    return nil
}
```

## TypeScript 代码规范

### 命名约定

**文件名**：
- 组件：PascalCase.tsx（如 `DateTimeWidget.tsx`）
- 工具：camelCase.ts（如 `pluginHelpers.ts`）
- Hooks：use 开头（如 `usePlugins.ts`）

**组件名**：
- PascalCase（如 `PluginCard`）
- 描述性名称

**变量/函数**：
- camelCase（如 `pluginList`、`loadPlugins`）
- 布尔值用 is/has 开头（如 `isLoading`、`hasError`）

**常量**：
- 全大写，下划线分隔（如 `MAX_COUNT`）
- 或 PascalCase（如 `DefaultConfig`）

**类型/接口**：
- PascalCase（如 `PluginMetadata`）
- 接口不加 I 前缀

### 组件结构

```typescript
// 1. 导入
import { useState, useEffect } from 'react'
import { PluginService } from '../bindings'

// 2. 类型定义
interface Props {
    pluginId: string
}

// 3. 组件定义
export function PluginCard({ pluginId }: Props) {
    // 4. Hooks
    const [plugin, setPlugin] = useState<PluginMetadata | null>(null)
    const [loading, setLoading] = useState(true)

    // 5. Effects
    useEffect(() => {
        loadPlugin()
    }, [pluginId])

    // 6. 辅助函数
    const loadPlugin = async () => {
        setLoading(true)
        const data = await PluginService.Get(pluginId)
        setPlugin(data)
        setLoading(false)
    }

    // 7. 事件处理
    const handleClick = () => {
        // ...
    }

    // 8. 渲染
    if (loading) return <div>Loading...</div>

    return (
        <div className="plugin-card">
            <h3>{plugin?.name}</h3>
            <button onClick={handleClick}>Enable</button>
        </div>
    )
}
```

### 类型定义

**接口**：
```typescript
interface PluginMetadata {
    id: string
    name: string
    version: string
    enabled: boolean
}
```

**类型别名**：
```typescript
type PluginType = 'builtin' | 'web' | 'native'
type Permission = 'filesystem' | 'network' | 'clipboard'
```

**泛型**：
```typescript
interface ApiResponse<T> {
    data: T
    error?: string
}
```

### 异步处理

**Async/Await**：
```typescript
// ✅ 好
const loadPlugins = async () => {
    try {
        const plugins = await PluginService.List()
        setPlugins(plugins)
    } catch (error) {
        console.error('Failed to load plugins:', error)
        setError(error.message)
    }
}

// ❌ 避免
PluginService.List().then(plugins => {
    setPlugins(plugins)
}).catch(error => {
    console.error(error)
})
```

### React Hooks 规范

**Hook 命名**：
- 以 `use` 开头（如 `usePlugins`）

**Hook 规则**：
- 只在顶层调用
- 只在 React 函数中调用

**依赖数组**：
```typescript
// ✅ 好 - 包含所有依赖
useEffect(() => {
    loadData(id)
}, [id])

// ❌ 不好 - 缺少依赖
useEffect(() => {
    loadData(id)
}, [])
```

## 通用规范

### 提交信息

使用约定式提交：

```
<type>(<scope>): <subject>

<body>

<footer>
```

**类型**：
- `feat`: 新功能
- `fix`: Bug 修复
- `docs`: 文档
- `style`: 格式
- `refactor`: 重构
- `test`: 测试
- `chore`: 构建/工具

**示例**：
```
feat(clipboard): 添加图片支持

- 支持从剪贴板读取图片
- 自动转换为 base64
- 添加图片预览

Closes #123
```

### 代码格式化

**Go**：
```bash
# 格式化代码
go fmt ./...

# 静态检查
go vet ./...
```

**TypeScript**：
```bash
# 使用 Prettier 格式化
npm run format

# 使用 ESLint 检查
npm run lint
```

### 测试规范

**Go 测试**：
```go
func TestMyFunction(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {"case1", "input1", "output1"},
        {"case2", "input2", "output2"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := MyFunction(tt.input)
            if result != tt.expected {
                t.Errorf("got %v, want %v", result, tt.expected)
            }
        })
    }
}
```

**React 测试**：
```typescript
describe('PluginCard', () => {
    it('should render plugin name', () => {
        render(<PluginCard pluginId="test" />)
        expect(screen.getByText('Test Plugin')).toBeInTheDocument()
    })
})
```

## 代码审查清单

提交代码前检查：

- [ ] 代码格式化
- [ ] 通过 lint 检查
- [ ] 添加必要注释
- [ ] 编写/更新测试
- [ ] 测试全部通过
- [ ] 更新相关文档
- [ ] 提交信息清晰
