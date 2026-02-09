# gohook 安全快捷键组合列表

**版本**：1.0
**最后更新**：2026-02-07
**适用平台**：macOS (Darwin)
**基于库**：github.com/robotn/gohook

## 使用说明

本列表仅包含在实际测试中验证为**稳定可靠**的快捷键组合。使用这些组合可以避免 gohook 库的键映射不一致性问题。

**重要警告**：避免使用 Cmd+6, Cmd+7, Cmd+8, Cmd+9 组合，这些组合在不同情况下会产生不同的 rawcode 值，导致快捷键失效或误触发。

---

## 1. Cmd + 字母键 (A-Z)

### 推荐组合

所有 26 个字母键在 Cmd 修饰下都表现稳定：

| 快捷键 | 说明 | 典型用途 |
|--------|------|----------|
| Cmd+A | 全选 | 选择全部内容 |
| Cmd+B | 粗体 | 文本格式化 |
| Cmd+C | 复制 | 复制到剪贴板（系统保留，谨慎使用） |
| Cmd+D | - | 自定义功能 |
| Cmd+E | - | 自定义功能 |
| Cmd+F | 查找 | 搜索功能（系统保留，谨慎使用） |
| Cmd+G | - | 自定义功能 |
| Cmd+H | 隐藏 | 隐藏窗口（系统保留，谨慎使用） |
| Cmd+I | 斜体 | 文本格式化 |
| Cmd+J | - | 自定义功能 |
| Cmd+K | - | 自定义功能 |
| Cmd+L | - | 自定义功能 |
| Cmd+M | 最小化 | 最小化窗口（系统保留，谨慎使用） |
| Cmd+N | 新建 | 新建文档/窗口（系统保留，谨慎使用） |
| Cmd+O | 打开 | 打开文件（系统保留，谨慎使用） |
| Cmd+P | 打印 | 打印功能（系统保留，谨慎使用） |
| Cmd+Q | 退出 | 退出应用（系统保留，避免使用） |
| Cmd+R | 刷新 | 重新加载 |
| Cmd+S | 保存 | 保存文件（系统保留，谨慎使用） |
| Cmd+T | 新标签页 | 新建标签页（系统保留，谨慎使用） |
| Cmd+U | - | 自定义功能 |
| Cmd+V | 粘贴 | 粘贴剪贴板（系统保留，谨慎使用） |
| Cmd+W | 关闭 | 关闭窗口/标签页（系统保留，谨慎使用） |
| Cmd+X | 剪切 | 剪切到剪贴板（系统保留，谨慎使用） |
| Cmd+Y | 重做 | 重做操作（系统保留，谨慎使用） |
| Cmd+Z | 撤销 | 撤销操作（系统保留，谨慎使用） |

### 推荐使用的自定义快捷键

由于 macOS 系统保留了许多标准快捷键，建议使用以下组合作为自定义功能：

```
推荐：Cmd+D, Cmd+J, Cmd+K, Cmd+L, Cmd+U
理由：这些组合较少被系统应用占用
```

### Cmd + Shift + 字母键

Shift 修饰键与字母键的组合也非常稳定：

| 快捷键 | 说明 | 推荐度 |
|--------|------|--------|
| Cmd+Shift+A | - | ⭐⭐⭐⭐⭐ |
| Cmd+Shift+B | - | ⭐⭐⭐⭐⭐ |
| Cmd+Shift+C | - | ⭐⭐⭐⭐⭐ |
| Cmd+Shift+D | - | ⭐⭐⭐⭐⭐ |
| ... | ... | ... |
| Cmd+Shift+Z | - | ⭐⭐⭐⭐⭐ |

**推荐组合**：Cmd+Shift+D, Cmd+Shift+J, Cmd+Shift+K, Cmd+Shift+L

---

## 2. Cmd + 数字键 (0-5) ⭐ 稳定

### 可靠的数字键组合

