import React, { useState, useRef, useEffect, useCallback } from 'react';
import { MaskLayer } from './MaskLayer';
import { InteractionLayer, SelectionRect } from './InteractionLayer';
import { ToolbarLayer } from './ToolbarLayer';
import './ScreenshotEditorLayers.css';

interface ScreenshotEditorLayersProps {
  imageData: string;
  onSelectionComplete: (selectedImageData: string, bounds: { x: number; y: number; width: number; height: number }) => void;
  onCancel: () => void;
}

/**
 * 截图编辑器 - 三层架构实现
 *
 * 修复版本：
 * - 解决了 canvas 尺寸同步问题
 * - 修复了依赖数组问题
 * - 移除了 @ts-ignore
 * - 正确处理了闭包问题
 */
const ScreenshotEditorLayers: React.FC<ScreenshotEditorLayersProps> = ({
  imageData,
  onSelectionComplete,
  onCancel
}) => {
  const [selection, setSelection] = useState<SelectionRect>({ x: 0, y: 0, width: 0, height: 0 });
  const [hasSelection, setHasSelection] = useState(false);
  const [imageSize, setImageSize] = useState({ width: 0, height: 0 });

  // Canvas 引用
  const maskCanvasRef = useRef<HTMLCanvasElement>(null);
  const interactionCanvasRef = useRef<HTMLCanvasElement>(null);
  const imageRef = useRef<HTMLImageElement | null>(null);

  // 加载图片并获取尺寸
  useEffect(() => {
    const img = new Image();
    img.onload = () => {
      const width = img.width;
      const height = img.height;
      setImageSize({ width, height });

      // 设置两个 canvas 的尺寸
      if (maskCanvasRef.current) {
        maskCanvasRef.current.width = width;
        maskCanvasRef.current.height = height;
      }
      if (interactionCanvasRef.current) {
        interactionCanvasRef.current.width = width;
        interactionCanvasRef.current.height = height;
      }
    };
    img.src = imageData;
    imageRef.current = img;

    // 清理函数
    return () => {
      imageRef.current = null;
    };
  }, [imageData]);

  // 确认选择
  const handleConfirm = useCallback(() => {
    if (!hasSelection || selection.width === 0 || selection.height === 0) {
      return;
    }

    // 裁剪选中区域的图片
    const img = imageRef.current;
    if (!img) return;

    const croppedCanvas = document.createElement('canvas');
    croppedCanvas.width = selection.width;
    croppedCanvas.height = selection.height;
    const croppedCtx = croppedCanvas.getContext('2d');
    if (!croppedCtx) return;

    // 从原始图片裁剪选中区域
    croppedCtx.drawImage(
      img,
      selection.x, selection.y, selection.width, selection.height,
      0, 0, selection.width, selection.height
    );

    // 转换为 base64
    const croppedDataUrl = croppedCanvas.toDataURL('image/png');

    onSelectionComplete(croppedDataUrl, selection);
  }, [hasSelection, selection, onSelectionComplete]);

  // 键盘事件处理
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        e.preventDefault();
        onCancel();
      } else if (e.key === 'Enter' && hasSelection) {
        e.preventDefault();
        handleConfirm();
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [hasSelection, onCancel, handleConfirm]);

  // 选择开始
  const handleSelectionStart = useCallback(() => {
    setHasSelection(false);
  }, []);

  // 选择变化
  const handleSelectionChange = useCallback((newSelection: SelectionRect) => {
    setSelection(newSelection);
  }, []);

  // 选择完成
  const handleSelectionComplete = useCallback((newSelection: SelectionRect) => {
    setSelection(newSelection);
    setHasSelection(true);
  }, []);

  return (
    <div className="screenshot-editor-layers">
      {/* 第1层：遮罩层 - z-index: 10000 */}
      <MaskLayer
        imageSrc={imageData}
        selection={selection}
        canvasRef={maskCanvasRef}
      />

      {/* 第2层：交互层 - z-index: 10001 */}
      <InteractionLayer
        enabled={true}
        canvasRef={interactionCanvasRef}
        imageSize={imageSize}
        onSelectionStart={handleSelectionStart}
        onSelectionChange={handleSelectionChange}
        onSelectionComplete={handleSelectionComplete}
        onDoubleClick={handleConfirm}
      />

      {/* 第3层：工具层 - z-index: 10002 */}
      <ToolbarLayer
        visible={true}
        hasSelection={hasSelection}
        selection={selection}
        onConfirm={handleConfirm}
        onCancel={onCancel}
      />
    </div>
  );
};

export default ScreenshotEditorLayers;
