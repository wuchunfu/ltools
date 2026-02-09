# 截图编辑器三层架构 - 实现建议总结

## 📐 架构概览

基于当前项目，我已经实现了一个清晰的**三层架构**来组织截图编辑器的代码：

```
┌─────────────────────────────────────────────────────────┐
│                   ScreenshotEditorLayers                │
│                      (主协调组件)                        │
│                                                          │
│  ┌──────────────────────────────────────────────────┐  │
│  │  工具层 (ToolbarLayer) - z-index: 10002         │  │
│  │  • 显示确认/取消按钮                             │  │
│  │  • 显示操作提示信息                              │  │
│  │  • 条件渲染（仅在有选择时显示）                  │  │
│  └──────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────┐  │
│  │  交互层 (InteractionLayer) - z-index: 10001     │  │
│  │  • 处理鼠标/触摸事件                            │  │
│  │  • 计算选择区域坐标和尺寸                        │  │
│  │  • 绘制选择边框和尺寸信息                        │  │
│  │  • 处理双击确认                                 │  │
│  └──────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────┐  │
│  │  遮罩层 (MaskLayer) - z-index: 10000            │  │
│  │  • 加载并渲染原始截图                            │  │
│  │  • 绘制半透明黑色遮罩                            │  │
│  │  • 清除选中区域的遮罩                            │  │
│  │  • 不处理用户事件                                │  │
│  └──────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

## 🎯 三层职责详解

### 1️⃣ 遮罩层 (MaskLayer)

**文件位置**: `components/ScreenshotEditor/MaskLayer.tsx`

**核心职责**:
- ✅ 渲染原始截图图片
- ✅ 绘制半透明遮罩 (rgba(0, 0, 0, 0.3))
- ✅ 清除选中区域的遮罩，保持清晰可见
- ✅ 纯视觉层，不处理事件

**关键实现**:
```tsx
// 使用 pointer-events: none 确保不拦截事件
canvas.className = "mask-canvas";
// CSS: pointer-events: none;

// 绘制遮罩逻辑
ctx.fillStyle = 'rgba(0, 0, 0, 0.3)';
ctx.fillRect(0, 0, canvas.width, canvas.height);

// 清除选中区域遮罩
ctx.clearRect(selection.x, selection.y, selection.width, selection.height);
```

### 2️⃣ 交互层 (InteractionLayer)

**文件位置**: `components/ScreenshotEditor/InteractionLayer.tsx`

**核心职责**:
- ✅ 监听所有鼠标/触摸事件
- ✅ 实时计算选择区域
- ✅ 绘制绿色选择边框 (#4CD964)
- ✅ 显示选择区域尺寸
- ✅ 处理双击确认

**关键实现**:
```tsx
// 事件处理
onMouseDown={handleMouseDown}
onMouseMove={handleMouseMove}
onMouseUp={handleMouseUp}
onClick={handleClick}

// 选择框绘制
ctx.strokeStyle = '#4CD964';
ctx.lineWidth = 2;
ctx.strokeRect(selection.x, selection.y, selection.width, selection.height);

// 尺寸显示
const sizeText = `${selection.width} x ${selection.height}`;
ctx.fillText(sizeText, selection.x + 6, selection.y - 10);
```

### 3️⃣ 工具层 (ToolbarLayer)

**文件位置**: `components/ScreenshotEditor/ToolbarLayer.tsx`

**核心职责**:
- ✅ 显示操作按钮（确认/取消）
- ✅ 显示提示信息
- ✅ 条件渲染
- ✅ 处理按钮点击事件

**关键实现**:
```tsx
// 条件渲染
{hasSelection ? (
  <>
    <div className="toolbar toolbar-main">
      <button onClick={onConfirm}>✓ 确认</button>
      <button onClick={onCancel}>✕ 取消</button>
    </div>
    <div className="toolbar-hint">操作提示...</div>
  </>
) : (
  <div className="toolbar-hint">拖拽选择截图范围</div>
)}

// 容器不处理事件，按钮可点击
// CSS: pointer-events: none (容器)
// CSS: pointer-events: auto (按钮)
```

## 💡 设计优势

### 1. 职责分离
每层都有明确的职责，互不干扰：
- 遮罩层只负责渲染
- 交互层只负责事件处理
- 工具层只负责 UI 显示

### 2. 易于维护
修改某一层的功能不会影响其他层：
- 修改遮罩透明度 → 只改 MaskLayer
- 添加新手势 → 只改 InteractionLayer
- 调整按钮样式 → 只改 ToolbarLayer

### 3. 可扩展性强
可以轻松添加新功能：
```tsx
// 添加标注层
<AnnotationLayer z-index={10003} />

// 添加滤镜层
<FilterLayer z-index={9999} />

// 添加撤销/重做
<HistoryLayer onUndo={...} onRedo={...} />
```

### 4. 独立测试
每层可以独立进行单元测试：
```tsx
// 测试遮罩层
test('MaskLayer 应该正确绘制遮罩', () => {
  render(<MaskLayer imageSrc="..." selection={{...}} />);
  // 测试逻辑
});

