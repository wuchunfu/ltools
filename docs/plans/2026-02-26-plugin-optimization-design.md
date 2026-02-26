# 插件优化设计方案

**日期**: 2026-02-26
**状态**: 已确认
**涉及插件**: Tunnel（内网穿透）、ProcessManager（进程管理器）

---

## 1. 概述

### 1.1 优化目标

| 插件 | 问题 | 解决方案 |
|------|------|----------|
| **Tunnel** | 原生 `<select>` 在深色主题下样式不协调 | 引入 shadcn Select 组件，适配玻璃态设计 |
| **ProcessManager** | 后端进程查询阻塞导致应用卡死 | 强化超时控制 + 进程黑名单机制 |

### 1.2 实施范围

**Tunnel 插件**：
- 安装 `@radix-ui/react-select` 和 `@radix-ui/react-icons`
- 创建 `Select` 组件（位于 `frontend/src/components/ui/`）
- 适配现有玻璃态设计变量
- 替换 TunnelWidget 中的 2 处原生 `<select>`

**ProcessManager 插件**：
- 将 CPU 获取超时从 10ms 降至 5ms
- 为所有 gopsutil 调用添加统一超时包装器
- 实现进程黑名单机制（跳过僵尸进程、权限受限进程）
- 添加跳过计数日志便于调试

---

## 2. Select 组件技术设计

### 2.1 文件结构

```
frontend/src/components/ui/
├── select.tsx           # shadcn Select 组件（适配玻璃态）
└── index.ts             # 组件导出（可选）
```

### 2.2 核心 API

```tsx
// 使用方式（与原生 select 类似）
<Select value={value} onValueChange={setValue}>
  <SelectTrigger className="...">
    <SelectValue placeholder="选择代理类型" />
  </SelectTrigger>
  <SelectContent>
    <SelectItem value="http">HTTP</SelectItem>
    <SelectItem value="https">HTTPS</SelectItem>
    <SelectItem value="tcp">TCP</SelectItem>
  </SelectContent>
</Select>
```

### 2.3 样式适配策略

将 shadcn 默认样式替换为玻璃态设计变量：

| 元素 | shadcn 默认 | 玻璃态适配 |
|------|-------------|------------|
| Trigger 背景 | `bg-background` | `bg-white/10` |
| Trigger 边框 | `border-input` | `border-white/20` |
| Content 背景 | `bg-popover` | `bg-[#1A1F2E]/95 backdrop-blur-xl` |
| Item hover | `focus:bg-accent` | `focus:bg-white/10` |
| 选中指示 | `bg-primary` | `bg-[#7C3AED]` |

### 2.4 依赖安装

```bash
npm install @radix-ui/react-select @radix-ui/react-icons
```

---

## 3. ProcessManager 后端优化

### 3.1 超时控制强化

**当前问题**：
- CPU 获取超时 10ms 可能仍不够
- 部分调用（如 `Username()`, `Exe()`）没有超时控制

**优化方案**：创建统一超时包装器

```go
// plugins/processmanager/processmanager.go

const (
    defaultTimeout    = 50 * time.Millisecond  // 单个操作超时
    cpuTimeout        = 5 * time.Millisecond   // CPU 获取专用（更严格）
    processTimeout    = 200 * time.Millisecond // 整个进程信息获取超时
)
```

### 3.2 进程黑名单机制

**黑名单策略**：

| 类型 | 判断条件 | 处理 |
|------|----------|------|
| 僵尸进程 | `Status() == "Z"` | 跳过，返回基本信息 |
| 权限受限 | `Exe()` 返回权限错误 | 跳过敏感字段 |
| 内核进程 | PID = 0 (kernel_task) | 跳过 CPU/内存查询 |
| 已知问题进程 | 维护配置列表 | 完全跳过 |

**黑名单数据结构**：

```go
type ProcessBlacklist struct {
    PIDs        map[int32]string // PID -> 原因
    Names       map[string]bool  // 进程名 -> 跳过
    SkipCPU     map[int32]bool   // 仅跳过 CPU 查询
}
```

### 3.3 降级返回策略

当获取完整信息失败时，返回降级数据而非 nil：

```go
// 成功：完整信息
// 超时/错误：基本信息（PID、Name）+ 错误标记
type ProcessInfo struct {
    // ... 现有字段
    PartialData bool   `json:"partialData"` // 是否为降级数据
    ErrorReason string `json:"errorReason,omitempty"`
}
```

---

## 4. 实施计划

### 4.1 变更文件清单

| 文件 | 操作 | 说明 |
|------|------|------|
| `frontend/package.json` | 修改 | 添加 radix-ui 依赖 |
| `frontend/src/components/ui/select.tsx` | 新建 | Select 组件（玻璃态适配） |
| `frontend/src/components/TunnelWidget.tsx` | 修改 | 替换原生 select 为 Select 组件 |
| `plugins/processmanager/processmanager.go` | 修改 | 超时控制 + 黑名单机制 |
| `plugins/processmanager/types.go` | 修改 | 添加 PartialData、ErrorReason 字段 |

### 4.2 实施顺序

```
Phase 1: Tunnel Select 组件
├── 1.1 安装依赖
├── 1.2 创建 ui/select.tsx
├── 1.3 替换 TunnelWidget 中的 select
└── 1.4 测试验证

Phase 2: ProcessManager 后端优化
├── 2.1 添加超时常量和包装器
├── 2.2 实现黑名单机制
├── 2.3 修改 getProcessInfo 使用降级策略
├── 2.4 添加调试日志
└── 2.5 测试验证
```

### 4.3 验收标准

- [ ] Tunnel: Select 下拉框显示玻璃态深色样式
- [ ] Tunnel: 所有选项功能正常，表单提交无误
- [ ] ProcessManager: 无明显卡死现象（响应时间 < 1s）
- [ ] ProcessManager: 黑名单进程正确跳过
- [ ] ProcessManager: 降级数据正确显示

---

## 5. 风险与缓解

| 风险 | 可能性 | 影响 | 缓解措施 |
|------|--------|------|----------|
| Radix UI 与 TailwindCSS 4 兼容问题 | 低 | 中 | 参考官方文档，使用 className 覆盖 |
| 超时设置过短导致信息丢失 | 中 | 低 | 使用降级策略，保留基本信息 |
| 黑名单遗漏问题进程 | 中 | 低 | 添加调试日志，持续更新黑名单 |
