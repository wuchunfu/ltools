import React, { useState, useRef, useEffect, useCallback } from 'react';
import './ScreenshotSelector.css';

interface ScreenshotSelectorProps {
  imageData: string;
  onSelectionComplete: (selectedImageData: string, bounds: { x: number; y: number; width: number; height: number }) => void;
  onCancel: () => void;
  onDebugLog?: (message: string) => void;  // 添加调试日志回调
}

const ScreenshotSelector: React.FC<ScreenshotSelectorProps> = ({
  imageData,
  onSelectionComplete,
  onCancel,
  onDebugLog  // 接收调试日志回调
}) => {
  const [isSelecting, setIsSelecting] = useState(false);
  const [selection, setSelection] = useState({ x: 0, y: 0, width: 0, height: 0 });
  const [hasSelection, setHasSelection] = useState(false);
  const [startPos, setStartPos] = useState({ x: 0, y: 0 });
  const clickTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const clickCountRef = useRef(0);
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const imageRef = useRef<HTMLImageElement>(null);

  // 调试日志辅助函数
  const logDebug = useCallback((message: string) => {
    console.log('[ScreenshotSelector]', message);
    if (onDebugLog) {
      onDebugLog(message);
    }
  }, [onDebugLog]);

  // 组件挂载时初始化
  useEffect(() => {
    logDebug('=== ScreenshotSelector 组件挂载 ===');
    logDebug(`图片数据长度: ${imageData.length}`);
    logDebug(`当前 URL: ${window.location.href}`);

    // 确保窗口获得焦点
    window.focus();
    logDebug('已调用 window.focus()');

    // 设置窗口标题以便调试
    document.title = 'Screenshot Selector - 按 ESC 取消';
  }, [imageData, logDebug]);

  // 确认选择 - 必须在 useEffect 之前定义
  const handleConfirm = useCallback(() => {
    logDebug('handleConfirm 被调用');
    if (selection.width === 0 || selection.height === 0) {
      return;
    }

    // 裁剪选中区域的图片
    const canvas = canvasRef.current;
    const image = imageRef.current;
    if (!canvas || !image) return;

    logDebug(`裁剪区域: x=${selection.x}, y=${selection.y}, w=${selection.width}, h=${selection.height}`);

    // 创建新的画布来存储裁剪后的图片
    const croppedCanvas = document.createElement('canvas');
    croppedCanvas.width = selection.width;
    croppedCanvas.height = selection.height;
    const croppedCtx = croppedCanvas.getContext('2d');
    if (!croppedCtx) return;

    // 从原始图片裁剪选中区域（坐标是1:1的）
    croppedCtx.drawImage(
      image,
      selection.x, selection.y, selection.width, selection.height,
      0, 0, selection.width, selection.height
    );

    // 转换为 base64
    const croppedDataUrl = croppedCanvas.toDataURL('image/png');

    logDebug('调用 onSelectionComplete');
    onSelectionComplete(croppedDataUrl, selection);
  }, [selection, onSelectionComplete, logDebug]);

  // 图片加载和 Canvas 初始化 - 只在 imageData 变化时运行
  useEffect(() => {
    const canvas = canvasRef.current;
    const image = imageRef.current;
    if (canvas && image) {
      image.onload = () => {
        logDebug(`=== 图片加载完成 ===`);
        logDebug(`图片原始尺寸: ${image.width} x ${image.height}`);
        logDebug(`窗口尺寸: ${window.innerWidth} x ${window.innerHeight}`);

        const ctx = canvas.getContext('2d', {
          alpha: false
        });
        if (ctx) {
          // Canvas 内部尺寸使用图片尺寸（原始截图分辨率）
          canvas.width = image.width;
          canvas.height = image.height;

          // CSS 显示尺寸使用窗口尺寸
          canvas.style.width = window.innerWidth + 'px';
          canvas.style.height = window.innerHeight + 'px';
          canvas.style.position = 'fixed';
          canvas.style.top = '0';
          canvas.style.left = '0';

          // 关键修复：禁用图像平滑，使用像素化渲染
          // 这样在 Retina 屏幕上会保持清晰
          ctx.imageSmoothingEnabled = false;
          (ctx as any).mozImageSmoothingEnabled = false;
          (ctx as any).webkitImageSmoothingEnabled = false;
          (ctx as any).msImageSmoothingEnabled = false;

          // 绘制图片，不进行任何缩放处理
          ctx.drawImage(image, 0, 0, canvas.width, canvas.height);

          // 绘制半透明遮罩 - 0.1透明度
          ctx.fillStyle = 'rgba(0, 0, 0, 0.1)';
          ctx.fillRect(0, 0, canvas.width, canvas.height);

          logDebug(`=== Canvas 初始化完成 ===`);
          logDebug(`Canvas 内部尺寸: ${canvas.width} x ${canvas.height}`);
          logDebug(`CSS 显示尺寸: ${window.innerWidth} x ${window.innerHeight}`);
          logDebug(`图像平滑: 禁用（像素化渲染）`);
        }
      };

      // 设置图片源
      image.src = imageData;
    }
  }, [imageData, logDebug]);

  // 键盘事件处理 - 使用 ref 存储最新的回调，避免频繁重新添加监听器
  useEffect(() => {
    // 键盘事件
    const handleKeyDown = (e: KeyboardEvent) => {
      console.log('[ScreenshotSelector] 键盘事件:', e.key);
      if (e.key === 'Escape') {
        console.log('[ScreenshotSelector] 检测到 ESC 键，取消截图');
        e.preventDefault();
        e.stopPropagation();
        onCancel();
      } else if (e.key === 'Enter' && hasSelection && selection.width > 0 && selection.height > 0) {
        console.log('[ScreenshotSelector] 检测到 Enter 键，确认选择');
        e.preventDefault();
        e.stopPropagation();
        handleConfirm();
      }
    };

    window.addEventListener('keydown', handleKeyDown, { capture: true }); // 使用 capture 确保事件被捕获
    console.log('[ScreenshotSelector] 键盘监听器已添加 (capture mode)');

    return () => {
      window.removeEventListener('keydown', handleKeyDown);
      if (clickTimerRef.current) {
        clearTimeout(clickTimerRef.current);
      }
      console.log('[ScreenshotSelector] 键盘监听器已移除');
    };
  }, [hasSelection, selection, onCancel, handleConfirm]);  // 移除 logDebug 依赖

  // 重绘画布
  const redrawCanvas = useCallback(() => {
    const canvas = canvasRef.current;
    const image = imageRef.current;
    if (!canvas || !image) return;

    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    // 禁用图像平滑，使用像素化渲染
    ctx.imageSmoothingEnabled = false;
    (ctx as any).mozImageSmoothingEnabled = false;
    (ctx as any).webkitImageSmoothingEnabled = false;
    (ctx as any).msImageSmoothingEnabled = false;

    // 清空并绘制原始图片
    ctx.clearRect(0, 0, canvas.width, canvas.height);
    ctx.drawImage(image, 0, 0, canvas.width, canvas.height);

    // 绘制半透明遮罩 - 0.1透明度
    ctx.fillStyle = 'rgba(0, 0, 0, 0.1)';
    ctx.fillRect(0, 0, canvas.width, canvas.height);

    // 如果有选择区域，清除遮罩并绘制选中区域
    if (selection.width > 0 && selection.height > 0) {
      const x = Math.floor(selection.x);
      const y = Math.floor(selection.y);
      const w = Math.floor(selection.width);
      const h = Math.floor(selection.height);

      // 清除选中区域的遮罩
      ctx.clearRect(x, y, w, h);

      // 重新绘制选中区域的原始图片
      ctx.drawImage(image, x, y, w, h, x, y, w, h);

      // 绘制选择边框
      ctx.strokeStyle = '#4CD964';
      ctx.lineWidth = 3;
      ctx.lineJoin = 'round';
      ctx.lineCap = 'round';
      ctx.strokeRect(x, y, w, h);

      // 绘制尺寸信息
      const sizeText = `${w} x ${h}`;
      ctx.font = 'bold 16px monospace';
      const textWidth = ctx.measureText(sizeText).width;

      // 背景框
      const bgX = x;
      const bgY = Math.max(0, y - 28);
      ctx.fillStyle = '#4CD964';
      ctx.fillRect(bgX, bgY, textWidth + 16, 24);

      // 文字
      ctx.fillStyle = '#000';
      ctx.textBaseline = 'middle';
      ctx.fillText(sizeText, x + 8, bgY + 12);
    }
  }, [selection]);

  useEffect(() => {
    redrawCanvas();
  }, [redrawCanvas]);

  // 处理鼠标按下
  const handleMouseDown = (e: React.MouseEvent<HTMLCanvasElement>) => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const rect = canvas.getBoundingClientRect();

    // 计算鼠标在Canvas内部的坐标
    const mouseX = e.clientX - rect.left;
    const mouseY = e.clientY - rect.top;

    // Canvas内部尺寸可能与显示尺寸不同，需要计算缩放比例
    const scaleX = canvas.width / rect.width;
    const scaleY = canvas.height / rect.height;

    // 转换为Canvas内部坐标
    const x = Math.round(mouseX * scaleX);
    const y = Math.round(mouseY * scaleY);

    logDebug(`🖱️ 鼠标按下: screen(${mouseX.toFixed(0)}, ${mouseY.toFixed(0)}) → canvas(${x}, ${y})`);
    logDebug(`   缩放比例: X=${scaleX.toFixed(2)}, Y=${scaleY.toFixed(2)}`);

    // 如果点击在选择区域外，重新开始选择
    const isInsideSelection =
      selection.width > 0 &&
      selection.height > 0 &&
      x >= selection.x &&
      x <= selection.x + selection.width &&
      y >= selection.y &&
      y <= selection.y + selection.height;

    if (!isInsideSelection || !hasSelection) {
      setIsSelecting(true);
      setHasSelection(false);
      setStartPos({ x, y });
      setSelection({ x, y, width: 0, height: 0 });
      logDebug(`   开始选择，起点: (${x}, ${y})`);
    }
  };

  // 处理鼠标移动
  const handleMouseMove = (e: React.MouseEvent<HTMLCanvasElement>) => {
    if (!isSelecting) return;

    const canvas = canvasRef.current;
    if (!canvas) return;

    const rect = canvas.getBoundingClientRect();

    // 计算鼠标在Canvas内部的坐标
    const mouseX = e.clientX - rect.left;
    const mouseY = e.clientY - rect.top;

    // Canvas内部尺寸可能与显示尺寸不同，需要计算缩放比例
    const scaleX = canvas.width / rect.width;
    const scaleY = canvas.height / rect.height;

    // 转换为Canvas内部坐标
    const x = Math.round(mouseX * scaleX);
    const y = Math.round(mouseY * scaleY);

    const width = x - startPos.x;
    const height = y - startPos.y;

    const newSelection = {
      x: width > 0 ? startPos.x : x,
      y: height > 0 ? startPos.y : y,
      width: Math.abs(width),
      height: Math.abs(height)
    };

    logDebug(`🖱️ 鼠标移动: ${newSelection.width} x ${newSelection.height} at (${newSelection.x}, ${newSelection.y})`);

    setSelection(newSelection);
  };

  // 处理鼠标释放
  const handleMouseUp = () => {
    if (isSelecting) {
      setIsSelecting(false);
      // 只有当选择区域大于一定尺寸时才标记为有选择
      if (selection.width > 10 && selection.height > 10) {
        setHasSelection(true);
        logDebug(`✅ 选择完成: ${selection.width} x ${selection.height}`);
      } else {
        logDebug(`⚠️ 选择区域太小 (${selection.width} x ${selection.height})，已忽略`);
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

    // 设置新的定时器，如果在 300ms 内没有第二次点击，则重置计数
    clickTimerRef.current = setTimeout(() => {
      clickCountRef.current = 0;
    }, 300);

    // 如果是第二次点击（双击）
    if (clickCountRef.current === 2 && hasSelection && selection.width > 0 && selection.height > 0) {
      handleConfirm();
      clickCountRef.current = 0;
    }
  };

  return (
    <div className="screenshot-selector">
      <img ref={imageRef} src={imageData} alt="Screenshot" style={{ display: 'none' }} />
      <canvas
        ref={canvasRef}
        onMouseDown={handleMouseDown}
        onMouseMove={handleMouseMove}
        onMouseUp={handleMouseUp}
        onClick={handleClick}
      />

      {/* 提示信息 */}
      {!hasSelection && (
        <div className="selector-hint">
          拖拽选择截图范围
        </div>
      )}

      {/* 工具栏 */}
      {hasSelection && selection.width > 0 && selection.height > 0 && (
        <>
          <div className="selector-toolbar">
            <button onClick={handleConfirm} className="confirm-button">
              ✓ 确认
            </button>
            <button onClick={onCancel} className="cancel-button">
              ✕ 取消
            </button>
          </div>
          <div className="selector-hint">
            双击或 Enter 确认 | 拖拽重新选择 | ESC 取消
          </div>
        </>
      )}
    </div>
  );
};

export default ScreenshotSelector;
