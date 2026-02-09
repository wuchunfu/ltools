import React, { useState, useRef, useEffect, useCallback } from 'react';
import './InteractionLayer.css';

interface InteractionLayerProps {
  enabled: boolean;
  canvasRef: React.RefObject<HTMLCanvasElement>;
  imageSize: { width: number; height: number };
  onSelectionStart?: () => void;
  onSelectionChange?: (selection: SelectionRect) => void;
  onSelectionComplete?: (selection: SelectionRect) => void;
  onDoubleClick?: () => void;
}

export interface SelectionRect {
  x: number;
  y: number;
  width: number;
  height: number;
}

/**
 * 交互层 - 负责处理所有用户交互
 * 职责：
 * 1. 监听鼠标/触摸事件（按下、移动、释放）
 * 2. 计算选择区域的坐标和尺寸
 * 3. 绘制选择边框和尺寸信息
 * 4. 处理双击确认
 *
 * 修复版本：
 * - 使用外部传入的 canvasRef
 * - 接收 imageSize 来确保尺寸正确
 * - 修复了绘制逻辑
 */
export const InteractionLayer: React.FC<InteractionLayerProps> = ({
  enabled,
  canvasRef,
  imageSize,
  onSelectionStart,
  onSelectionChange,
  onSelectionComplete,
  onDoubleClick
}) => {
  const [isSelecting, setIsSelecting] = useState(false);
  const [selection, setSelection] = useState<SelectionRect>({ x: 0, y: 0, width: 0, height: 0 });
  const [hasSelection, setHasSelection] = useState(false);

  const startPosRef = useRef<{ x: number; y: number }>({ x: 0, y: 0 });
  const clickCountRef = useRef(0);
  const clickTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  // 绘制选择框和尺寸信息
  const drawSelection = useCallback(() => {
    const canvas = canvasRef.current;
    if (!canvas || selection.width === 0 || selection.height === 0) return;

    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    // 清除之前的绘制
    ctx.clearRect(0, 0, canvas.width, canvas.height);

    // 绘制选择边框
    ctx.strokeStyle = '#4CD964';
    ctx.lineWidth = 2;
    ctx.strokeRect(selection.x, selection.y, selection.width, selection.height);

    // 绘制尺寸信息
    const sizeText = `${Math.abs(selection.width)} x ${Math.abs(selection.height)}`;
    ctx.font = '14px monospace';
    const textWidth = ctx.measureText(sizeText).width;

    // 尺寸标签背景
    ctx.fillStyle = '#4CD964';
    ctx.fillRect(selection.x, selection.y - 24, textWidth + 12, 20);

    // 尺寸文本
    ctx.fillStyle = '#000';
    ctx.fillText(sizeText, selection.x + 6, selection.y - 10);
  }, [canvasRef, selection]);

  useEffect(() => {
    drawSelection();
  }, [drawSelection]);

  // 处理鼠标按下
  const handleMouseDown = (e: React.MouseEvent<HTMLCanvasElement>) => {
    if (!enabled) return;

    const canvas = canvasRef.current;
    if (!canvas) return;

    const rect = canvas.getBoundingClientRect();
    const scaleX = canvas.width / rect.width;
    const scaleY = canvas.height / rect.height;

    const x = (e.clientX - rect.left) * scaleX;
    const y = (e.clientY - rect.top) * scaleY;

    // 检查是否点击在选择区域内
    const isInsideSelection =
      hasSelection &&
      selection.width > 0 &&
      selection.height > 0 &&
      x >= selection.x &&
      x <= selection.x + selection.width &&
      y >= selection.y &&
      y <= selection.y + selection.height;

    // 如果点击在选择区域外，重新开始选择
    if (!isInsideSelection) {
      setIsSelecting(true);
      setHasSelection(false);
      setSelection({ x, y, width: 0, height: 0 });
      startPosRef.current = { x, y };
      onSelectionStart?.();
    }
  };

  // 处理鼠标移动
  const handleMouseMove = (e: React.MouseEvent<HTMLCanvasElement>) => {
    if (!isSelecting || !enabled) return;

    const canvas = canvasRef.current;
    if (!canvas) return;

    const rect = canvas.getBoundingClientRect();
    const scaleX = canvas.width / rect.width;
    const scaleY = canvas.height / rect.height;

    const x = (e.clientX - rect.left) * scaleX;
    const y = (e.clientY - rect.top) * scaleY;

    const width = x - startPosRef.current.x;
    const height = y - startPosRef.current.y;

    const newSelection: SelectionRect = {
      x: width > 0 ? startPosRef.current.x : x,
      y: height > 0 ? startPosRef.current.y : y,
      width: Math.abs(width),
      height: Math.abs(height)
    };

    setSelection(newSelection);
    onSelectionChange?.(newSelection);
  };

  // 处理鼠标释放
  const handleMouseUp = () => {
    if (isSelecting) {
      setIsSelecting(false);

      // 只有当选择区域大于一定尺寸时才标记为有效选择
      if (selection.width > 10 && selection.height > 10) {
        setHasSelection(true);
        onSelectionComplete?.(selection);
      } else {
        // 选择区域太小，清除选择
        setSelection({ x: 0, y: 0, width: 0, height: 0 });
      }
    }
  };

  // 处理点击（用于检测双击）
  const handleClick = () => {
    clickCountRef.current += 1;

    // 清除之前的定时器
    if (clickTimerRef.current) {
      clearTimeout(clickTimerRef.current);
    }

    // 设置新的定时器
    clickTimerRef.current = setTimeout(() => {
      clickCountRef.current = 0;
    }, 300);

    // 检测双击
    if (clickCountRef.current === 2 && hasSelection) {
      onDoubleClick?.();
      clickCountRef.current = 0;
    }
  };

  // 清理定时器
  useEffect(() => {
    return () => {
      if (clickTimerRef.current) {
        clearTimeout(clickTimerRef.current);
      }
    };
  }, []);

  return (
    <canvas
      ref={canvasRef}
      className="interaction-canvas"
      onMouseDown={handleMouseDown}
      onMouseMove={handleMouseMove}
      onMouseUp={handleMouseUp}
      onClick={handleClick}
      style={{ pointerEvents: enabled ? 'auto' : 'none' }}
    />
  );
};

export default InteractionLayer;