// 测试交互层
test('InteractionLayer 应该正确处理鼠标事件', () => {
  render(<InteractionLayer enabled={true} />);
  // 测试逻辑
});

// 测试工具层
test('ToolbarLayer 应该在有选择时显示', () => {
  render(<ToolbarLayer visible={true} hasSelection={true} />);
  // 测试逻辑
});
```

## 🚀 使用示例

### 基础用法

```tsx
import { ScreenshotEditorLayers } from '@/components/ScreenshotEditor';

function MyScreenshotApp() {
  const handleComplete = (imageData: string, bounds: Bounds) => {
    // 处理完成的截图
    console.log('截图区域:', bounds);
    saveScreenshot(imageData);
  };

  const handleCancel = () => {
    // 处理取消
    console.log('用户取消');
  };

  return (
    <ScreenshotEditorLayers
      imageData={screenshotData}
      onSelectionComplete={handleComplete}
      onCancel={handleCancel}
    />
  );
}
```

### 高级用法 - 自定义工具

```tsx
import { MaskLayer, InteractionLayer } from '@/components/ScreenshotEditor';

function CustomEditor() {
  const [selection, setSelection] = useState<SelectionRect>({...});

  return (
    <>
      <MaskLayer imageSrc={...} selection={selection} canvasRef={...} />
      <InteractionLayer
        enabled={true}
        onSelectionComplete={setSelection}
      />
      {/* 自定义工具栏 */}
      <div className="custom-toolbar">
        <button>💾 保存</button>
        <button>📋 复制</button>
        <button>🎨 标注</button>
        <button>🔍 滤镜</button>
      </div>
    </>
  );
}
```

## 📊 性能优化建议

### 1. 使用 React.memo

```tsx
export const MaskLayer = React.memo<MaskLayerProps>(
  ({ imageSrc, selection, canvasRef }) => {
    // 组件实现
  },
  (prevProps, nextProps) => {
    // 仅当 selection 变化时重新渲染
    return (
      prevProps.selection.x === nextProps.selection.x &&
      prevProps.selection.y === nextProps.selection.y &&
      prevProps.selection.width === nextProps.selection.width &&
      prevProps.selection.height === nextProps.selection.height
    );
  }
);
```

### 2. 使用 useCallback

```tsx
const handleMouseDown = useCallback((e: React.MouseEvent) => {
  // 事件处理逻辑
}, [依赖项]);
```

### 3. 使用 requestAnimationFrame

```tsx
const drawSelection = useCallback(() => {
  requestAnimationFrame(() => {
    // 绘制逻辑
  });
}, [selection]);
```

## 📝 与现有代码的集成

### 替换现有的 ScreenshotSelector

```tsx
// 旧代码
// import ScreenshotSelector from './ScreenshotSelector';

// 新代码
import { ScreenshotEditorLayers } from './ScreenshotEditor';

// 使用方式相同
<ScreenshotEditorLayers
  imageData={imageData}
  onSelectionComplete={onSelectionComplete}
  onCancel={onCancel}
/>
```

### 向后兼容

保持相同的接口，确保无需修改调用代码：

```tsx
interface ScreenshotSelectorProps {
  imageData: string;
  onSelectionComplete: (selectedImageData: string, bounds: Bounds) => void;
  onCancel: () => void;
}
```

## 🎨 样式定制

### 修改遮罩透明度

```css
/* MaskLayer.css */
.mask-canvas {
  /* 修改遮罩透明度为 50% */
  /* 在组件中修改: rgba(0, 0, 0, 0.5) */
}
```

### 修改选择框样式

```tsx
/* InteractionLayer.tsx */
ctx.strokeStyle = '#FF6B6B'; // 改为红色
ctx.lineWidth = 3; // 加粗边框
ctx.setLineDash([5, 5]); // 虚线边框
```

### 修改工具栏样式

```css
/* ToolbarLayer.css */
.toolbar-button-confirm {
  background: rgba(59, 130, 246, 0.9); /* 蓝色 */
  border-radius: 12px; /* 更圆的角 */
  padding: 12px 24px; /* 更大的点击区域 */
}
```

## 🔜 未来扩展方向

### 短期扩展
- [ ] 添加矩形、箭头、文字标注工具
- [ ] 支持撤销/重做功能
- [ ] 添加保存和复制到剪贴板功能

### 中期扩展
- [ ] 支持模糊和马赛克工具
- [ ] 添加颜色选择器
- [ ] 支持多选区域

### 长期扩展
- [ ] 支持视频截图
- [ ] 添加 OCR 文字识别
- [ ] 支持云端保存和分享
- [ ] 添加插件系统

## 📚 相关文档

- **README.md**: 详细的架构文档
- **QUICKSTART.md**: 快速开始指南
- **源代码**: `components/ScreenshotEditor/`

---

**总结**: 这个三层架构提供了一个清晰、可维护、可扩展的截图编辑器实现方案。通过职责分离，每层都可以独立开发和测试，大大降低了代码复杂度和维护成本。