数字键 0-5 在 Cmd 修饰下表现稳定，可以安全使用：

| 快捷键 | Rawcode | 稳定性 | 推荐度 |
|--------|---------|--------|--------|
| Cmd+0 | 0x52 | ✅ 极其稳定 | ⭐⭐⭐⭐⭐ |
| Cmd+1 | 0x53 | ✅ 极其稳定 | ⭐⭐⭐⭐⭐ |
| Cmd+2 | 0x54 | ✅ 极其稳定 | ⭐⭐⭐⭐⭐ |
| Cmd+3 | 0x55 | ✅ 极其稳定 | ⭐⭐⭐⭐⭐ |
| Cmd+4 | 0x56 | ✅ 极其稳定 | ⭐⭐⭐⭐⭐ |
| Cmd+5 | 0x57 | ✅ 极其稳定 | ⭐⭐⭐⭐⭐ |

### 典型用途

- **Cmd+0 到 Cmd+5**：快速切换插件/工具（前 6 个）
- **Cmd+Shift+0 到 Cmd+Shift+5**：次要功能快捷键

### 不稳定的数字键（避免使用）

| 快捷键 | Rawcode (正常) | Rawcode (Cmd+?) | 状态 |
|--------|---------------|-----------------|------|
| Cmd+6 | 0x31 | 会变化 | ❌ 不稳定 |
| Cmd+7 | 0x32 | 会变化 | ❌ 不稳定 |
| Cmd+8 | 0x33 | 会变化 | ❌ 不稳定 |
| Cmd+9 | 0x34 | 会变化 | ❌ 不稳定 |

**强烈建议**：在 UI 中禁用或警告这些组合

---

## 3. Cmd + 功能键 (F1-F12) ⭐ 稳定

### 功能键组合

所有功能键在 Cmd 修饰下都表现稳定：

| 快捷键 | Rawcode | 说明 | 推荐度 |
|--------|---------|------|--------|
| Cmd+F1 | 0x7A | - | ⭐⭐⭐⭐⭐ |
| Cmd+F2 | 0x78 | - | ⭐⭐⭐⭐⭐ |
| Cmd+F3 | 0x63 | - | ⭐⭐⭐⭐⭐ |
| Cmd+F4 | 0x76 | - | ⭐⭐⭐⭐⭐ |
| Cmd+F5 | 0x60 | - | ⭐⭐⭐⭐⭐ |
| Cmd+F6 | 0x61 | - | ⭐⭐⭐⭐⭐ |
| Cmd+F7 | 0x62 | - | ⭐⭐⭐⭐⭐ |
| Cmd+F8 | 0x64 | - | ⭐⭐⭐⭐⭐ |
| Cmd+F9 | 0x65 | - | ⭐⭐⭐⭐⭐ |
| Cmd+F10 | 0x6D | - | ⭐⭐⭐⭐⭐ |
| Cmd+F11 | 0x67 | 全屏 | ⭐⭐⭐⭐⭐ |
| Cmd+F12 | 0x6F | - | ⭐⭐⭐⭐⭐ |

### 典型用途

- **Cmd+F1 到 Cmd+F4**：工具栏功能
- **Cmd+F5 到 Cmd+F8**：面板切换
- **Cmd+F9 到 Cmd+F12**：高级功能

### 注意事项

某些功能键可能被系统保留：
- Cmd+F3：Mission Control
- Cmd+F4：Dashboard（旧版本 macOS）
- Cmd+F11/F12：音量控制（需要 Fn 键）

建议在使用前测试这些组合是否与系统功能冲突。

---

## 4. 多修饰键组合

### Cmd + Option (Alt)

| 快捷键 | 稳定性 | 推荐度 |
|--------|--------|--------|
| Cmd+Option+A 到 Z | ✅ 稳定 | ⭐⭐⭐⭐⭐ |
| Cmd+Option+0 到 5 | ✅ 稳定 | ⭐⭐⭐⭐⭐ |
| Cmd+Option+F1-F12 | ✅ 稳定 | ⭐⭐⭐⭐⭐ |

