import React, { useEffect, useState, useCallback, useRef } from 'react';
import { Events } from '@wailsio/runtime';
import * as Screenshot2Service from '../../../bindings/ltools/plugins/screenshot2/screenshot2service';
import './styles.css';

// 选区类型
interface Selection {
  x: number;
  y: number;
  width: number;
  height: number;
}

// 显示器信息
interface DisplayInfo {
  index: number;
  width: number;
  height: number;
  x: number;
  y: number;
  primary: boolean;
  name: string;
  scaleFactor: number;
}

// 拖拽手柄类型
type HandleType = 'nw' | 'n' | 'ne' | 'e' | 'se' | 's' | 'sw' | 'w' | null;

// 工具栏组件
const Toolbar: React.FC<{
  selection: Selection | null;
  canvasWidth: number;
  canvasHeight: number;
  scaleFactor: number;
  onCopy: () => void;
  onSave: () => void;
  onCancel: () => void;
}> = ({ selection, canvasWidth, canvasHeight, scaleFactor, onCopy, onSave, onCancel }) => {
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

  // 计算工具栏位置：选区下方居中，但确保不超出屏幕
  const toolbarHeight = 44;
  const toolbarWidth = 120;
  const margin = 10;

  // 默认放在选区下方居中（使用逻辑坐标）
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
      className="screenshot2-toolbar"
      style={{
        position: 'absolute',
        left: toolbarX,
        top: toolbarY,
        transform: 'translateX(-50%)',
      }}
    >
      <button onClick={onCopy} title="复制到剪贴板 (Enter)">
        <span className="icon">📋</span>
      </button>
      <button onClick={onSave} title="保存文件">
        <span className="icon">💾</span>
      </button>
      <button onClick={onCancel} title="取消 (ESC)">
        <span className="icon">✕</span>
      </button>
    </div>
  );
};

// Toast 组件
const Toast: React.FC<{ message: string; visible: boolean }> = ({ message, visible }) => {
  if (!visible) return null;
  return (
    <div className="screenshot2-toast">
      {message}
    </div>
  );
};

