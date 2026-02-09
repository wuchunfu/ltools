# 快速开始 - 三层架构截图编辑器

## 🚀 集成到现有项目

### 步骤 1: 更新 ScreenshotSelector 使用新架构

```tsx
// frontend/src/components/ScreenshotSelector.tsx
import { ScreenshotEditorLayers } from './ScreenshotEditor';

interface ScreenshotSelectorProps {
  imageData: string;
  onSelectionComplete: (selectedImageData: string, bounds: Bounds) => void;
  onCancel: () => void;
}

export const ScreenshotSelector: React.FC<ScreenshotSelectorProps> = ({
  imageData,
  onSelectionComplete,
  onCancel
}) => {
  return (
    <ScreenshotEditorLayers
      imageData={imageData}
      onSelectionComplete={onSelectionComplete}
      onCancel={onCancel}
    />
  );
};
```

### 步骤 2: 测试基本功能

```bash
# 运行开发服务器
cd frontend
npm run dev

# 触发截图（在主应用中）
# 按下 Cmd+Shift+6
```

## 📦 文件结构

```
frontend/src/components/ScreenshotEditor/
├── MaskLayer.tsx              # 遮罩层组件
├── MaskLayer.css              # 遮罩层样式
├── InteractionLayer.tsx       # 交互层组件
├── InteractionLayer.css       # 交互层样式
├── ToolbarLayer.tsx           # 工具层组件
├── ToolbarLayer.css           # 工具层样式
├── ScreenshotEditorLayers.tsx # 主整合组件
├── ScreenshotEditorLayers.css # 主组件样式
├── index.ts                   # 导出文件
├── README.md                  # 详细文档
└── QUICKSTART.md              # 本文件
```

## 🎯 核心概念

### 层级关系

```
z-index: 10002 ──► 工具层 (按钮、提示)
z-index: 10001 ──► 交互层 (选择框、事件)
z-index: 10000 ──► 遮罩层 (图片、半透明遮罩)
```

### 数据流

```
用户操作 → 交互层 → 状态更新 → 遮罩层重绘 → 工具层更新
         ↓
    回调函数 → 父组件 → 后端服务
```

## 🔧 常见问题

### Q: 如何修改遮罩透明度？

A: 编辑 `MaskLayer.tsx` 第 37 行：

```tsx
ctx.fillStyle = 'rgba(0, 0, 0, 0.5)'; // 改为 50%
```

### Q: 如何添加更多工具按钮？

A: 编辑 `ToolbarLayer.tsx`，添加新按钮：

```tsx
<button className="toolbar-button" onClick={handleSave}>
  <span className="toolbar-icon">💾</span>
  <span>保存</span>
</button>
```

### Q: 如何支持触摸设备？

A: 在 `InteractionLayer.tsx` 添加触摸事件：

```tsx
onTouchStart={handleTouchStart}
onTouchMove={handleTouchMove}
onTouchEnd={handleTouchEnd}
```

## 💡 最佳实践

### 1. 保持职责分离

```tsx
// ✅ 好的做法
<MaskLayer />      // 只负责渲染
<InteractionLayer /> // 只负责交互
<ToolbarLayer />    // 只负责UI

// ❌ 不好的做法
<MixedLayer /> // 混合了所有逻辑
```

### 2. 使用 TypeScript 类型

```tsx
interface SelectionRect {
  x: number;
  y: number;
  width: number;
  height: number;
}
```

### 3. 添加错误处理

```tsx
const handleConfirm = () => {
  try {
    if (!hasSelection) {
      throw new Error('没有选择区域');
    }
    // 处理确认逻辑
  } catch (error) {
    console.error('确认失败:', error);
  }
};
```

## 📚 下一步

1. **阅读完整文档**: 查看 `README.md` 了解详细架构
2. **自定义样式**: 修改 CSS 文件以匹配你的设计
3. **扩展功能**: 添加标注工具、滤镜等
4. **性能优化**: 使用 React.memo 和 useMemo
5. **测试覆盖**: 添加单元测试和集成测试

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！