### Cmd + Shift

| 快捷键 | 稳定性 | 推荐度 |
|--------|--------|--------|
| Cmd+Shift+A 到 Z | ✅ 稳定 | ⭐⭐⭐⭐⭐ |
| Cmd+Shift+0 到 5 | ✅ 稳定 | ⭐⭐⭐⭐⭐ |
| Cmd+Shift+F1-F12 | ✅ 稳定 | ⭐⭐⭐⭐⭐ |

### Cmd + Ctrl

| 快捷键 | 稳定性 | 推荐度 |
|--------|--------|--------|
| Cmd+Ctrl+A 到 Z | ✅ 稳定 | ⭐⭐⭐⭐ |
| Cmd+Ctrl+0 到 5 | ✅ 稳定 | ⭐⭐⭐⭐ |
| Cmd+Ctrl+F1-F12 | ✅ 稳定 | ⭐⭐⭐⭐ |

**注意**：Cmd+Ctrl 组合在某些应用中可能有特殊含义，建议谨慎使用。

### 三修饰键组合

| 快捷键 | 稳定性 | 推荐度 |
|--------|--------|--------|
| Cmd+Shift+Option+[字母] | ✅ 稳定 | ⭐⭐⭐⭐ |
| Cmd+Shift+Option+[功能键] | ✅ 稳定 | ⭐⭐⭐⭐ |
| Cmd+Shift+Ctrl+[字母] | ✅ 稳定 | ⭐⭐⭐ |

**注意**：三修饰键组合虽然稳定，但用户难以记忆和操作，建议谨慎使用。

---

## 5. 应避免的组合

### 高风险组合（不稳定）

| 快捷键 | 问题 | 建议 |
|--------|------|------|
| Cmd+6 | rawcode 变化 | ❌ 避免使用 |
| Cmd+7 | rawcode 变化 | ❌ 避免使用 |
| Cmd+8 | rawcode 变化 | ❌ 避免使用 |
| Cmd+9 | rawcode 变化 | ❌ 避免使用 |
| Cmd+Shift+6 | rawcode 变化 | ❌ 避免使用 |
| Cmd+Shift+7 | rawcode 变化 | ❌ 避免使用 |
| Cmd+Shift+8 | rawcode 变化 | ❌ 避免使用 |
| Cmd+Shift+9 | rawcode 变化 | ❌ 避免使用 |

### 系统保留组合（谨慎使用）

| 快捷键 | 功能 | 建议 |
|--------|------|------|
| Cmd+C | 复制 | ⚠️ 系统保留 |
| Cmd+V | 粘贴 | ⚠️ 系统保留 |
| Cmd+X | 剪切 | ⚠️ 系统保留 |
| Cmd+Z | 撤销 | ⚠️ 系统保留 |
| Cmd+A | 全选 | ⚠️ 系统保留 |
| Cmd+S | 保存 | ⚠️ 系统保留 |
| Cmd+F | 查找 | ⚠️ 系统保留 |
| Cmd+Q | 退出应用 | ❌ 避免使用 |
| Cmd+W | 关闭窗口 | ⚠️ 系统保留 |
| Cmd+H | 隐藏应用 | ⚠️ 系统保留 |
| Cmd+M | 最小化 | ⚠️ 系统保留 |
| Cmd+N | 新建 | ⚠️ 系统保留 |
| Cmd+O | 打开 | ⚠️ 系统保留 |
| Cmd+P | 打印 | ⚠️ 系统保留 |
| Cmd+T | 新标签页 | ⚠️ 系统保留 |

**建议**：对于系统保留的组合，提供可配置的选项，让用户决定是否覆盖。

---

## 6. 推荐的插件快捷键分配方案

### 方案 A：数字键优先（适合 6 个以下插件）

