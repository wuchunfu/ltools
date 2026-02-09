import React from 'react';
import './ToolbarLayer.css';

interface ToolbarLayerProps {
  visible: boolean;
  hasSelection: boolean;
  onConfirm: () => void;
  onCancel: () => void;
  selection?: { width: number; height: number };
}

/**
 * 工具层 - 负责显示操作按钮和提示信息
 * 职责：
 * 1. 显示确认/取消按钮
 * 2. 显示操作提示信息
 * 3. 显示选择区域尺寸
 * 4. 条件渲染（仅在有选择时显示）
 */
export const ToolbarLayer: React.FC<ToolbarLayerProps> = ({
  visible,
  hasSelection,
  onConfirm,
  onCancel,
  selection
}) => {
  if (!visible) return null;

  return (
    <div className="toolbar-layer">
      {/* 主工具栏 */}
      {hasSelection && selection && selection.width > 0 && selection.height > 0 ? (
        <>
          {/* 确认/取消工具栏 */}
          <div className="toolbar toolbar-main">
            <button onClick={onConfirm} className="toolbar-button toolbar-button-confirm">
              <span className="toolbar-icon">✓</span>
              <span>确认</span>
            </button>
            <button onClick={onCancel} className="toolbar-button toolbar-button-cancel">
              <span className="toolbar-icon">✕</span>
              <span>取消</span>
            </button>
          </div>

          {/* 底部提示信息 */}
          <div className="toolbar-hint">
            <span className="hint-shortcut">双击</span> 或
            <span className="hint-shortcut">Enter</span> 确认 |
            拖拽重新选择 |
            <span className="hint-shortcut">ESC</span> 取消
          </div>
        </>
      ) : (
        /* 初始提示信息 */
        <div className="toolbar-hint">
          拖拽选择截图范围
        </div>
      )}
    </div>
  );
};

export default ToolbarLayer;
