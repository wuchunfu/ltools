# gohook 键映射参考文档

**版本**：1.0
**最后更新**：2026-02-07
**适用平台**：macOS (Darwin), Windows, Linux
**基于库**：github.com/robotn/gohook v1.x

## 文档概述

本文档提供了 gohook 库的完整键映射参考，包括所有可识别的按键及其对应的 rawcode/keycode 值。特别关注 macOS 平台的问题和解决方案。

---

## 目录

1. [快速参考](#快速参考)
2. [修饰键](#修饰键)
3. [功能键 (F1-F24)](#功能键-f1-f24)
4. [数字键 (0-9)](#数字键-0-9)
5. [字母键 (A-Z)](#字母键-a-z)
6. [符号键和标点](#符号键和标点)
7. [导航键](#导航键)
8. [编辑键](#编辑键)
9. [平台特定问题](#平台特定问题)
10. [已知问题和解决方案](#已知问题和解决方案)

---

## 快速参考

### 事件类型

gohook 提供以下事件类型：

| 事件常量 | 说明 | 典型用途 |
|----------|------|----------|
| `hook.KeyDown` | 键按下 | 检测快捷键触发 |
| `hook.KeyUp` | 键释放 | 清理按键状态 |
| `hook.KeyHold` | 键按住 | 长按检测 |
| `hook.MouseHook` | 鼠标事件 | 鼠标钩子 |

### 事件结构

```go
type Event struct {
    Kind     uint16     // 事件类型 (KeyDown, KeyUp, etc.)
    Keycode  uint16     // 虚拟键码 (平台相关)
    Rawcode  uint16     // 硬件扫描码
    Keychar  rune       // 生成的字符 (受修饰键影响)
}
```

### 优先级顺序

在 `mapRawcodeToKey()` 函数中，键名解析的优先级为：

1. **Rawcode 映射** (最可靠，推荐)
2. **Keycode 映射** (备选方案)
3. **Keychar 映射** (最后选择，最不可靠)

---

## 修饰键

### macOS 修饰键

| 键名 | Rawcode | Keycode | 左键 | 右键 | 稳定性 |
|------|---------|---------|------|------|--------|
| **Cmd (Command)** | 0x37/0x3B | 0x37/0x3B | 0x37 | 0x3B | ⭐⭐⭐⭐⭐ |
| **Shift** | 0x38/0x3C | 0x38/0x3C | 0x38 | 0x3C | ⭐⭐⭐⭐⭐ |
| **Alt (Option)** | 0x3A/0x3D | 0x3A/0x3D | 0x3A | 0x3D | ⭐⭐⭐⭐⭐ |
| **Ctrl (Control)** | 0x36/0x3E | 0x36/0x3E | 0x36 | 0x3E | ⭐⭐⭐⭐⭐ |
| **Fn** | - | - | - | - | ⚠️ 特殊处理 |
| **Caps Lock** | 0x39 | 0x39 | 0x39 | - | ⭐⭐⭐⭐ |

### Windows 修饰键

| 键名 | Rawcode | Keycode | 说明 |
|------|---------|---------|------|
| **Win (Meta)** | 0x5B/0x5C | 0x5B/0x5C | 左/右 Windows 键 |
| **Ctrl** | 0x1D | 0x11 | 左/右 Ctrl 键 |
| **Alt** | 0x38 | 0x12 | 左/右 Alt 键 |
| **Shift** | 0x2A/0x36 | 0x10 | 左/右 Shift 键 |

### Linux (X11) 修饰键

| 键名 | Rawcode | Keycode | 说明 |
|------|---------|---------|------|
| **Super (Meta)** | 0x85 | 0x85 | 左/右 Super 键 |
| **Ctrl** | 0x25 | 0x25 | 左/右 Ctrl 键 |
| **Alt** | 0x40 | 0x26 | 左/右 Alt 键 |
| **Shift** | 0x32 | 0x32 | 左/右 Shift 键 |

### 修饰键行为说明

#### Cmd (Command) 键的特殊性

macOS 的 Cmd 键有以下特点：

1. **系统保留**：许多 Cmd+X 组合被系统保留（如 Cmd+C 复制、Cmd+V 粘贴）
2. **字符抑制**：按下 Cmd 键时，通常不会产生可打印字符
3. **Rawcode 影响**：某些键的 rawcode 在 Cmd 修饰下会改变

```go
// 示例：检测 Cmd 键状态
func (m *GlobalHotkeyManager) isCmdPressed() bool {
    return m.keyState["cmd"] || m.keyState["meta"]
}
```

#### 修饰键组合匹配

```go
// 检查 Cmd+Shift+D 组合
func (m *GlobalHotkeyManager) matchesHotkey(keyCombo string) bool {
    parts := splitKeyCombo(keyCombo) // ["cmd", "shift", "d"]

    // 检查所有键是否都在按下状态
    for _, part := range parts {
        normalized := toLower(part)
        if !m.keyState[normalized] {
            return false
        }
    }
    return true
}
```

---

## 功能键 (F1-F24)

### macOS F键映射

| 功能键 | Rawcode | Keycode | Cmd 组合稳定性 | 说明 |
|--------|---------|---------|---------------|------|
| **F1** | 0x7A | 0x7A | ⭐⭐⭐⭐⭐ | 亮度减 (系统保留) |
| **F2** | 0x78 | 0x78 | ⭐⭐⭐⭐⭐ | 亮度增 (系统保留) |
| **F3** | 0x63 | 0x63 | ⭐⭐⭐⭐⭐ | Mission Control (系统保留) |
| **F4** | 0x76 | 0x76 | ⭐⭐⭐⭐⭐ | Launchpad (系统保留) |
| **F5** | 0x60 | 0x60 | ⭐⭐⭐⭐⭐ | 键盘灯减 (系统保留) |
| **F6** | 0x61 | 0x61 | ⭐⭐⭐⭐⭐ | 键盘灯增 (系统保留) |
| **F7** | 0x62 | 0x62 | ⭐⭐⭐⭐⭐ | 后退 (音乐控制) |
| **F8** | 0x64 | 0x64 | ⭐⭐⭐⭐⭐ | 播放/暂停 (音乐控制) |
| **F9** | 0x65 | 0x65 | ⭐⭐⭐⭐⭐ | 前进 (音乐控制) |
| **F10** | 0x6D | 0x6D | ⭐⭐⭐⭐⭐ | 静音 |
| **F11** | 0x67 | 0x67 | ⭐⭐⭐⭐⭐ | 音量减 |
| **F12** | 0x6F | 0x6F | ⭐⭐⭐⭐⭐ | 音量增 |
| **F13** | 0x69 | 0x69 | ⭐⭐⭐⭐⭐ | 打印屏幕 |
| **F14** | 0x6B | 0x6B | ⭐⭐⭐⭐⭐ | 滚动锁定 |
| **F15** | 0x71 | 0x71 | ⭐⭐⭐⭐⭐ | 暂停 |
| **F16** | 0x6A | 0x6A | ⭐⭐⭐⭐⭐ | - |
| **F17** | 0x40 | 0x40 | ⭐⭐⭐⭐⭐ | - |
| **F18** | 0x4F | 0x4F | ⭐⭐⭐⭐⭐ | - |
| **F19** | 0x50 | 0x50 | ⭐⭐⭐⭐⭐ | - |

### Windows F键映射

| 功能键 | Rawcode | Keycode | 说明 |
|--------|---------|---------|------|
| **F1-F12** | 0x3B-0x46 | 0x70-0x7B | 标准 F键 |
| **F13-F24** | 0x68-0x73 | 0x7C-0x87 | 扩展 F键 |

### 使用建议

```
✅ 推荐使用：
- F13-F19：几乎不被系统保留，完全可用于自定义功能
- F7-F9：虽然有媒体控制，但通常可以覆盖

⚠️ 谨慎使用：
- F1-F6：被系统功能保留
- F10-F12：音量控制键

❌ 避免使用：
- F3-F4：Mission Control 和 Launchpad
```

### Fn 键处理

在 Mac 上，按住 Fn 键可以改变 F键的行为：

```
F1 单独按下：    功能键 (亮度减)
Fn+F1 按下：    F1

F10 单独按下：   音量控制
Fn+F10 按下：   F10
```

在 gohook 中，需要根据 `Keycode` 来区分这些行为。

---

## 数字键 (0-9)

### ⚠️ 重要：数字键的不稳定性

在 macOS 上，数字键 0-9 存在严重的映射不一致性问题：

| 键 | 正常 Rawcode | Cmd+? Rawcode | 稳定性 | 说明 |
|---|--------------|---------------|--------|------|
| **0** | 0x52 (82) | 0x52 | ⭐⭐⭐⭐⭐ | 稳定 |
| **1** | 0x53 (83) | 0x53 | ⭐⭐⭐⭐⭐ | 稳定 |
| **2** | 0x54 (84) | 0x54 | ⭐⭐⭐⭐⭐ | 稳定 |
| **3** | 0x55 (85) | 0x55 | ⭐⭐⭐⭐⭐ | 稳定 |
| **4** | 0x56 (86) | 0x56 | ⭐⭐⭐⭐⭐ | 稳定 |
| **5** | 0x57 (87) | 0x57 | ⭐⭐⭐⭐⭐ | 稳定 |
| **6** | 0x31 (49) | 会变化 | ❌ | **不稳定** |
| **7** | 0x32 (50) | 会变化 | ❌ | **不稳定** |
| **8** | 0x33 (51) | 会变化 | ❌ | **不稳定** |
| **9** | 0x34 (52) | 会变化 | ❌ | **不稳定** |

### Rawcode 映射表

```go
// 当前实现中的映射（gohook_hotkey.go）
case 0x52: return "0"  // numpad 0 位置
case 0x53: return "1"  // numpad 1 位置
case 0x54: return "2"  // numpad 2 位置
case 0x55: return "3"  // numpad 3 位置
case 0x56: return "4"  // numpad 4 位置
case 0x57: return "5"  // numpad 5 位置
case 0x31: return "6"  // 实际键 '6'
case 0x32: return "7"  // 实际键 '7'
case 0x33: return "8"  // 实际键 '8'
case 0x34: return "9"  // 实际键 '9'
```

### 问题根源

数字键被分为两组不同的编码区域：

1. **0-5 键**：使用 numpad 位置的 rawcode (0x52-0x57)
2. **6-9 键**：使用标准键盘的 rawcode (0x31-0x34)

当 Cmd 键按下时，键 6-9 的 rawcode 会发生变化（例如 +40），导致快捷键匹配失败。

### 解决方案

**方案 1：仅使用稳定的键（推荐）**
```
✅ 使用：Cmd+0, Cmd+1, Cmd+2, Cmd+3, Cmd+4, Cmd+5
❌ 避免：Cmd+6, Cmd+7, Cmd+8, Cmd+9
```

**方案 2：使用字母键代替**
```
Cmd+1 → Cmd+Q
Cmd+6 → Cmd+A
Cmd+7 → Cmd+S
Cmd+8 → Cmd+D
Cmd+9 → Cmd+F
```

**方案 3：使用功能键**
```
Cmd+6 → Cmd+F6
Cmd+7 → Cmd+F7
Cmd+8 → Cmd+F8
Cmd+9 → Cmd+F9
```

### Numpad 数字小键盘

小键盘有独立的 rawcode 值：

| 键 | Numpad Rawcode | 说明 |
|---|----------------|------|
| **0** | 0x52 | 小键盘 0 |
| **1** | 0x53 | 小键盘 1 |
| **2** | 0x54 | 小键盘 2 |
| **3** | 0x55 | 小键盘 3 |
| **4** | 0x56 | 小键盘 4 |
| **5** | 0x57 | 小键盘 5 |
| **6** | 0x58 | 小键盘 6 |
| **7** | 0x59 | 小键盘 7 |
| **8** | 0x5B | 小键盘 8 |
| **9** | 0x5C | 小键盘 9 |

**注意**：小键盘的键码与主键盘数字键的键码不同，需要单独处理。

---

## 字母键 (A-Z)

### macOS 字母键映射

所有字母键使用连续的 rawcode 值，表现非常稳定：

| 键 | Rawcode | Keycode | Cmd 组合稳定性 | 说明 |
|----|---------|---------|---------------|------|
| **A** | 0x00 | 0x00 | ⭐⭐⭐⭐⭐ | 第一行左侧 |
| **S** | 0x01 | 0x01 | ⭐⭐⭐⭐⭐ | 第一行中间 |
| **D** | 0x02 | 0x02 | ⭐⭐⭐⭐⭐ | 第一行右侧 |
| **F** | 0x03 | 0x03 | ⭐⭐⭐⭐⭐ | 第一行右侧 |
| **H** | 0x04 | 0x04 | ⭐⭐⭐⭐⭐ | 第二行左侧 |
| **G** | 0x05 | 0x05 | ⭐⭐⭐⭐⭐ | 第二行左侧 |
| **Z** | 0x06 | 0x06 | ⭐⭐⭐⭐⭐ | 底行左侧 |
| **X** | 0x07 | 0x07 | ⭐⭐⭐⭐⭐ | 底行左侧 |
| **C** | 0x08 | 0x08 | ⭐⭐⭐⭐⭐ | 底行中间 |
| **V** | 0x09 | 0x09 | ⭐⭐⭐⭐⭐ | 底行中间 |
| **B** | 0x0B | 0x0B | ⭐⭐⭐⭐⭐ | 底行中间 |
| **Q** | 0x0B | 0x0B | ⭐⭐⭐⭐⭐ | 第一行极左 |
| **W** | 0x0C | 0x0C | ⭐⭐⭐⭐⭐ | 第一行左侧 |
| **E** | 0x0D | 0x0D | ⭐⭐⭐⭐⭐ | 第一行左侧 |
| **R** | 0x0E | 0x0E | ⭐⭐⭐⭐⭐ | 第一行中间 |
| **T** | 0x0F | 0x0F | ⭐⭐⭐⭐⭐ | 第二行左侧 |
| **Y** | 0x10 | 0x10 | ⭐⭐⭐⭐⭐ | 第一行中间 |
| **...** | ... | ... | ⭐⭐⭐⭐⭐ | 所有字母键稳定 |

### 完整字母键 Rawcode 表

```go
// QWERTY 布局
case 0x00: return "a"
case 0x01: return "s"
case 0x02: return "d"
case 0x03: return "f"
case 0x04: return "h"
case 0x05: return "g"
case 0x06: return "z"
case 0x07: return "x"
case 0x08: return "c"
case 0x09: return "v"
case 0x0A: return "b"  // 需要验证
case 0x0B: return "q"
case 0x0C: return "w"
case 0x0D: return "e"
case 0x0E: return "r"
case 0x0F: return "y"
case 0x10: return "t"
case 0x11: return "1"  // 实际上是数字键
case 0x12: return "2"  // 实际上是数字键
case 0x13: return "3"  // 实际上是数字键
// ... 其他字母键
```

### 推荐用途

由于所有字母键都极其稳定，建议优先使用：

```
插件快速访问：
- Cmd+A: 全部/所有
- Cmd+S: 搜索 (Search)
- Cmd+D: 字典 (Dictionary)
- Cmd+C: 计算器 (Calculator)
- Cmd+N: 笔记 (Note)
- Cmd+T: 终端 (Terminal)
- Cmd+G: Git
- Cmd+K: 代码片段 (Code snippet)
```

### 大小写处理

```go
// gohook 报告的键名始终是小写
// 大小写由 Shift 键状态决定

func (m *GlobalHotkeyManager) getKeyName(event hook.Event) string {
    keyName := m.mapRawcodeToKey(event.Rawcode)

    // 检查 Shift 状态
    if m.keyState["shift"] {
        return strings.ToUpper(keyName)
    }
    return keyName
}
```

---

## 符号键和标点

### macOS 符号键映射

| 符号 | 键位 | Rawcode | Keycode | Cmd 组合稳定性 |
|------|------|---------|---------|---------------|
| **;** | 分号 | 0x29 | 0x29 | ⭐⭐⭐⭐ |
| **'** | 单引号 | 0x27 | 0x27 | ⭐⭐⭐⭐ |
| **,** | 逗号 | 0x2B | 0x2B | ⭐⭐⭐⭐ |
| **.** | 句号 | 0x2F | 0x2F | ⭐⭐⭐⭐ |
| **/** | 斜杠 | 0x2C | 0x2C | ⭐⭐⭐⭐ |
| **\\** | 反斜杠 | 0x2A | 0x2A | ⭐⭐⭐⭐ |
| **-** | 减号 | 0x1D | 0x1D | ⭐⭐⭐⭐ |
| **=** | 等号 | 0x18 | 0x18 | ⭐⭐⭐⭐ |
| **[** | 左方括号 | 0x21 | 0x21 | ⭐⭐⭐⭐ |
| **]** | 右方括号 | 0x1E | 0x1E | ⭐⭐⭐⭐ |
| **`** | 反引号 | 0x32 | 0x32 | ⭐⭐⭐ |
| **~** | 波浪号 | 0x32 + Shift | - | ⭐⭐⭐ |
| **!** | 感叹号 | 0x1B + Shift | - | ⭐⭐⭐ |
| **@** | 艾特符号 | 0x1F + Shift | - | ⭐⭐⭐ |
| **#** | 井号 | 0x1C + Shift | - | ⭐⭐⭐ |
| **$** | 美元符号 | 0x1D + Shift | - | ⭐⭐⭐ |
| **%** | 百分号 | 0x1E + Shift | - | ⭐⭐⭐ |
| **^** | 脱字符 | 0x21 + Shift | - | ⭐⭐⭐ |
| **&** | 和符号 | 0x22 + Shift | - | ⭐⭐⭐ |
| ***** | 星号 | 0x23 + Shift | - | ⭐⭐⭐ |
| **(** | 左圆括号 | 0x24 + Shift | - | ⭐⭐⭐ |
| **)** | 右圆括号 | 0x25 + Shift | - | ⭐⭐⭐ |
| **_** | 下划线 | 0x26 + Shift | - | ⭐⭐⭐ |
| **+** | 加号 | 0x27 + Shift | - | ⭐⭐⭐ |
| **{** | 左花括号 | 0x28 + Shift | - | ⭐⭐⭐ |
| **}** | 右花括号 | 0x29 + Shift | - | ⭐⭐⭐ |
| **|** | 竖线 | 0x2A + Shift | - | ⭐⭐⭐ |
| **:** | 冒号 | 0x29 + Shift | - | ⭐⭐⭐ |
| **"** | 双引号 | 0x27 + Shift | - | ⭐⭐⭐ |
| **<** | 小于号 | 0x2B + Shift | - | ⭐⭐⭐ |
| **>** | 大于号 | 0x2F + Shift | - | ⭐⭐⭐ |
| **?** | 问号 | 0x2C + Shift | - | ⭐⭐⭐ |

### 使用建议

```
⚠️ 谨慎使用：符号键在 Cmd 修饰下可能不稳定
✅ 推荐使用：逗号、句号、分号等常用标点
❌ 避免使用：需要 Shift 组合的符号（如 @, #, $, %）
```

### 特殊符号键

| 符号 | 键位 | Rawcode | 说明 |
|------|------|---------|------|
| **Space** | 空格 | 0x31 | ⭐⭐⭐⭐⭐ |
| **Tab** | 制表符 | 0x30 | ⭐⭐⭐⭐⭐ |
| **Enter** | 回车 | 0x24 | ⭐⭐⭐⭐⭐ |
| **Return** | 返回 | 0x24 | 同 Enter |
| **Esc** | 退出 | 0x35 | ⭐⭐⭐⭐⭐ |

---

## 导航键

### macOS 导航键映射

| 键 | Rawcode | Keycode | 说明 |
|----|---------|---------|------|
| **Arrow Up (↑)** | 0x7E | 0x26 | 上箭头 |
| **Arrow Down (↓)** | 0x7D | 0x28 | 下箭头 |
| **Arrow Left (←)** | 0x7B | 0x25 | 左箭头 |
| **Arrow Right (→)** | 0x7C | 0x27 | 右箭头 |
| **Page Up** | 0x74 | 0x21 | 上页 |
| **Page Down** | 0x79 | 0x22 | 下页 |
| **Home** | 0x73 | 0x23 | 起始 |
| **End** | 0x77 | 0x24 | 结束 |

### 快捷键组合

```
Cmd+↑:   移动到开头
Cmd+↓:   移动到结尾
Cmd+←:   向前一个单词
Cmd+→:   向后一个单词

Option+↑:   向上一段
Option+↓:   向下一段
Option+←:   向前一个单词
Option+→:   向后一个单词
```

### 稳定性

导航键在 Cmd 修饰下通常稳定，但需要注意：
- 某些组合被系统保留（如 Cmd+↑ 用于移动到开头）
- 建议测试后使用

---

## 编辑键

### macOS 编辑键映射

| 键 | Rawcode | Keycode | 说明 | 稳定性 |
|----|---------|---------|------|--------|
| **Delete (后退)** | 0x33 | 0x08 | 删除光标前字符 | ⭐⭐⭐⭐⭐ |
| **Forward Delete** | 0x75 | 0x2E | 删除光标后字符 | ⭐⭐⭐⭐⭐ |
| **Insert** | 0x72 | 0x2D | 插入模式 | ⭐⭐⭐⭐ |
| **Backspace** | 0x33 | 0x08 | 同 Delete | ⭐⭐⭐⭐⭐ |

### 快捷键组合

```
Cmd+Delete:        删除到开头
Cmd+Forward Delete: 删除到结尾
Option+Delete:     删除前一个单词
Option+Forward Delete: 删除后一个单词
```

### 使用建议

```
✅ 推荐：独立使用（不与 Cmd 组合）
⚠️ 谨慎：与 Cmd 组合使用（可能与系统功能冲突）
```

---

## 平台特定问题

### macOS 特定问题

#### 1. Cmd 键导致的 Rawcode 变化

**问题**：数字键 6-9 在 Cmd 修饰下 rawcode 会变化

**解决方案**：
```go
// 检测 Cmd 键状态
func (m *GlobalHotkeyManager) handleKeyDown(event hook.Event) {
    keyName := m.getKeyName(event)

    // 如果 Cmd 按下且是数字键 6-9，使用特殊处理
    if m.keyState["cmd"] && isUnstableNumberKey(keyName) {
        // 使用备用映射
        keyName = m.getAlternateMapping(event)
    }

    m.keyState[keyName] = true
}
```

#### 2. 辅助功能权限

**问题**：macOS 需要辅助功能权限才能监听全局键盘事件

**检查权限**：
```go
func CheckAccessibilityPermissions() (bool, error) {
    // 使用系统 API 检查权限
    // 返回 true 如果已授权，false 否则
}
```

**请求权限**：
```go
func OpenAccessibilitySettings() error {
    // 打开系统设置 > 安全性与隐私 > 辅助功能
    return exec.Command("open", "x-apple.systempreferences:com.apple.preference.security?Privacy_Accessibility").Run()
}
```

#### 3. 系统保留快捷键

macOS 保留了许多 Cmd+X 组合：

| 快捷键 | 功能 | 建议 |
|--------|------|------|
| Cmd+Q | 退出应用 | ❌ 避免使用 |
| Cmd+W | 关闭窗口 | ⚠️ 谨慎使用 |
| Cmd+C | 复制 | ⚠️ 谨慎使用 |
| Cmd+V | 粘贴 | ⚠️ 谨慎使用 |
| Cmd+X | 剪切 | ⚠️ 谨慎使用 |
| Cmd+Z | 撤销 | ⚠️ 谨慎使用 |
| Cmd+A | 全选 | ⚠️ 谨慎使用 |
| Cmd+S | 保存 | ⚠️ 谨慎使用 |
| Cmd+F | 查找 | ⚠️ 谨慎使用 |
| Cmd+H | 隐藏 | ⚠️ 谨慎使用 |
| Cmd+M | 最小化 | ⚠️ 谨慎使用 |

### Windows 特定问题

#### 1. UAC 权限

某些全局快捷键可能需要管理员权限

#### 2. 键盘布局差异

不同语言键盘布局的键码可能不同

### Linux 特定问题

#### 1. X11 vs Wayland

Wayland 的安全模型限制了全局键盘钩子

#### 2. 权限要求

可能需要配置 XSecurity 或类似机制

---

## 已知问题和解决方案

### 问题 1：数字键 6-9 不稳定

**症状**：Cmd+6, Cmd+7, Cmd+8, Cmd+9 快捷键失效

**根本原因**：这些键的 rawcode 在 Cmd 修饰下会改变

**解决方案**：

1. **避免使用这些键**（推荐）
```go
// 在注册时验证
func (s *ShortcutService) SetShortcut(keyCombo, pluginID string) error {
    if isUnstableKeyCombo(keyCombo) {
        return fmt.Errorf("key combo %s is unstable", keyCombo)
    }
    // ...
}
```

2. **使用 keycode 代替 rawcode**
```go
func (m *GlobalHotkeyManager) getKeyName(event hook.Event) string {
    // 优先使用 keycode
    if keyName := m.mapKeycodeToKey(event.Keycode); keyName != "" {
        return keyName
    }
    // 备用 rawcode
    return m.mapRawcodeToKey(event.Rawcode)
}
```

3. **使用备用映射表**
```go
var unstableKeyAlternatives = map[string]string{
    "6": "f6",
    "7": "f7",
    "8": "f8",
    "9": "f9",
}
```

### 问题 2：缺少 KeyDown 事件

**症状**：某些键（特别是 Shift）不触发 KeyDown 事件

**GitHub Issue**: [#41](https://github.com/robotn/gohook/issues/41)

**解决方案**：

```go
// 监听 KeyHold 事件作为备用
func (m *GlobalHotkeyManager) processEvents() {
    for {
        select {
        case event := <-m.evChan:
            switch event.Kind {
            case hook.KeyHold:
                // 处理可能缺失的 KeyDown
                m.handleKeyHold(event)
            case hook.KeyDown:
                m.handleKeyDown(event)
            }
        }
    }
}
```

### 问题 3：不同键盘布局的兼容性

**症状**：非 QWERTY 布局下键码映射错误

**解决方案**：

```go
// 检测系统键盘布局
func getKeyboardLayout() string {
    // macOS: 使用 NSCurrentLocale
    // Windows: 使用 GetKeyboardLayout
    // Linux: 使用 XkbGetLayout
    return "QWERTY" // 示例
}

// 根据布局使用不同的映射表
func (m *GlobalHotkeyManager) getMappingTable(layout string) map[uint16]string {
    switch layout {
    case "QWERTZ":
        return germanKeyMapping
    case "AZERTY":
        return frenchKeyMapping
    default:
        return qwertyKeyMapping
    }
}
```

### 问题 4：与系统快捷键冲突

**症状**：注册的快捷键被系统拦截

**解决方案**：

1. **检查系统快捷键**
```bash
# macOS: 查看系统快捷键
# 系统设置 > 键盘 > 快捷键
```

2. **提供覆盖选项**
```go
type ShortcutOptions struct {
    OverrideSystem bool // 是否覆盖系统快捷键
}

func (s *ShortcutService) SetShortcutWithOptions(
    keyCombo, pluginID string,
    options ShortcutOptions,
) error {
    // ...
}
```

3. **使用备用组合**
```go
var safeAlternatives = map[string]string{
    "cmd+c": "cmd+shift+c",
    "cmd+v": "cmd+shift+v",
    "cmd+f": "cmd+shift+f",
}
```

---

## 实用代码示例

### 完整的键匹配函数

```go
// matchesHotkey 检查当前键状态是否匹配快捷键组合
func (m *GlobalHotkeyManager) matchesHotkey(keyCombo string) bool {
    parts := splitKeyCombo(keyCombo)

    // 检查所有键是否都在按下状态
    for _, part := range parts {
        normalized := toLower(part)
        if !m.keyState[normalized] {
            return false
        }
    }

    // 检查是否有额外的键按下（可选）
    if len(parts) < len(m.keyState) {
        return false // 要求精确匹配
    }

    return true
}
```

### 键组合规范化

```go
// normalizeKeyCombo 规范化键组合字符串
func normalizeKeyCombo(keyCombo string) string {
    parts := splitKeyCombo(keyCombo)

    var normalizedParts []string
    for _, part := range parts {
        normalized := toLower(part)
        normalizedParts = append(normalizedParts, normalized)
    }

    // 按固定顺序排列以确保一致性
    sort.Strings(normalizedParts)

    return strings.Join(normalizedParts, "+")
}
```

### 调试输出

```go
// debugKeyEvent 输出键事件的详细信息
func (m *GlobalHotkeyManager) debugKeyEvent(event hook.Event) {
    fmt.Printf("Event: Kind=%d, Keycode=0x%02X, Rawcode=0x%02X, Keychar='%c'\n",
        event.Kind, event.Keycode, event.Rawcode, event.Keychar)

    keyName := m.getKeyName(event)
    fmt.Printf("Resolved key name: '%s'\n", keyName)

    fmt.Printf("Current key state: %+v\n", m.keyState)
}
```

---

## 参考资料

### 官方文档

- [gohook GitHub 仓库](https://github.com/robotn/gohook)
- [libuiohook 文档](https://github.com/kwhat/libuiohook)
- [macOS CGEvent 参考](https://developer.apple.com/documentation/coregraphics/cgevent)

### 相关 Issues

- [gohook Issue #41: Key code standards](https://github.com/robotn/gohook/issues/41)
- [gohook Issue #45: Missing KeyDown events](https://github.com/robotn/gohook/issues/45) (假设)

### 标准参考

- [Mozilla Key Code 标准](https://developer.mozilla.org/en-US/docs/Web/API/KeyboardEvent/keyCode)
- [UI Events KeyboardEvent Code Values](https://www.w3.org/TR/uievents-code/)

### 本项目实现

- `/Users/yanglian/code/ltools/internal/plugins/gohook_hotkey.go` - 全局热键管理器实现
- `/Users/yanglian/code/ltools/internal/plugins/shortcut_service.go` - 快捷键服务

---

## 更新日志

- **v1.0** (2026-02-07):
  - 初始版本
  - 完整的键映射参考
  - macOS 平台问题详细分析
  - 数字键不稳定性问题记录
  - 已知问题和解决方案

---

## 贡献

如果发现新的键映射信息或问题，请更新本文档并提交 Pull Request。

**模板**：
```markdown
### 键名

| 键 | Rawcode | Keycode | Cmd 组合稳定性 | 说明 |
|----|---------|---------|---------------|------|
| **X** | 0xXX | 0xXX | ⭐⭐⭐⭐⭐ | 说明文字 |
```

---

**文档结束**

如有问题或建议，请在项目 Issues 中提出。