// 主覆盖层组件
const Screenshot2Overlay: React.FC = () => {
  const [_sessionId, setSessionId] = useState<string>('');
  const [_displays, setDisplays] = useState<DisplayInfo[]>([]);
  const [imageData, setImageData] = useState<string>('');
  const [selection, setSelection] = useState<Selection | null>(null);
  const [isSelecting, setIsSelecting] = useState(false);
  const [isDraggingHandle, setIsDraggingHandle] = useState(false);
  const [activeHandle, setActiveHandle] = useState<HandleType>(null);
  const [startPos, setStartPos] = useState({ x: 0, y: 0 });
  const [originalSelection, setOriginalSelection] = useState<Selection | null>(null);
  const [toast, setToast] = useState({ message: '', visible: false });

  const canvasRef = useRef<HTMLCanvasElement>(null);
  const imageRef = useRef<HTMLImageElement | null>(null);
  const selectionRef = useRef<Selection | null>(null);
  const [canvasSize, setCanvasSize] = useState({ width: 0, height: 0 });

  // 从 URL 获取参数
  const getDisplayIndex = (): number => {
    const params = new URLSearchParams(window.location.search);
    return parseInt(params.get('display') || '0', 10);
  };

  const getScaleFactor = (): number => {
    const params = new URLSearchParams(window.location.search);
    return parseFloat(params.get('scale') || '1.0');
  };

  const currentDisplayIndex = getDisplayIndex();
  const scaleFactor = getScaleFactor();

  // 显示 Toast
  const showToast = useCallback((message: string) => {
    setToast({ message, visible: true });
    setTimeout(() => setToast({ message: '', visible: false }), 2000);
  }, []);

  // 初始化事件监听
  useEffect(() => {
    console.log('[Screenshot2Overlay] Component mounted, registering event listeners...');

    const unsubscribeSessionStart = Events.On('screenshot2:session-start', (ev: any) => {
      console.log('[Screenshot2Overlay] Session started:', ev.data);
      setSessionId(ev.data);
      setSelection(null);
      setIsSelecting(false);
    });

    const unsubscribeDisplaysInfo = Events.On('screenshot2:displays-info', (ev: any) => {
      try {
        const displaysData = JSON.parse(ev.data);
        setDisplays(displaysData);
      } catch (e) {
        console.error('[Screenshot2Overlay] Failed to parse displays info:', e);
      }
    });

    const unsubscribeImageData = Events.On('screenshot2:image-data', (ev: any) => {
      try {
        const data = JSON.parse(ev.data);
        if (data.displayIndex === currentDisplayIndex) {
          setImageData(data.imageData);
          setSessionId(data.sessionId);
        }
      } catch (e) {
        console.error('[Screenshot2Overlay] Failed to parse image data:', e);
      }
    });

    const unsubscribeSessionEnd = Events.On('screenshot2:session-end', () => {
      setSessionId('');
      setImageData('');
      setSelection(null);
    });

    // 监听其他窗口的选区开始事件，清除本窗口选区
    const unsubscribeSelectionStarted = Events.On('screenshot2:selection-started', (ev: any) => {
      const sourceDisplayIndex = ev.data;
      if (sourceDisplayIndex !== currentDisplayIndex) {
        console.log('[Screenshot2Overlay] Clearing selection, selection started on display', sourceDisplayIndex);
        setSelection(null);
        setIsSelecting(false);
        setIsDraggingHandle(false);
        setActiveHandle(null);
      }
    });

    return () => {
      unsubscribeSessionStart();
      unsubscribeDisplaysInfo();
      unsubscribeImageData();
      unsubscribeSessionEnd();
      unsubscribeSelectionStarted();
    };
  }, [currentDisplayIndex]);

  // 加载图片
  useEffect(() => {
    if (imageData && canvasRef.current) {
      const img = new Image();
      img.onload = () => {
        imageRef.current = img;
        const canvas = canvasRef.current!;
        const ctx = canvas.getContext('2d')!;
        canvas.width = img.width;
        canvas.height = img.height;
        ctx.drawImage(img, 0, 0);
        setCanvasSize({ width: img.width, height: img.height });
        console.log('[Screenshot2Overlay] Image loaded: ', img.width, 'x', img.height);
      };
      img.src = imageData;
    }
  }, [imageData]);

  // 同步 selection 到 ref（用于双击等事件处理）
  useEffect(() => {
    selectionRef.current = selection;
  }, [selection]);

  // 获取手柄位置
  const getHandlePositions = useCallback((sel: Selection): Record<string, { x: number; y: number }> => {
    return {
      nw: { x: sel.x, y: sel.y },
      n: { x: sel.x + sel.width / 2, y: sel.y },
      ne: { x: sel.x + sel.width, y: sel.y },
      e: { x: sel.x + sel.width, y: sel.y + sel.height / 2 },
      se: { x: sel.x + sel.width, y: sel.y + sel.height },
      s: { x: sel.x + sel.width / 2, y: sel.y + sel.height },
      sw: { x: sel.x, y: sel.y + sel.height },
      w: { x: sel.x, y: sel.y + sel.height / 2 },
    };
  }, []);

  // 检测点击的手柄
  const detectHandle = useCallback((x: number, y: number, sel: Selection): HandleType => {
    const handleSize = 12; // 点击检测范围
    const handles = getHandlePositions(sel);

    for (const [type, pos] of Object.entries(handles)) {
      if (type === 'null') continue;
      if (Math.abs(x - pos.x) <= handleSize && Math.abs(y - pos.y) <= handleSize) {
        return type as HandleType;
      }
    }
    return null;
  }, [getHandlePositions]);

  // 获取物理坐标
  const getPhysicalCoords = useCallback((e: React.MouseEvent): { x: number; y: number } => {
    const rect = canvasRef.current?.getBoundingClientRect();
    if (!rect) return { x: 0, y: 0 };
    const logicalX = e.clientX - rect.left;
    const logicalY = e.clientY - rect.top;
    return {
      x: logicalX * scaleFactor,
      y: logicalY * scaleFactor,
    };
  }, [scaleFactor]);

  // 鼠标按下
  const handleMouseDown = useCallback(async (e: React.MouseEvent) => {
    if (e.button !== 0) return;

    // 先聚焦当前窗口
    try {
      await Screenshot2Service.FocusDisplayWindow(currentDisplayIndex);
    } catch (err) {
      console.error('[Screenshot2Overlay] Failed to focus window:', err);
    }

    const { x, y } = getPhysicalCoords(e);

    // 广播选区开始事件，通知其他窗口清除选区
    Events.Emit('screenshot2:selection-started', currentDisplayIndex);

    // 如果已有选区，检查是否点击了手柄
    if (selection && selection.width > 0 && selection.height > 0) {
      const handle = detectHandle(x, y, selection);
      if (handle) {
        setIsDraggingHandle(true);
        setActiveHandle(handle);
        setStartPos({ x, y });
        setOriginalSelection({ ...selection });
        return;
      }
    }

    // 开始新的选区
    setIsSelecting(true);
    setStartPos({ x, y });
    setSelection({ x, y, width: 0, height: 0 });
  }, [selection, detectHandle, getPhysicalCoords]);

  // 鼠标移动
  const handleMouseMove = useCallback((e: React.MouseEvent) => {
    const { x, y } = getPhysicalCoords(e);

    // 更新鼠标样式
    if (selection && selection.width > 0 && selection.height > 0) {
      const handle = detectHandle(x, y, selection);
      if (handle) {
        const cursorMap: Record<string, string> = {
          nw: 'nwse-resize', se: 'nwse-resize',
          ne: 'nesw-resize', sw: 'nesw-resize',
          n: 'ns-resize', s: 'ns-resize',
          e: 'ew-resize', w: 'ew-resize',
        };
        canvasRef.current!.style.cursor = cursorMap[handle] || 'crosshair';
      } else {
        canvasRef.current!.style.cursor = 'crosshair';
      }
    }

    // 拖拽手柄调整选区
    if (isDraggingHandle && activeHandle && originalSelection) {
      const dx = x - startPos.x;
      const dy = y - startPos.y;

      let newSel = { ...originalSelection };

      switch (activeHandle) {
        case 'nw':
          newSel.x = originalSelection.x + dx;
          newSel.y = originalSelection.y + dy;
          newSel.width = originalSelection.width - dx;
          newSel.height = originalSelection.height - dy;
          break;
        case 'n':
          newSel.y = originalSelection.y + dy;
          newSel.height = originalSelection.height - dy;
          break;
        case 'ne':
          newSel.y = originalSelection.y + dy;
          newSel.width = originalSelection.width + dx;
          newSel.height = originalSelection.height - dy;
          break;
        case 'e':
          newSel.width = originalSelection.width + dx;
          break;
        case 'se':
          newSel.width = originalSelection.width + dx;
          newSel.height = originalSelection.height + dy;
          break;
        case 's':
          newSel.height = originalSelection.height + dy;
          break;
        case 'sw':
          newSel.x = originalSelection.x + dx;
          newSel.width = originalSelection.width - dx;
          newSel.height = originalSelection.height + dy;
          break;
        case 'w':
          newSel.x = originalSelection.x + dx;
          newSel.width = originalSelection.width - dx;
          break;
      }

      // 规范化选区（处理负宽高）
      if (newSel.width < 0) {
        newSel.x += newSel.width;
        newSel.width = -newSel.width;
      }
      if (newSel.height < 0) {
        newSel.y += newSel.height;
        newSel.height = -newSel.height;
      }

      setSelection(newSel);
      return;
    }

    // 创建新选区
    if (!isSelecting) return;

    const selX = Math.min(startPos.x, x);
    const selY = Math.min(startPos.y, y);
    const selWidth = Math.abs(x - startPos.x);
    const selHeight = Math.abs(y - startPos.y);

    setSelection({ x: selX, y: selY, width: selWidth, height: selHeight });
  }, [isSelecting, isDraggingHandle, activeHandle, startPos, selection, originalSelection, detectHandle, getPhysicalCoords]);

  // 鼠标松开
  const handleMouseUp = useCallback(() => {
    setIsSelecting(false);
    setIsDraggingHandle(false);
    setActiveHandle(null);
    setOriginalSelection(null);
  }, []);

  // 鼠标进入窗口时聚焦该窗口
  const handleMouseEnter = useCallback(async () => {
    console.log('[Screenshot2Overlay] Mouse entered, focusing window for display', currentDisplayIndex);
    try {
      await Screenshot2Service.FocusDisplayWindow(currentDisplayIndex);
    } catch (e) {
      console.error('[Screenshot2Overlay] Failed to focus window:', e);
    }
  }, [currentDisplayIndex]);

  // 双击复制 - 使用 ref 获取最新选区状态
  const handleDoubleClick = useCallback(async () => {
    const currentSelection = selectionRef.current;
    const currentImage = imageRef.current;

    console.log('[Screenshot2Overlay] Double click detected, selection:', currentSelection);

    if (!currentSelection || currentSelection.width < 10 || currentSelection.height < 10) {
      console.log('[Screenshot2Overlay] Selection too small or null');
      return;
    }

    if (!currentImage) {
      console.log('[Screenshot2Overlay] No image loaded');
      return;
    }

    try {
      const cropCanvas = document.createElement('canvas');
      cropCanvas.width = currentSelection.width;
      cropCanvas.height = currentSelection.height;
      const ctx = cropCanvas.getContext('2d')!;

      // 单屏幕模式：直接从当前屏幕截图裁剪
      ctx.drawImage(
        currentImage,
        currentSelection.x, currentSelection.y, currentSelection.width, currentSelection.height,
        0, 0, currentSelection.width, currentSelection.height
      );

      const base64Data = cropCanvas.toDataURL('image/png');
      console.log('[Screenshot2Overlay] Copying to clipboard...');
      await Screenshot2Service.CopyToClipboard(base64Data);

      showToast('已复制到剪贴板');

      setTimeout(async () => {
        await Screenshot2Service.CancelCapture();
      }, 500);
    } catch (e) {
      console.error('[Screenshot2Overlay] Double-click copy failed:', e);
      showToast('复制失败');
    }
  }, [showToast]);

  // 复制到剪贴板
  const handleCopy = useCallback(async () => {
    if (!selection || !imageRef.current) return;

    try {
      const cropCanvas = document.createElement('canvas');
      cropCanvas.width = selection.width;
      cropCanvas.height = selection.height;
      const ctx = cropCanvas.getContext('2d')!;
      ctx.drawImage(
        imageRef.current,
        selection.x, selection.y, selection.width, selection.height,
        0, 0, selection.width, selection.height
      );

      const base64Data = cropCanvas.toDataURL('image/png');
      await Screenshot2Service.CopyToClipboard(base64Data);

      showToast('已复制到剪贴板');

      // 延迟关闭，让用户看到提示
      setTimeout(async () => {
        await Screenshot2Service.CancelCapture();
      }, 500);
    } catch (e) {
      console.error('[Screenshot2Overlay] Copy failed:', e);
      showToast('复制失败');
    }
  }, [selection, showToast]);

  // 保存文件
  const handleSave = useCallback(async () => {
    if (!selection || !imageRef.current) return;

    try {
      const cropCanvas = document.createElement('canvas');
      cropCanvas.width = selection.width;
      cropCanvas.height = selection.height;
      const ctx = cropCanvas.getContext('2d')!;
      ctx.drawImage(
        imageRef.current,
        selection.x, selection.y, selection.width, selection.height,
        0, 0, selection.width, selection.height
      );

      const base64Data = cropCanvas.toDataURL('image/png');
      await Screenshot2Service.SaveImageWithDialog(base64Data);
      await Screenshot2Service.CancelCapture();
    } catch (e) {
      console.error('[Screenshot2Overlay] Save failed:', e);
    }
  }, [selection]);

  // 取消
  const handleCancel = useCallback(async () => {
    try {
      await Screenshot2Service.CancelCapture();
    } catch (e) {
      console.error('[Screenshot2Overlay] Cancel failed:', e);
    }
  }, []);

  // 绘制遮罩和选区
  useEffect(() => {
    if (!canvasRef.current || !imageRef.current) return;

    const canvas = canvasRef.current;
    const ctx = canvas.getContext('2d')!;
    const img = imageRef.current;

    ctx.clearRect(0, 0, canvas.width, canvas.height);
    ctx.drawImage(img, 0, 0);
    ctx.fillStyle = 'rgba(0, 0, 0, 0.4)';
    ctx.fillRect(0, 0, canvas.width, canvas.height);

    if (selection && selection.width > 0 && selection.height > 0) {
      ctx.clearRect(selection.x, selection.y, selection.width, selection.height);
      ctx.drawImage(
        img,
        selection.x, selection.y, selection.width, selection.height,
        selection.x, selection.y, selection.width, selection.height
      );

      // 选区边框
      ctx.strokeStyle = '#00a8ff';
      ctx.lineWidth = 2;
      ctx.setLineDash([]);
      ctx.strokeRect(selection.x, selection.y, selection.width, selection.height);

      // 8个调整手柄
      const handleSize = 8;
      ctx.fillStyle = '#00a8ff';
      const handles = getHandlePositions(selection);
      Object.entries(handles).forEach(([, pos]) => {
        ctx.fillRect(pos.x - handleSize / 2, pos.y - handleSize / 2, handleSize, handleSize);
      });

      // 尺寸标签
      ctx.fillStyle = '#00a8ff';
      ctx.font = '12px sans-serif';
      const sizeText = `${Math.round(selection.width)} × ${Math.round(selection.height)}`;
      ctx.fillText(sizeText, selection.x + selection.width / 2 - ctx.measureText(sizeText).width / 2, selection.y - 5);
    }
  }, [selection, imageData, getHandlePositions]);

  // 键盘事件
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        handleCancel();
      } else if (e.key === 'Enter' && selection && selection.width > 0 && selection.height > 0) {
        handleCopy();
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [handleCancel, handleCopy, selection]);

  if (!imageData) {
    return (
      <div className="screenshot2-overlay loading">
        <div className="loading-text">正在加载截图...</div>
      </div>
    );
  }

  return (
    <div className="screenshot2-overlay">
      <canvas
        ref={canvasRef}
        onMouseDown={handleMouseDown}
        onMouseMove={handleMouseMove}
        onMouseUp={handleMouseUp}
        onMouseLeave={handleMouseUp}
        onMouseEnter={handleMouseEnter}
        onDoubleClick={handleDoubleClick}
      />
      {selection && (
        <Toolbar
          selection={selection}
          canvasWidth={canvasSize.width}
          canvasHeight={canvasSize.height}
          scaleFactor={scaleFactor}
          onCopy={handleCopy}
          onSave={handleSave}
          onCancel={handleCancel}
        />
      )}
      <Toast message={toast.message} visible={toast.visible} />
    </div>
  );
};

export default Screenshot2Overlay;
