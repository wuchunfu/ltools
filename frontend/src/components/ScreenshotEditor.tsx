import React, { useEffect, useState, useRef, useCallback } from 'react';
import * as ScreenshotService from '../../bindings/ltools/plugins/screenshot/screenshotservice';
import { Events } from '@wailsio/runtime';
import ScreenshotSelector from './ScreenshotSelector';
import { useToast } from '../hooks/useToast';
import './ScreenshotEditor.css';

// 工具类型
type ToolType =
  | 'select'
  | 'rect'
  | 'ellipse'
  | 'arrow'
  | 'text'
  | 'brush'
  | 'blur'
  | 'mosaic'
  | 'crop';

// 颜色选项
const COLORS = [
  { name: 'red', value: '#FF3B30' },
  { name: 'orange', value: '#FF9500' },
  { name: 'yellow', value: '#FFCC00' },
  { name: 'green', value: '#4CD964' },
  { name: 'blue', value: '#007AFF' },
  { name: 'purple', value: '#5856D6' },
  { name: 'white', value: '#FFFFFF' },
  { name: 'black', value: '#000000' },
];

// 工具配置
const TOOLS: { type: ToolType; icon: string; label: string }[] = [
  { type: 'select', icon: '↖', label: '选择' },
  { type: 'rect', icon: '□', label: '矩形' },
  { type: 'ellipse', icon: '○', label: '椭圆' },
  { type: 'arrow', icon: '→', label: '箭头' },
  { type: 'text', icon: 'T', label: '文字' },
  { type: 'brush', icon: '✎', label: '画笔' },
  { type: 'blur', icon: '◎', label: '模糊' },
  { type: 'mosaic', icon: '▦', label: '马赛克' },
  { type: 'crop', icon: '⤡', label: '裁剪' },
];

interface ScreenshotEditorProps {
  imageData?: string;
  onClose?: () => void;
}

interface Annotation {
  id: string;
  type: ToolType;
  x: number;
  y: number;
  width?: number;
  height?: number;
  color: string;
  lineWidth: number;
  text?: string;
  points?: { x: number; y: number }[];
}

type EditorMode = 'selecting' | 'editing';

