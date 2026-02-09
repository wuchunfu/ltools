# Wails v3 窗口隐藏行为测试

## 测试日期
2026-02-08

## 测试目的
验证 Wails v3 窗口隐藏时，前端 JavaScript 代码是否继续执行。

## 测试方法
使用三种 JavaScript 定时机制持续记录日志：
1. `setInterval(100ms)` - 标准周期性定时器
2. `setTimeout(200ms)` - 递归链式定时器
3. `requestAnimationFrame` - 浏览器渲染循环

## 测试结果

### ✅ 结论：窗口隐藏时 JavaScript 继续执行

当通过系统托盘菜单隐藏窗口后，所有三种定时器**继续运行**，日志中出现了隐藏期间的时间戳记录。

## 技术分析

### Wails v3 的行为
- **WebView 持续活动**：隐藏窗口不会暂停 WebView 的 JS 引擎
- **渲染可能降级**：虽然 JS 执行，但 UI 渲染可能被优化
- **与浏览器不同**：浏览器标签页后台会降频/暂停，Wails 不会

### 对比其他框架

| 框架 | 隐藏窗口行为 |
|------|-------------|
| Wails v3 | ✅ JS 继续执行 |
| Electron | ✅ JS 继续执行 |
| Tauri | ⚠️ 可能暂停部分渲染 |
| 浏览器 | ❌ 后台标签页降频/暂停 |

## 实际应用影响

### ✅ 积极影响
- **后台任务可靠**：剪贴板监控、定时任务等不会被打断
- **状态保持简单**：切换窗口无需复杂的恢复逻辑
- **实时数据连续**：WebSocket、轮询等可持续工作

### ⚠️ 需要注意
- **CPU 消耗**：隐藏窗口仍占用 CPU，高频定时器会继续消耗资源
- **内存占用**：持续的数据累积可能导致内存泄漏
- **电池影响**：便携设备上后台持续运行会消耗电池

## 优化建议

### 使用 Page Visibility API 降低后台活动

```typescript
useEffect(() => {
  const handleVisibilityChange = () => {
    if (document.hidden) {
      // 窗口隐藏：降低频率
      clearInterval(intervalRef.current)
      intervalRef.current = setInterval(update, 1000) // 从 100ms 降到 1s
    } else {
      // 窗口显示：恢复频率
      clearInterval(intervalRef.current)
      intervalRef.current = setInterval(update, 100)
    }
  }

  document.addEventListener('visibilitychange', handleVisibilityChange)
  return () => document.removeEventListener('visibilitychange', handleVisibilityChange)
}, [])
```

### 针对不同场景的优化策略

| 场景 | 可见时 | 隐藏时 |
|------|--------|--------|
| 剪贴板监控 | 正常轮询 | 正常轮询（需要实时） |
| 数据刷新 | 1-5秒 | 30-60秒 |
| 动画效果 | 60fps | 暂停 |
| 定时统计 | 实时记录 | 批量记录 |

## 测试工具

测试工具位于 `frontend/src/plugins/WindowHiddenTest.tsx`（已移除）

如需重新测试，可以恢复该文件并在 `App.tsx` 中添加对应的菜单项和路由。

## 相关文档
- [Page Visibility API](https://developer.mozilla.org/en-US/docs/Web/API/Page_Visibility_API)
- [Wails v3 Documentation](https://v3.wails.io/)
