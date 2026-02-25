import React from 'react';
import { Icon } from '../../Icon';
import { AnnotationType } from '../hooks/useAnnotation';
import './toolbar.css';

// 工具定义
const ANNOTATION_TOOLS: { type: AnnotationType; icon: string; label: string; shortcut: string }[] = [
  { type: 'rect', icon: 'rectangle', label: '矩形', shortcut: 'R' },
  { type: 'arrow', icon: 'arrow-right', label: '箭头', shortcut: 'A' },
  { type: 'text', icon: 'type', label: '文字', shortcut: 'T' },
  { type: 'brush', icon: 'brush', label: '画笔', shortcut: 'B' },
  { type: 'mosaic', icon: 'mosaic', label: '马赛克', shortcut: 'M' },
];

// 预设颜色
const COLORS = [
  '#FF0000', // 红
  '#FFFF00', // 黄
  '#00FF00', // 绿
  '#00FFFF', // 青
  '#0080FF', // 蓝
  '#FF00FF', // 紫
  '#FFFFFF', // 白
  '#000000', // 黑
];

// 线宽选项
const STROKE_WIDTHS = [2, 4, 6];

interface ToolbarProps {
  // 选区信息
  selection: { x: number; y: number; width: number; height: number } | null;
  canvasWidth: number;
  canvasHeight: number;
  scaleFactor: number;

  // 标注工具状态
  currentTool: AnnotationType;
  currentColor: string;
  strokeWidth: number;
  canUndo: boolean;
  canRedo: boolean;

  // 操作回调
  onToolChange: (tool: AnnotationType) => void;
  onColorChange: (color: string) => void;
  onStrokeWidthChange: (width: number) => void;
  onUndo: () => void;
  onRedo: () => void;
  onClear: () => void;
  onCopy: () => void;
  onSave: () => void;
  onCancel: () => void;
}

/**
 * 截图工具栏组件
 * 包含标注工具、颜色选择、线宽选择、撤销重做、复制保存取消
 */
export const Toolbar: React.FC<ToolbarProps> = ({
  selection,
  canvasWidth,
  canvasHeight,
  scaleFactor,
  currentTool,
  currentColor,
  strokeWidth,
  canUndo,
  canRedo,
  onToolChange,
  onColorChange,
  onStrokeWidthChange,
  onUndo,
  onRedo,
  onClear,
  onCopy,
  onSave,
  onCancel,
}) => {
  if (!selection || selection.width === 0 || selection.height === 0) {
    return null;
  }

  // 将物理像素坐标转换为逻辑像素（用于 CSS 定位）
  const logicalX = selection.x / scaleFactor;
  const logicalY = selection.y / scaleFactor;
  const logicalWidth = selection.width / scaleFactor;
  const logicalHeight = selection.height / scaleFactor;
  const logicalCanvasWidth = canvasWidth / scaleFactor;
  const logicalCanvasHeight = canvasHeight / scaleFactor;

  // 计算工具栏位置
  const toolbarHeight = 44;
  const toolbarWidth = 420; // 估计宽度
  const margin = 10;

  // 默认放在选区下方居中
  let toolbarY = logicalY + logicalHeight + margin;
  let toolbarX = logicalX + logicalWidth / 2;

  // 如果下方放不下，放在选区上方
  if (toolbarY + toolbarHeight > logicalCanvasHeight) {
    toolbarY = logicalY - toolbarHeight - margin;
  }

  // 确保工具栏不超出左右边界
  const halfWidth = toolbarWidth / 2;
  if (toolbarX - halfWidth < 0) {
    toolbarX = halfWidth;
  } else if (toolbarX + halfWidth > logicalCanvasWidth) {
    toolbarX = logicalCanvasWidth - halfWidth;
  }

  return (
    <div
      className="screenshot-toolbar"
      style={{
        position: 'absolute',
        left: toolbarX,
        top: toolbarY,
        transform: 'translateX(-50%)',
      }}
    >
      {/* 标注工具组 */}
      <div className="toolbar-group">
        {ANNOTATION_TOOLS.map((tool) => (
          <button
            key={tool.type}
            className={`toolbar-btn ${currentTool === tool.type ? 'active' : ''}`}
            onClick={() => onToolChange(tool.type)}
            title={`${tool.label} (${tool.shortcut})`}
          >
            <Icon name={tool.icon as any} size={16} />
            <span className="shortcut">{tool.shortcut}</span>
          </button>
        ))}
      </div>

      {/* 颜色选择器 */}
      <div className="toolbar-group color-picker">
        {COLORS.map((color) => (
          <button
            key={color}
            className={`color-btn ${currentColor === color ? 'active' : ''}`}
            style={{ backgroundColor: color }}
            onClick={() => onColorChange(color)}
            title={`颜色 ${color}`}
          />
        ))}
      </div>

      {/* 线宽选择器 */}
      <div className="toolbar-group stroke-picker">
        {STROKE_WIDTHS.map((width) => (
          <button
            key={width}
            className={`stroke-btn ${strokeWidth === width ? 'active' : ''}`}
            onClick={() => onStrokeWidthChange(width)}
            title={`线宽 ${width}px`}
          >
            <div className="stroke-indicator">
              <span style={{ width: width + 4, height: width }} />
            </div>
          </button>
        ))}
      </div>

      {/* 编辑操作组 */}
      <div className="toolbar-group">
        <button
          className="toolbar-btn"
          onClick={onUndo}
          disabled={!canUndo}
          title="撤销 (Ctrl+Z)"
        >
          <Icon name="undo" size={16} />
        </button>
        <button
          className="toolbar-btn"
          onClick={onRedo}
          disabled={!canRedo}
          title="重做 (Ctrl+Shift+Z)"
        >
          <Icon name="redo" size={16} />
        </button>
        <button
          className="toolbar-btn"
          onClick={onClear}
          title="清除标注 (Delete)"
        >
          <Icon name="trash" size={16} />
        </button>
      </div>

      {/* 最终操作组 */}
      <div className="toolbar-group">
        <button
          className="toolbar-btn"
          onClick={onCopy}
          title="复制到剪贴板 (Enter)"
        >
          <Icon name="copy" size={16} />
        </button>
        <button
          className="toolbar-btn"
          onClick={onSave}
          title="保存文件"
        >
          <Icon name="download" size={16} />
        </button>
        <button
          className="toolbar-btn"
          onClick={onCancel}
          title="取消 (ESC)"
        >
          <Icon name="x-mark" size={16} />
        </button>
      </div>
    </div>
  );
};

export default Toolbar;