const ScreenshotEditor: React.FC<ScreenshotEditorProps> = ({ imageData, onClose }) => {
  const { success, error } = useToast();
  const [mode, setMode] = useState<EditorMode>('selecting');
  const [originalImage, setOriginalImage] = useState<string>('');
  const [currentImage, setCurrentImage] = useState<string>('');
  const [selectedTool, setSelectedTool] = useState<ToolType>('rect');
  const [selectedColor, setSelectedColor] = useState(COLORS[3].value); // 默认绿色
  const [lineWidth, setLineWidth] = useState(3);
  const [annotations, setAnnotations] = useState<Annotation[]>([]);
  const [isDrawing, setIsDrawing] = useState(false);
  const [startPos, setStartPos] = useState({ x: 0, y: 0 });
  const [currentAnnotation, setCurrentAnnotation] = useState<Annotation | null>(null);
  const [debugLogs, setDebugLogs] = useState<string[]>([]);

  const canvasRef = useRef<HTMLCanvasElement>(null);
  const imageRef = useRef<HTMLImageElement>(null);
  const lastSessionIdRef = useRef<string>('');

  // 使用 ref 存储最新的函数引用，避免键盘监听器频繁重建
  const handleCopyAndCloseRef = useRef<(() => Promise<void>) | null>(null);
  const handleCancelRef = useRef<(() => Promise<void>) | null>(null);
  const handleSaveRef = useRef<(() => Promise<void>) | null>(null);
  const handleUndoRef = useRef<(() => void) | null>(null);

  // 添加调试日志函数
  const addLog = useCallback((message: string) => {
    const timestamp = new Date().toLocaleTimeString();
    setDebugLogs(prev => [`[${timestamp}] ${message}`, ...prev].slice(0, 20)); // 只保留最近20条
    console.log(`[ScreenshotEditor] ${message}`);
  }, []);

  // 组件挂载时初始化
  useEffect(() => {
    console.log('[ScreenshotEditor] 组件挂载');
    console.log('[ScreenshotEditor] 当前 URL:', window.location.href);

    // 确保窗口获得焦点
    window.focus();
    console.log('[ScreenshotEditor] 调用 window.focus()');

    // 设置窗口标题以便调试
    document.title = 'Screenshot Editor - 按 ESC 取消';

    // 监听会话开始事件（窗口复用时触发状态重置）
    const unsubscribeSessionStart = Events.On('screenshot:session-start', (event) => {
      const newSessionId = event.data as string;
      console.log('[ScreenshotEditor] 收到 session-start 事件，会话 ID:', newSessionId);
      if (newSessionId && newSessionId !== lastSessionIdRef.current) {
        console.log('[ScreenshotEditor] 新会话，重置所有状态');
        lastSessionIdRef.current = newSessionId;
        // 重置所有状态
        setMode('selecting');
        setAnnotations([]);
        setCurrentAnnotation(null);
        setSelectedTool('rect');
        setIsDrawing(false);
      }
    });

    // 监听会话结束事件（窗口关闭时清理状态）
    const unsubscribeSessionEnd = Events.On('screenshot:session-end', () => {
      console.log('[ScreenshotEditor] 收到 session-end 事件，清理所有状态');
      lastSessionIdRef.current = '';
      // 重置所有状态
      setMode('selecting');
      setOriginalImage('');
      setCurrentImage('');
      setAnnotations([]);
      setCurrentAnnotation(null);
      setSelectedTool('rect');
      setIsDrawing(false);
    });

    // 注意：Wails 窗口已通过 MacWindowLevelScreenSaver 设置为最高级别
    // 不需要请求浏览器全屏，这可能会与 Wails 窗口系统冲突

    return () => {
      unsubscribeSessionStart();
      unsubscribeSessionEnd();
    };
  }, []);

  // 从事件系统接收图片数据
  useEffect(() => {
    console.log('[ScreenshotEditor] 设置图片数据事件监听器');

    const unsubscribe = Events.On('screenshot:image-data', (event) => {
      const data = event.data as string;
      console.log('[ScreenshotEditor] 收到图片数据事件，长度:', data?.length);
      if (data) {
        // 检查是否包含会话 ID（格式：data:image/png;base64,...|sessionId）
        let imageData = data;
        let dataSessionId = '';

        const pipeIndex = data.lastIndexOf('|');
        if (pipeIndex > 0 && pipeIndex < data.length - 1) {
          // 提取会话 ID
          dataSessionId = data.substring(pipeIndex + 1);
          imageData = data.substring(0, pipeIndex);
          console.log('[ScreenshotEditor] 提取到会话 ID:', dataSessionId);

          // 如果会话 ID 与上次不同，重置所有状态
          if (dataSessionId && dataSessionId !== lastSessionIdRef.current) {
            console.log('[ScreenshotEditor] 检测到新会话，重置所有状态');
            lastSessionIdRef.current = dataSessionId;
            setMode('selecting');
            setAnnotations([]);
            setCurrentAnnotation(null);
            setSelectedTool('rect');
            setIsDrawing(false);
          }
        }

        console.log('[ScreenshotEditor] 设置原始图片');
        setOriginalImage(imageData);
        setCurrentImage(imageData);
      } else {
        console.log('[ScreenshotEditor] 警告：图片数据为空！');
      }
    });

    // 同时尝试从 URL 参数获取（用于兼容）
    const urlParams = new URLSearchParams(window.location.search);
    const urlData = urlParams.get('data');
    if (urlData) {
      console.log('[ScreenshotEditor] 从 URL 获取图片数据，长度:', urlData.length);
      setOriginalImage(urlData);
      setCurrentImage(urlData);
    } else if (imageData) {
      console.log('[ScreenshotEditor] 使用传入的图片数据');
      setOriginalImage(imageData);
      setCurrentImage(imageData);
    }

    return () => {
      unsubscribe();
      console.log('[ScreenshotEditor] 图片数据事件监听器已移除');
    };
  }, [imageData]);

  // 处理区域选择完成
  const handleSelectionComplete = useCallback((selectedImageData: string) => {
    setCurrentImage(selectedImageData);
    setMode('editing');
  }, []);

  // 渲染画布
  const renderCanvas = useCallback(() => {
    const canvas = canvasRef.current;
    const image = imageRef.current;
    if (!canvas || !image || mode !== 'editing') return;

    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    // 设置画布尺寸
    canvas.width = image.width;
    canvas.height = image.height;

    // 绘制原始图片
    ctx.drawImage(image, 0, 0);

    // 绘制所有标注
    annotations.forEach(annotation => {
      drawAnnotation(ctx, annotation);
    });

    // 绘制当前正在绘制的标注
    if (currentAnnotation) {
      drawAnnotation(ctx, currentAnnotation);
    }
  }, [annotations, currentAnnotation, mode]);

  // 绘制单个标注
  const drawAnnotation = (ctx: CanvasRenderingContext2D, annotation: Annotation) => {
    ctx.strokeStyle = annotation.color;
    ctx.lineWidth = annotation.lineWidth;
    ctx.fillStyle = annotation.color;

    switch (annotation.type) {
      case 'rect':
        if (annotation.width && annotation.height) {
          ctx.strokeRect(annotation.x, annotation.y, annotation.width, annotation.height);
        }
        break;
      case 'ellipse':
        if (annotation.width && annotation.height) {
          ctx.beginPath();
          ctx.ellipse(
            annotation.x + annotation.width / 2,
            annotation.y + annotation.height / 2,
            Math.abs(annotation.width) / 2,
            Math.abs(annotation.height) / 2,
            0, 0, Math.PI * 2
          );
          ctx.stroke();
        }
        break;
      case 'arrow':
        if (annotation.width && annotation.height) {
          drawArrow(ctx, annotation.x, annotation.y, annotation.x + annotation.width, annotation.y + annotation.height, annotation.lineWidth);
        }
        break;
      case 'brush':
        if (annotation.points && annotation.points.length > 1) {
          ctx.beginPath();
          ctx.moveTo(annotation.points[0].x, annotation.points[0].y);
          annotation.points.forEach(point => {
            ctx.lineTo(point.x, point.y);
          });
          ctx.stroke();
        }
        break;
      case 'text':
        if (annotation.text) {
          ctx.font = `${annotation.lineWidth * 5}px Arial`;
          // 设置文本基线为顶部，让文字从鼠标位置向下绘制
          ctx.textBaseline = 'top';
          ctx.fillText(annotation.text, annotation.x, annotation.y);
          // 重置为默认值（虽然后续绘制会覆盖，但保持状态一致性）
          ctx.textBaseline = 'alphabetic';
        }
        break;
      case 'blur':
        if (annotation.width && annotation.height && annotation.height > 5) {
          // 判断是否是预览状态（当前正在绘制的标注）
          const isPreview = (annotation as any).isPreview === true;
          drawBlurEffect(ctx, annotation.x, annotation.y, annotation.width, annotation.height, isPreview);
        }
        break;
      case 'mosaic':
        if (annotation.width && annotation.height && annotation.height > 5) {
          // 判断是否是预览状态（当前正在绘制的标注）
          const isPreview = (annotation as any).isPreview === true;
          drawMosaicEffect(ctx, annotation.x, annotation.y, annotation.width, annotation.height, isPreview);
        }
        break;
      case 'crop':
        if (annotation.width && annotation.height) {
          // 裁剪工具只显示边框
          ctx.strokeStyle = '#FFFFFF';
          ctx.setLineDash([5, 5]);
          ctx.strokeRect(annotation.x, annotation.y, annotation.width, annotation.height);
          ctx.setLineDash([]);

          // 添加裁剪角标记
          const cornerSize = 10;
          ctx.strokeStyle = annotation.color;
          ctx.lineWidth = 2;

          // 左上角
          ctx.beginPath();
          ctx.moveTo(annotation.x, annotation.y + cornerSize);
          ctx.lineTo(annotation.x, annotation.y);
          ctx.lineTo(annotation.x + cornerSize, annotation.y);
          ctx.stroke();

          // 右上角
          ctx.beginPath();
          ctx.moveTo(annotation.x + annotation.width - cornerSize, annotation.y);
          ctx.lineTo(annotation.x + annotation.width, annotation.y);
          ctx.lineTo(annotation.x + annotation.width, annotation.y + cornerSize);
          ctx.stroke();

          // 右下角
          ctx.beginPath();
          ctx.moveTo(annotation.x + annotation.width, annotation.y + annotation.height - cornerSize);
          ctx.lineTo(annotation.x + annotation.width, annotation.y + annotation.height);
          ctx.lineTo(annotation.x + annotation.width - cornerSize, annotation.y + annotation.height);
          ctx.stroke();

          // 左下角
          ctx.beginPath();
          ctx.moveTo(annotation.x + cornerSize, annotation.y + annotation.height);
          ctx.lineTo(annotation.x, annotation.y + annotation.height);
          ctx.lineTo(annotation.x, annotation.y + annotation.height - cornerSize);
          ctx.stroke();
        }
        break;
    }
  };

  // 绘制模糊效果
  const drawBlurEffect = (
    ctx: CanvasRenderingContext2D,
    x: number,
    y: number,
    width: number,
    height: number,
    isPreview: boolean = false
  ) => {
    // 保存当前上下文状态
    ctx.save();

    const canvas = canvasRef.current;
    if (!canvas) return;

    // 如果是预览模式，只绘制边框和半透明遮罩
    if (isPreview) {
      ctx.strokeStyle = 'rgba(255, 255, 255, 0.8)';
      ctx.lineWidth = 2;
      ctx.setLineDash([5, 5]);
      ctx.strokeRect(x, y, width, height);
      ctx.setLineDash([]);

      // 绘制半透明遮罩表示模糊区域
      ctx.fillStyle = 'rgba(150, 150, 255, 0.3)';
      ctx.fillRect(x, y, width, height);

      // 在区域中心显示模糊预览提示
      ctx.fillStyle = '#FFFFFF';
      ctx.font = '14px Arial';
      ctx.textAlign = 'center';
      ctx.textBaseline = 'middle';
      ctx.fillText('模糊区域', x + width / 2, y + height / 2);

      ctx.restore();
      return;
    }

    // 创建临时画布来处理模糊
    const tempCanvas = document.createElement('canvas');
    const tempCtx = tempCanvas.getContext('2d');
    if (!tempCtx) return;

    tempCanvas.width = Math.abs(width);
    tempCanvas.height = Math.abs(height);

    // 从主画布复制选中区域
    tempCtx.drawImage(
      canvas,
      x, y, Math.abs(width), Math.abs(height),
      0, 0, Math.abs(width), Math.abs(height)
    );

    // 应用更强的模糊滤镜（从 10px 增加到 25px）
    tempCtx.filter = 'blur(25px)';
    tempCtx.drawImage(tempCanvas, 0, 0);

    // 将模糊后的图像绘制回主画布
    ctx.drawImage(tempCanvas, x, y);

    ctx.restore();
  };

  // 绘制马赛克效果
  const drawMosaicEffect = (
    ctx: CanvasRenderingContext2D,
    x: number,
    y: number,
    width: number,
    height: number,
    isPreview: boolean = false
  ) => {
    // 如果是预览模式，只绘制边框和半透明遮罩
    if (isPreview) {
      ctx.strokeStyle = 'rgba(255, 255, 255, 0.8)';
      ctx.lineWidth = 2;
      ctx.setLineDash([5, 5]);
      ctx.strokeRect(x, y, width, height);
      ctx.setLineDash([]);

      // 绘制半透明遮罩表示马赛克区域
      ctx.fillStyle = 'rgba(150, 255, 150, 0.3)';
      ctx.fillRect(x, y, width, height);

      // 在区域中心显示马赛克预览提示
      ctx.fillStyle = '#FFFFFF';
      ctx.font = '14px Arial';
      ctx.textAlign = 'center';
      ctx.textBaseline = 'middle';
      ctx.fillText('马赛克区域', x + width / 2, y + height / 2);

      return;
    }

    const canvas = canvasRef.current;
    if (!canvas) return;

    // 马赛克块大小（根据区域大小动态调整）
    const blockSize = Math.max(10, Math.min(20, Math.floor(Math.max(width, height) / 20)));
    const absWidth = Math.abs(width);
    const absHeight = Math.abs(height);

    // 获取图像数据
    const imageData = ctx.getImageData(
      Math.min(x, x + width),
      Math.min(y, y + height),
      absWidth,
      absHeight
    );

    const data = imageData.data;

    // 对每个块进行像素化处理
    for (let by = 0; by < absHeight; by += blockSize) {
      for (let bx = 0; bx < absWidth; bx += blockSize) {
        // 计算块的平均颜色
        let r = 0, g = 0, b = 0, count = 0;

        for (let py = by; py < by + blockSize && py < absHeight; py++) {
          for (let px = bx; px < bx + blockSize && px < absWidth; px++) {
            const i = (py * absWidth + px) * 4;
            r += data[i];
            g += data[i + 1];
            b += data[i + 2];
            count++;
          }
        }

        r = Math.floor(r / count);
        g = Math.floor(g / count);
        b = Math.floor(b / count);

        // 填充整个块
        for (let py = by; py < by + blockSize && py < absHeight; py++) {
          for (let px = bx; px < bx + blockSize && px < absWidth; px++) {
            const i = (py * absWidth + px) * 4;
            data[i] = r;
            data[i + 1] = g;
            data[i + 2] = b;
          }
        }
      }
    }

    // 将处理后的图像数据放回画布
    ctx.putImageData(imageData, Math.min(x, x + width), Math.min(y, y + height));
  };

  // 绘制箭头
  const drawArrow = (
    ctx: CanvasRenderingContext2D,
    fromX: number,
    fromY: number,
    toX: number,
    toY: number,
    lineWidth: number = 3
  ) => {
    // 箭头头部大小随线条粗细变化，基础大小 15，每增加 1px 粗细增加 2
    const headLength = 15 + (lineWidth - 3) * 2;
    // 箭头头部角度（弧度）
    const headAngle = Math.PI / 6; // 30度

    const angle = Math.atan2(toY - fromY, toX - fromX);

    // 计算箭头头部开始的点（线条终点）
    const lineEndX = toX - headLength * Math.cos(angle);
    const lineEndY = toY - headLength * Math.sin(angle);

    // 绘制线条（画到箭头头部开始的点）
    ctx.beginPath();
    ctx.moveTo(fromX, fromY);
    ctx.lineTo(lineEndX, lineEndY);
    ctx.stroke();

    // 绘制箭头头部
    ctx.beginPath();
    ctx.moveTo(toX, toY);
    ctx.lineTo(toX - headLength * Math.cos(angle - headAngle), toY - headLength * Math.sin(angle - headAngle));
    ctx.lineTo(toX - headLength * Math.cos(angle + headAngle), toY - headLength * Math.sin(angle + headAngle));
    ctx.closePath();
    ctx.fill();
  };

  // 处理鼠标按下
  const handleMouseDown = (e: React.MouseEvent<HTMLCanvasElement>) => {
    if (mode !== 'editing') return;

    const canvas = canvasRef.current;
    if (!canvas) return;

    const rect = canvas.getBoundingClientRect();
    // 计算画布的缩放比例（实际像素尺寸 / 显示尺寸）
    const scaleX = canvas.width / rect.width;
    const scaleY = canvas.height / rect.height;
    // 转换为画布实际像素坐标
    const x = (e.clientX - rect.left) * scaleX;
    const y = (e.clientY - rect.top) * scaleY;

    setIsDrawing(true);
    setStartPos({ x, y });

    if (selectedTool === 'brush') {
      setCurrentAnnotation({
        id: Date.now().toString(),
        type: selectedTool,
        x, y,
        color: selectedColor,
        lineWidth,
        points: [{ x, y }],
      });
    } else if (selectedTool === 'text') {
      // 文字工具：显示内联输入框
      const input = document.createElement('input');
      input.type = 'text';
      input.placeholder = '输入文字...';
      input.style.position = 'fixed';
      // 直接使用鼠标屏幕坐标定位输入框
      input.style.left = `${e.clientX}px`;
      input.style.top = `${e.clientY}px`;
      input.style.background = 'rgba(0, 0, 0, 0.8)';
      input.style.color = '#fff';
      input.style.border = `2px solid ${selectedColor}`;
      input.style.borderRadius = '4px';
      input.style.padding = '8px';
      input.style.fontSize = `${lineWidth * 5}px`;
      input.style.fontFamily = 'Arial, sans-serif';
      input.style.zIndex = '10001';
      input.style.minWidth = '200px';

      document.body.appendChild(input);
      input.focus();

      const handleTextConfirm = () => {
        const text = input.value.trim();
        if (text) {
          const newAnnotation: Annotation = {
            id: Date.now().toString(),
            type: selectedTool,
            x, y,
            color: selectedColor,
            lineWidth,
            text,
          };
          setAnnotations(prev => [...prev, newAnnotation]);
        }
        document.body.removeChild(input);
        setIsDrawing(false);
      };

      const handleTextCancel = () => {
        document.body.removeChild(input);
        setIsDrawing(false);
      };

      // 确认输入：Enter 或失去焦点
      input.addEventListener('keydown', (e) => {
        if (e.key === 'Enter') {
          handleTextConfirm();
        } else if (e.key === 'Escape') {
          handleTextCancel();
        }
      });

      // 延迟添加失去焦点监听，避免立即触发
      setTimeout(() => {
        input.addEventListener('blur', () => {
          if (document.body.contains(input)) {
            handleTextConfirm();
          }
        });
      }, 100);
    } else if (selectedTool === 'blur' || selectedTool === 'mosaic' || selectedTool === 'crop') {
      // 这些工具需要区域选择，初始化当前标注
      setCurrentAnnotation({
        id: Date.now().toString(),
        type: selectedTool,
        x, y,
        width: 0,
        height: 0,
        color: selectedColor,
        lineWidth,
      });
    } else {
      // 矩形、椭圆、箭头工具
      setCurrentAnnotation({
        id: Date.now().toString(),
        type: selectedTool,
        x, y,
        width: 0,
        height: 0,
        color: selectedColor,
        lineWidth,
      });
    }
  };

  // 处理鼠标移动
  const handleMouseMove = (e: React.MouseEvent<HTMLCanvasElement>) => {
    if (!isDrawing || mode !== 'editing') return;

    const canvas = canvasRef.current;
    if (!canvas) return;

    const rect = canvas.getBoundingClientRect();
    // 计算画布的缩放比例（实际像素尺寸 / 显示尺寸）
    const scaleX = canvas.width / rect.width;
    const scaleY = canvas.height / rect.height;
    // 转换为画布实际像素坐标
    const x = (e.clientX - rect.left) * scaleX;
    const y = (e.clientY - rect.top) * scaleY;

    if (selectedTool === 'brush' && currentAnnotation) {
      setCurrentAnnotation({
        ...currentAnnotation,
        points: [...(currentAnnotation.points || []), { x, y }],
      });
    } else {
      const newAnnotation: Annotation = {
        id: Date.now().toString(),
        type: selectedTool,
        x: startPos.x,
        y: startPos.y,
        width: x - startPos.x,
        height: y - startPos.y,
        color: selectedColor,
        lineWidth,
      };
      // 为模糊和马赛克工具添加预览标记
      if (selectedTool === 'blur' || selectedTool === 'mosaic') {
        (newAnnotation as any).isPreview = true;
      }
      setCurrentAnnotation(newAnnotation);
    }
  };

  // 处理鼠标释放
  const handleMouseUp = () => {
    if (isDrawing && currentAnnotation) {
      // 裁剪工具特殊处理
      if (currentAnnotation.type === 'crop' &&
          currentAnnotation.width &&
          currentAnnotation.height &&
          Math.abs(currentAnnotation.width) > 20 &&
          Math.abs(currentAnnotation.height) > 20) {

        const canvas = canvasRef.current;
        const image = imageRef.current;
        if (canvas && image) {
          // 创建临时画布来存储裁剪后的图像
          const croppedCanvas = document.createElement('canvas');
          const x = Math.min(currentAnnotation.x, currentAnnotation.x + currentAnnotation.width);
          const y = Math.min(currentAnnotation.y, currentAnnotation.y + currentAnnotation.height);
          const w = Math.abs(currentAnnotation.width);
          const h = Math.abs(currentAnnotation.height);

          croppedCanvas.width = w;
          croppedCanvas.height = h;

          const croppedCtx = croppedCanvas.getContext('2d');
          if (croppedCtx) {
            // 从原始图像裁剪区域
            croppedCtx.drawImage(
              image,
              x, y, w, h,
              0, 0, w, h
            );

            // 更新当前图像为裁剪后的版本
            const croppedDataUrl = croppedCanvas.toDataURL('image/png');
            setCurrentImage(croppedDataUrl);

            // 重置标注（裁剪后旧标注不再有效）
            setAnnotations([]);

            addLog(`已裁剪图像: ${w}x${h}`);
          }
        }
      } else if (currentAnnotation.type === 'blur' || currentAnnotation.type === 'mosaic') {
        // 模糊和马赛克效果是破坏性的，直接应用到画布
        const canvas = canvasRef.current;
        if (canvas) {
          const ctx = canvas.getContext('2d');
          if (ctx) {
            // 重绘所有旧标注
            renderCanvas();
            // 应用当前效果
            drawAnnotation(ctx, currentAnnotation);
            // 保存到标注列表中
            setAnnotations([...annotations, currentAnnotation]);
          }
        }
      } else {
        // 其他工具直接添加到标注列表
        setAnnotations([...annotations, currentAnnotation]);
      }
    }
    setIsDrawing(false);
    setCurrentAnnotation(null);
  };

  // 撤销
  const handleUndo = useCallback(() => {
    if (annotations.length > 0) {
      setAnnotations(annotations.slice(0, -1));
    }
  }, [annotations]);

  // 关闭编辑器 - 移到这里，避免依赖顺序问题
  const handleClose = useCallback(() => {
    if (onClose) {
      onClose();
    } else {
      window.close();
    }
  }, [onClose]);

  // 保存图片 - 使用系统保存对话框
  const handleSave = useCallback(async () => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const dataUrl = canvas.toDataURL('image/png');
    try {
      // 使用带对话框的保存方法，让用户选择保存路径
      const savedPath = await ScreenshotService.SaveImageWithDataWithDialog(dataUrl);
      if (savedPath) {
        success(`图片已保存到: ${savedPath}`);
        handleClose();
      }
      // 如果 savedPath 为空，说明用户取消了保存，不关闭编辑器
    } catch (err) {
      console.error('保存失败:', err);
      const errorMessage = err instanceof Error ? err.message : String(err);
      // 区分用户取消和真正的错误
      if (errorMessage.includes('cancelled') || errorMessage.includes('用户取消')) {
        addLog('用户取消了保存');
      } else {
        error('保存失败: ' + errorMessage);
      }
    }
  }, [success, error, handleClose, addLog]);

  // 更新 ref 引用
  useEffect(() => {
    handleSaveRef.current = handleSave;
  }, [handleSave]);

  // 更新 ref 引用
  useEffect(() => {
    handleUndoRef.current = handleUndo;
  }, [handleUndo]);

  // 复制并关闭
  const handleCopyAndClose = useCallback(async () => {
    addLog('=== 开始复制到剪贴板 ===');

    const canvas = canvasRef.current;
    if (!canvas) {
      addLog('✗ 错误: Canvas 引用不存在');
      return;
    }

    addLog('正在生成 PNG 数据...');
    const dataUrl = canvas.toDataURL('image/png');
    addLog(`PNG 数据长度: ${dataUrl.length} 字符`);
    addLog(`PNG 数据前缀: ${dataUrl.substring(0, 50)}...`);

    try {
      addLog('调用 CopyImageDataToClipboard...');
      await ScreenshotService.CopyImageDataToClipboard(dataUrl);
      addLog('✓ CopyImageDataToClipboard 调用成功');
      success('图片已复制到剪贴板');
      addLog('✓ 已显示成功提示');
      handleClose();
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : String(err);
      addLog(`✗ 复制失败: ${errorMessage}`);
      error('复制失败: ' + errorMessage);
    }
    addLog('=== 复制操作结束 ===');
  }, [success, error, addLog, handleClose]);

  // 更新 ref 引用
  useEffect(() => {
    handleCopyAndCloseRef.current = handleCopyAndClose;
  }, [handleCopyAndClose]);

  // 取消
  const handleCancel = useCallback(async () => {
    console.log('[ScreenshotEditor] handleCancel 被调用');
    try {
      console.log('[ScreenshotEditor] 调用 ScreenshotService.CancelCapture...');
      await ScreenshotService.CancelCapture();
      console.log('[ScreenshotEditor] CancelCapture 调用成功，调用 handleClose');
      handleClose();
    } catch (error) {
      console.error('[ScreenshotEditor] 取消失败:', error);
    }
  }, [handleClose]);

  // 更新 ref 引用
  useEffect(() => {
    handleCancelRef.current = handleCancel;
  }, [handleCancel]);

  // 监听键盘快捷键
  useEffect(() => {
    const handleKeyPress = (e: KeyboardEvent) => {
      // 记录所有键盘事件到调试面板
      if (e.key === 'Enter') {
        addLog(`键盘事件: Enter 键 (mode: ${mode})`);
      }

      if (mode === 'selecting') {
        // 选择模式下的快捷键由 ScreenshotSelector 处理
        return;
      }

      // 编辑模式下的快捷键 - 使用 ref 避免依赖变化导致监听器重建
      if (e.key === 'Escape') {
        e.preventDefault();
        e.stopPropagation();
        handleCancelRef.current?.();
      } else if (e.key === 'Enter') {
        e.preventDefault();
        e.stopPropagation();
        addLog('Enter 键被触发，执行复制操作');
        // 使用 ref 调用，避免依赖问题
        handleCopyAndCloseRef.current?.();
      } else if ((e.metaKey || e.ctrlKey) && e.key === 's') {
        e.preventDefault();
        e.stopPropagation();
        handleSaveRef.current?.();
      } else if ((e.metaKey || e.ctrlKey) && e.key === 'z') {
        e.preventDefault();
        e.stopPropagation();
        handleUndoRef.current?.();
      } else if ((e.metaKey || e.ctrlKey) && e.shiftKey && e.key === 'z') {
        e.preventDefault();
        e.stopPropagation();
        // handleRedo 暂不处理
      } else if (e.key >= '1' && e.key <= '9') {
        const toolIndex = parseInt(e.key) - 1;
        if (toolIndex < TOOLS.length) {
          setSelectedTool(TOOLS[toolIndex].type);
        }
      }
    };

    window.addEventListener('keydown', handleKeyPress, { capture: true });

    return () => {
      window.removeEventListener('keydown', handleKeyPress);
    };
  }, [mode, selectedTool, addLog]); // 只依赖真正需要的状态，不依赖函数

  // 当图片加载完成后渲染
  useEffect(() => {
    if (currentImage && imageRef.current) {
      imageRef.current.onload = () => {
        renderCanvas();
      };
    }
  }, [currentImage, renderCanvas]);

  // 当标注变化时重新渲染
  useEffect(() => {
    renderCanvas();
  }, [annotations, currentAnnotation, renderCanvas]);

  // 如果在选择模式，显示区域选择器
  if (mode === 'selecting') {
    return (
      <ScreenshotSelector
        imageData={originalImage}
        onSelectionComplete={handleSelectionComplete}
        onCancel={handleCancel}
      />
    );
  }

  // 编辑模式
  return (
    <div className="screenshot-editor">
      {/* 工具栏 */}
      <div className="toolbar">
        <div className="toolbar-section">
          {TOOLS.map(tool => (
            <button
              key={tool.type}
              className={`tool-button ${selectedTool === tool.type ? 'active' : ''}`}
              onClick={() => setSelectedTool(tool.type)}
              title={`${tool.label} (${TOOLS.indexOf(tool) + 1})`}
            >
              {tool.icon}
            </button>
          ))}
        </div>

        <div className="toolbar-section">
          {COLORS.map(color => (
            <button
              key={color.name}
              className={`color-button ${selectedColor === color.value ? 'active' : ''}`}
              style={{ backgroundColor: color.value }}
              onClick={() => setSelectedColor(color.value)}
              title={color.name}
            />
          ))}
        </div>

        <div className="toolbar-section">
          <input
            type="range"
            min="1"
            max="10"
            value={lineWidth}
            onChange={(e) => setLineWidth(parseInt(e.target.value))}
            className="line-width-slider"
          />
          <span className="line-width-label">{lineWidth}px</span>
        </div>

        <div className="toolbar-section actions">
          <button onClick={handleUndo} title="撤销 (Ctrl+Z)" disabled={annotations.length === 0}>
            ↶
          </button>
          <button onClick={handleSave} title="保存 (Ctrl+S)">
            💾
          </button>
          <button onClick={handleCopyAndClose} title="复制并关闭 (Enter)">
            📋
          </button>
          <button onClick={handleCancel} title="取消 (Esc)" className="cancel-button">
            ✕
          </button>
        </div>
      </div>

      {/* 画布 */}
      <div className="canvas-container">
        <img ref={imageRef} src={currentImage} alt="Screenshot" style={{ display: 'none' }} />
        <canvas
          ref={canvasRef}
          onMouseDown={handleMouseDown}
          onMouseMove={handleMouseMove}
          onMouseUp={handleMouseUp}
          onMouseLeave={handleMouseUp}
        />
      </div>

      {/* 调试日志面板 */}
      {debugLogs.length > 0 && (
        <div style={{
          position: 'absolute',
          bottom: '10px',
          right: '10px',
          background: 'rgba(0, 0, 0, 0.8)',
          color: '#00ff00',
          padding: '10px',
          borderRadius: '5px',
          fontFamily: 'monospace',
          fontSize: '11px',
          maxWidth: '400px',
          maxHeight: '200px',
          overflow: 'auto',
          zIndex: 1000,
          border: '1px solid #333'
        }}>
          <div style={{ fontWeight: 'bold', marginBottom: '5px', color: '#fff' }}>调试日志:</div>
          {debugLogs.map((log, index) => (
            <div key={index} style={{ marginBottom: '2px' }}>{log}</div>
          ))}
        </div>
      )}
    </div>
  );
};

export default ScreenshotEditor;
