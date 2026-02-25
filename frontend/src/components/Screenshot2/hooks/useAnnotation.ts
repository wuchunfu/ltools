import { useState, useCallback, useRef } from 'react';

// 标注类型
export type AnnotationType = 'rect' | 'ellipse' | 'arrow' | 'brush' | 'text' | 'mosaic' | 'blur' | null;

// 标注数据
export interface Annotation {
  id: string;
  type: AnnotationType;
  x: number;
  y: number;
  width?: number;
  height?: number;
  points?: { x: number; y: number }[]; // for brush
  text?: string; // for text
  fontSize?: number; // 文字大小
  color?: string;
  strokeWidth?: number;
}

// 标注状态
export interface AnnotationState {
  annotations: Annotation[];
  currentType: AnnotationType;
  currentColor: string;
  strokeWidth: number;
  fontSize: number; // 文字大小
  isDrawing: boolean;
  currentAnnotation: Annotation | null;
}

/**
 * 标注管理 Hook
 * 提供标注工具的状态管理和操作
 */
export function useAnnotation() {
  const [state, setState] = useState<AnnotationState>({
    annotations: [],
    currentType: null,
    currentColor: '#ff0000',
    strokeWidth: 2,
    fontSize: 18, // 默认文字大小
    isDrawing: false,
    currentAnnotation: null,
  });

  const historyRef = useRef<Annotation[][]>([[]]);
  const historyIndexRef = useRef(0);

  // 设置当前工具
  const setTool = useCallback((type: AnnotationType) => {
    setState(prev => ({ ...prev, currentType: type }));
  }, []);

  // 设置颜色
  const setColor = useCallback((color: string) => {
    setState(prev => ({ ...prev, currentColor: color }));
  }, []);

  // 设置线宽
  const setStrokeWidth = useCallback((width: number) => {
    setState(prev => ({ ...prev, strokeWidth: width }));
  }, []);

  // 设置文字大小
  const setFontSize = useCallback((size: number) => {
    setState(prev => ({ ...prev, fontSize: size }));
  }, []);

  // 开始绘制
  const startDrawing = useCallback((x: number, y: number) => {
    if (!state.currentType) return;

    const newAnnotation: Annotation = {
      id: `ann-${Date.now()}`,
      type: state.currentType,
      x,
      y,
      color: state.currentColor,
      strokeWidth: state.strokeWidth,
    };

    if (state.currentType === 'brush') {
      newAnnotation.points = [{ x, y }];
    }

    setState(prev => ({
      ...prev,
      isDrawing: true,
      currentAnnotation: newAnnotation,
    }));
  }, [state.currentType, state.currentColor, state.strokeWidth]);

  // 更新绘制
  const updateDrawing = useCallback((x: number, y: number) => {
    if (!state.isDrawing || !state.currentAnnotation) return;

    setState(prev => {
      if (!prev.currentAnnotation) return prev;

      const ann = prev.currentAnnotation;

      if (ann.type === 'brush' && ann.points) {
        return {
          ...prev,
          currentAnnotation: {
            ...ann,
            points: [...ann.points, { x, y }],
          },
        };
      }

      // 对于箭头，保持方向性（允许负的 width/height）
      if (ann.type === 'arrow') {
        return {
          ...prev,
          currentAnnotation: {
            ...ann,
            width: x - ann.x,
            height: y - ann.y,
          },
        };
      }

      // 对于矩形、椭圆等，使用绝对值
      return {
        ...prev,
        currentAnnotation: {
          ...ann,
          width: Math.abs(x - ann.x),
          height: Math.abs(y - ann.y),
          x: Math.min(ann.x, x),
          y: Math.min(ann.y, y),
        },
      };
    });
  }, [state.isDrawing, state.currentAnnotation]);

  // 结束绘制
  const finishDrawing = useCallback(() => {
    if (!state.currentAnnotation) return;

    // 保存到历史
    historyRef.current = historyRef.current.slice(0, historyIndexRef.current + 1);
    historyRef.current.push([...state.annotations, state.currentAnnotation]);
    historyIndexRef.current = historyRef.current.length - 1;

    setState(prev => ({
      ...prev,
      isDrawing: false,
      annotations: [...prev.annotations, prev.currentAnnotation!],
      currentAnnotation: null,
    }));
  }, [state.currentAnnotation, state.annotations]);

  // 撤销
  const undo = useCallback(() => {
    if (historyIndexRef.current > 0) {
      historyIndexRef.current--;
      setState(prev => ({
        ...prev,
        annotations: historyRef.current[historyIndexRef.current] || [],
      }));
    }
  }, []);

  // 重做
  const redo = useCallback(() => {
    if (historyIndexRef.current < historyRef.current.length - 1) {
      historyIndexRef.current++;
      setState(prev => ({
        ...prev,
        annotations: historyRef.current[historyIndexRef.current],
      }));
    }
  }, []);

  // 清除所有标注
  const clearAnnotations = useCallback(() => {
    historyRef.current = [[]];
    historyIndexRef.current = 0;
    setState(prev => ({ ...prev, annotations: [] }));
  }, []);

  // 删除指定标注
  const deleteAnnotation = useCallback((id: string) => {
    setState(prev => {
      const newAnnotations = prev.annotations.filter(a => a.id !== id);
      historyRef.current = historyRef.current.slice(0, historyIndexRef.current + 1);
      historyRef.current.push(newAnnotations);
      historyIndexRef.current = historyRef.current.length - 1;
      return { ...prev, annotations: newAnnotations };
    });
  }, []);

  // 添加文字标注
  const addTextAnnotation = useCallback((x: number, y: number, text: string) => {
    const newAnnotation: Annotation = {
      id: `ann-${Date.now()}`,
      type: 'text',
      x,
      y,
      text,
      color: state.currentColor,
      fontSize: state.fontSize,
      strokeWidth: state.strokeWidth,
    };

    // 保存到历史
    historyRef.current = historyRef.current.slice(0, historyIndexRef.current + 1);
    historyRef.current.push([...state.annotations, newAnnotation]);
    historyIndexRef.current = historyRef.current.length - 1;

    setState(prev => ({
      ...prev,
      annotations: [...prev.annotations, newAnnotation],
    }));
  }, [state.currentColor, state.fontSize, state.strokeWidth, state.annotations]);

  return {
    annotations: state.annotations,
    currentType: state.currentType,
    currentColor: state.currentColor,
    strokeWidth: state.strokeWidth,
    fontSize: state.fontSize,
    isDrawing: state.isDrawing,
    currentAnnotation: state.currentAnnotation,
    setTool,
    setColor,
    setStrokeWidth,
    setFontSize,
    startDrawing,
    updateDrawing,
    finishDrawing,
    undo,
    redo,
    clearAnnotations,
    deleteAnnotation,
    addTextAnnotation,
    canUndo: historyIndexRef.current > 0,
    canRedo: historyIndexRef.current < historyRef.current.length - 1,
  };
}