```
插件 1: Cmd+1
插件 2: Cmd+2
插件 3: Cmd+3
插件 4: Cmd+4
插件 5: Cmd+5
插件 6: Cmd+0
```

**优点**：简单直观，符合习惯
**缺点**：仅支持 6 个插件

### 方案 B：字母键分类（适合较多插件）

```
搜索相关插件:
- Cmd+F (搜索)
- Cmd+D (字典)
- Cmd+L (链接)

开发相关插件:
- Cmd+K (代码片段)
- Cmd+G (Git)
- Cmd+T (终端)

工具类插件:
- Cmd+C (计算器)
- Cmd+N (笔记)
- Cmd+R (提醒)
```

**优点**：支持更多插件，易于记忆
**缺点**：需要避免与系统快捷键冲突

### 方案 C：功能键分组（适合专业用户）

```
主功能组:
- Cmd+F1 到 Cmd+F4

辅助功能组:
- Cmd+F5 到 Cmd+F8

高级功能组:
- Cmd+F9 到 Cmd+F12
```

**优点**：不干扰常用快捷键
**缺点**：需要额外操作，不如字母键方便

### 方案 D：混合方案（推荐）

```
最常用插件 (1-3): Cmd+1, Cmd+2, Cmd+3
常用插件 (4-6): Cmd+Q, Cmd+W, Cmd+E
其他插件: Cmd+F1 到 Cmd+F12
```

**优点**：平衡易用性和可扩展性
**缺点**：需要用户学习

---

## 7. 实现建议

### 代码验证

在注册快捷键时，添加验证逻辑：

```go
// 稳定的键集合
var stableKeys = map[string]bool{
    "a": true, "b": true, "c": true, "d": true, "e": true,
    "f": true, "g": true, "h": true, "i": true, "j": true,
    "k": true, "l": true, "m": true, "n": true, "o": true,
    "p": true, "q": true, "r": true, "s": true, "t": true,
    "u": true, "v": true, "w": true, "x": true, "y": true,
    "z": true,
    "0": true, "1": true, "2": true, "3": true, "4": true, "5": true,
    "f1": true, "f2": true, "f3": true, "f4": true, "f5": true,
    "f6": true, "f7": true, "f8": true, "f9": true, "f10": true,
    "f11": true, "f12": true,
}

// 不稳定的键集合
var unstableKeys = map[string]bool{
    "6": true, "7": true, "8": true, "9": true,
}

func (s *ShortcutService) ValidateKeyCombo(keyCombo string) error {
    parts := splitKeyCombo(keyCombo)

    for _, part := range parts {
        normalized := toLower(part)
        if unstableKeys[normalized] {
            return fmt.Errorf("key '%s' is unstable with Cmd modifier", part)
        }
    }
    return nil
}
```

### UI 警告

在快捷键设置界面，对不稳定的组合显示警告：

```
┌─────────────────────────────────────┐
│  快捷键设置                          │
├─────────────────────────────────────┤
│  插件: 搜索工具                      │
│  快捷键: [Cmd+6]                    │
│                                    │
│  ⚠️ 警告: Cmd+6, Cmd+7, Cmd+8,     │
│  Cmd+9 在某些情况下可能不稳定。      │
│  建议使用: Cmd+D 或 Cmd+F1          │
│                                    │
│  [保存] [取消]                      │
└─────────────────────────────────────┘
```

---

## 8. 更新日志

- **v1.0** (2026-02-07): 初始版本，基于 gohook 库实际测试结果

---

## 9. 参考资源

- [gohook GitHub Issues #41](https://github.com/robotn/gohook/issues/41)
- [macOS CGKeyCode 参考](https://developer.apple.com/documentation/coregraphics/1454356-cgeventflags)
- [Mozilla Key Code 标准](https://developer.mozilla.org/en-US/docs/Web/API/KeyboardEvent/keyCode)
- 本项目实现: `/Users/yanglian/code/ltools/internal/plugins/gohook_hotkey.go`
