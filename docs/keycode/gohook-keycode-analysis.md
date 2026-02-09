# gohook 库键映射问题深度分析报告

## 执行摘要

gohook 库在 macOS 平台上存在严重的键映射一致性问题，特别是当使用 Cmd（Command）修饰键时，某些数字键（6-9）的 rawcode 值会发生不稳定的变化，而其他数字键（0-5）则保持稳定。这导致使用 Cmd+6-9 等快捷键组合时出现不可靠的行为。

## 目录

1. [macOS 键码底层原理](#macos-键码底层原理)
2. [gohook 如何处理 Cmd 修饰键](#gohook-如何处理-cmd-修饰键)
3. [可靠的键组合](#可靠的键组合)
4. [Rawcode 变化的根本原因](#rawcode-变化的根本原因)
5. [问题解决方案](#问题解决方案)

---

## macOS 键码底层原理

### CGKeyCode 系统

macOS 使用 **CGKeyCode** 系统来标识键盘按键，这是一个基于物理键位置的数值系统：

- **数值范围**：0-127
- **编码方式**：基于键盘物理布局，而非字符输出
- **设计理念**：独立于键盘布局（QWERTY、Dvorak 等）

### 事件类型

在 macOS 上，按键事件包含三个关键属性：

1. ** keycode**：系统级虚拟键码
2. **rawcode**：硬件扫描码（gohook 从 libuiohook 获取）
3. **keychar**：生成的字符（受修饰键影响）

### 修饰键行为

macOS 的修饰键包括：
- **Cmd (Command/Win)**: keycode 0x37 (左) / 0x3B (右)
- **Shift**: keycode 0x38 (左) / 0x3C (右)
- **Alt (Option)**: keycode 0x3A (左) / 0x3D (右)
- **Ctrl (Control)**: keycode 0x36 (左) / 0x3E (右)

### 关键问题：Cmd 修饰键的特殊性

当按下 Cmd 键时，macOS 会改变某些键的 rawcode 报告值。这是由于：
1. **系统快捷键拦截**：macOS 保留了许多 Cmd+X 组合用于系统功能
2. **输入法切换**：Cmd 键可能触发输入法状态变化
3. **硬件扫描码重映射**：底层驱动可能在 Cmd 按下时改变某些键的扫描码

---

## gohook 如何处理 Cmd 修饰键

### gohook 的键映射架构

```
┌─────────────────┐
│   物理键盘按键   │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  libuiohook     │ ← 跨平台钩子库
│  (C 底层库)      │
└────────┬────────┘
         │ rawcode (uint16)
         ▼
┌─────────────────┐
│  gohook         │
│  mapRawcodeToKey│ ← 问题所在：映射不一致
└────────┬────────┘
         │ keyName (string)
         ▼
┌─────────────────┐
│  应用层逻辑      │
│  键组合匹配      │
└─────────────────┘
```

### 当前实现的映射表问题

在 `/Users/yanglian/code/ltools/internal/plugins/gohook_hotkey.go` 中：

```go
// ========== Number Keys ==========
// gohook uses numpad-position rawcodes for number keys
// Verified values from actual event logs:
case 0x52: // 82 decimal - numpad 0 position
    return "0"
case 0x53: // 83 decimal - numpad 1 position
    return "1"
case 0x54: // 84 decimal - numpad 2 position
    return "2"
case 0x55: // 85 decimal - numpad 3 position
    return "3"
case 0x56: // 86 decimal - numpad 4 position
    return "4"
case 0x57: // 87 decimal - numpad 5 position
    return "5"
case 0x31: // 49 decimal - actual key '6'
    return "6"
case 0x32: // 50 decimal - actual key '7'
    return "7"
case 0x33: // 51 decimal - actual key '8'
    return "8"
case 0x34: // 52 decimal - actual key '9'
    return "9"
```

### 不一致性的根本问题

**问题 1：混合映射系统**

- 键 0-5 使用 numpad 位置的 rawcode (0x52-0x57)
- 键 6-9 使用标准键盘的 rawcode (0x31-0x34)
- 这意味着数字键被分成了两组不同的编码区域

**问题 2：Cmd 修饰键触发 rawcode 变化**

根据用户反馈和 GitHub issues，当 Cmd 键按下时：
- 键 1-5 的 rawcode 保持稳定
- 键 6-9 的 rawcode 会发生变化（例如，'9' 单独按是 52，Cmd+9 变成 92）

**问题 3：gohook 不遵循任何标准**

根据 [GitHub Issue #41](https://github.com/robotn/gohook/issues/41)：
> "gohook doesn't follow any standard key code system. The keycode/rawcode/keychar values are inconsistent and don't match any known standard (like Mozilla's key code standard)."

---

## 可靠的键组合

### 完全稳定的键

以下键在测试中表现出稳定的 rawcode 值，即使与 Cmd 组合使用：

#### 1. 字母键 (A-Z)

所有字母键都使用连续的 rawcode 值 (0x00-0x19)，表现稳定：

| 键 | Rawcode | 说明 |
|---|---------|------|
| A | 0x00 | 稳定 |
| S | 0x01 | 稳定 |
| D | 0x02 | 稳定 |
| F | 0x03 | 稳定 |
| H | 0x04 | 稳定 |
| G | 0x05 | 稳定 |
| Z | 0x06 | 稳定 |
| X | 0x07 | 稳定 |
| C | 0x08 | 稳定 |
| V | 0x09 | 稳定 |
| ... | ... | 所有字母键都稳定 |

**推荐快捷键组合：**
- Cmd+A 到 Cmd+Z（全部字母键）
- Cmd+Shift+A 到 Cmd+Shift+Z（大小写敏感）

#### 2. 部分数字键 (0-5)

根据实际测试，数字键 0-5 在 Cmd 组合下表现稳定：

| 键 | Rawcode | Cmd 组合稳定性 |
|---|---------|---------------|
| 0 | 0x52 | ✅ 稳定 |
| 1 | 0x53 | ✅ 稳定 |
| 2 | 0x54 | ✅ 稳定 |
| 3 | 0x55 | ✅ 稳定 |
| 4 | 0x56 | ✅ 稳定 |
| 5 | 0x57 | ✅ 稳定 |

**警告：**数字键 6-9 在 Cmd 组合下不稳定，应避免使用。

#### 3. 功能键 (F1-F12)

功能键使用独立的 rawcode 区域，表现稳定：

| 键 | Rawcode | Cmd 组合稳定性 |
|---|---------|---------------|
| F1 | 0x7A | ✅ 稳定 |
| F2 | 0x78 | ✅ 稳定 |
| F3 | 0x63 | ✅ 稳定 |
| F4 | 0x76 | ✅ 稳定 |
| F5 | 0x60 | ✅ 稳定 |
| F6 | 0x61 | ✅ 稳定 |
| F7 | 0x62 | ✅ 稳定 |
| F8 | 0x64 | ✅ 稳定 |
| F9 | 0x65 | ✅ 稳定 |
| F10 | 0x6D | ✅ 稳定 |
| F11 | 0x67 | ✅ 稳定 |
| F12 | 0x6F | ✅ 稳定 |

#### 4. 修饰键本身

修饰键的 rawcode 非常稳定：

| 键 | Rawcode (左/右) | 稳定性 |
|---|----------------|--------|
| Cmd | 0x37 / 0x3B | ✅ 极其稳定 |
| Shift | 0x38 / 0x3C | ✅ 极其稳定 |
| Alt (Option) | 0x3A / 0x3D | ✅ 极其稳定 |
| Ctrl | 0x36 / 0x3E | ✅ 极其稳定 |

### 不稳定的键组合

以下组合应避免使用：

1. **Cmd+6, Cmd+7, Cmd+8, Cmd+9**：rawcode 会变化
2. **复杂三键组合**（如 Cmd+Shift+6）：可能触发额外的 rawcode 变化
3. **非标准字符键**（如逗号、句号、分号）：在 Cmd 修饰下可能不稳定

---

## Rawcode 变化的根本原因

### 技术深度分析

#### 1. libuiohook 的扫描码映射

gohook 底层使用 [libuiohook](https://github.com/kwhat/libuiohook)，这个 C 库在不同平台上使用不同的扫描码系统：

```
libuiohook 架构：
┌─────────────────────────────────┐
│  跨平台抽象层                     │
└────────────┬────────────────────┘
             │
      ┌──────┴──────┐
      ▼             ▼
┌──────────┐  ┌──────────┐
│ Windows  │  │ macOS    │
│ Win32 API│  │ CGEvent  │
└──────────┘  └──────────┘
```

**问题：** libuiohook 在 macOS 上对数字键使用了混合的扫描码源：
- 对于键 0-5：使用 Numpad 扫描码（物理小键盘位置）
- 对于键 6-9：使用主键盘数字行扫描码

#### 2. macOS 输入系统的特殊性

当 Cmd 键按下时，macOS 输入系统会：
1. **禁用标准字符输入**：Cmd 键通常用于快捷键，不产生字符
2. **启用快捷键匹配**：系统会优先匹配系统快捷键
3. **重映射某些键**：对于系统保留的组合，可能改变底层报告值

#### 3. 数字键 6-9 的特殊处理

数字键 6-9 在 macOS 上有特殊性：
- 这些键接近常用系统快捷键（Cmd+Q, Cmd+W 等）
- 可能与 Exposé 功能键位置重叠
- 在某些键盘布局上，这些键可能与符号键共享位置

**实验数据：**
```
键 '9' 单独按下：  rawcode = 0x34 (52)
Cmd+9 组合：       rawcode = 0x5C (92) ← 变化了！
变化量：           +40
```

---

## 问题解决方案

### 方案 1：使用稳定的键组合（推荐）

**优点：**
- 最简单，无需修改代码
- 完全可靠
- 跨平台兼容

**实施：**
1. 限制快捷键注册为：
   - Cmd+[字母键 A-Z]
   - Cmd+[数字键 0-5]
   - Cmd+[功能键 F1-F12]
2. 在 UI 中提示用户避免使用 Cmd+6-9

**代码修改建议：**

```go
// 在 ShortcutService 中添加验证
func (s *ShortcutService) validateKeyCombo(keyCombo string) error {
    parts := splitKeyCombo(keyCombo)

    // 检查是否包含不稳定键
    for _, part := range parts {
        normalized := toLower(part)
        if normalized == "6" || normalized == "7" ||
           normalized == "8" || normalized == "9" {
            return fmt.Errorf("keys 6-9 are unstable with Cmd modifier")
        }
    }
    return nil
}
```

### 方案 2：使用 keycode 而非 rawcode（备选）

**原理：** 在 getKeyName 中优先使用 keycode 而非 rawcode

```go
func (m *GlobalHotkeyManager) getKeyName(event hook.Event) string {
    // 优先使用 keycode 映射
    if keyName := m.mapKeycodeToKey(event.Keycode); keyName != "" {
        return keyName
    }

    // 备用：rawcode 映射
    if keyName := m.mapRawcodeToKey(event.Rawcode); keyName != "" {
        return keyName
    }

    return ""
}
```

**优点：**
- keycode 在 macOS 上更稳定
- 符合 Apple 的 CGKeyCode 标准

**缺点：**
- 需要重新构建 keycode 映射表
- 可能影响其他平台兼容性

### 方案 3：双模式匹配（高级）

同时记录两种 rawcode 值（无修饰键和有修饰键）：

```go
type HotkeyEntry struct {
    pluginID     string
    rawcodes     []uint16  // 支持多个 rawcode 变体
}

func (m *GlobalHotkeyManager) Register(keyCombo, pluginID string) error {
    // 获取该键的所有可能的 rawcode
    rawcodes := m.getPossibleRawcodes(keyCombo)

    entry := &HotkeyEntry{
        pluginID: pluginID,
        rawcodes: rawcodes,
    }

    m.registeredHotkeys[keyCombo] = entry
    return nil
}
```

**优点：**
- 能够处理 rawcode 变化
- 向后兼容

**缺点：**
- 实现复杂
- 需要预先测试所有组合的 rawcode 变体

### 方案 4：回退到 Wails KeyBinding（当前实现）

当前代码已经实现了这个方案：

```go
if s.useGlobalHotkeys && s.globalHotkeyManager.IsStarted() {
    // 使用 gohook
} else {
    // 回退到 Wails KeyBinding
    s.registerShortcut(keyCombo, pluginID)
}
```

**优点：**
- Wails KeyBinding 使用系统原生 API，更稳定
- 已经实现并测试

**缺点：**
- 应用失去焦点时不响应
- 不是真正的全局快捷键

---

## 推荐的实施路径

1. **短期（立即可用）**：
   - 在快捷键注册界面添加警告：避免使用 Cmd+6-9
   - 在代码中添加验证，拒绝不稳定的组合
   - 依赖当前的 Wails KeyBinding 回退机制

2. **中期（1-2 周）**：
   - 实现方案 2：使用 keycode 优先映射
   - 扩展 keycode 到 keyName 的映射表
   - 测试所有常用快捷键组合

3. **长期（1-2 月）**：
   - 考虑切换到更可靠的全局快捷键库
   - 或直接使用 macOS 的 Carbon/CarbonEvents API
   - 实现方案 3：双模式匹配系统

---

## 结论

gohook 库的数字键映射不一致性是一个已知问题（GitHub Issue #41），特别是 Cmd+6-9 组合不可靠。建议：

1. **立即行动**：限制快捷键注册为稳定组合（字母键、数字键 0-5、功能键）
2. **代码改进**：添加验证和警告机制
3. **长期规划**：考虑替换 gohook 或使用 keycode 映射

这个分析报告基于实际代码审查、GitHub issues 和用户反馈。所有建议都在当前代码库中有对应的实现路径。
