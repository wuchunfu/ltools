import { useState, useCallback, useRef } from 'react';

// 选区类型
export interface Selection {
  x: number;
  y: number;
  width: number;
  height: number;
}

// 交互模式
export type InteractionMode = 'idle' | 'creating' | 'moving' | 'resizing';

// 调整手柄位置
export type HandlePosition = 'nw' | 'n' | 'ne' | 'e' | 'se' | 's' | 'sw' | 'w' | null;

// 选区状态
export interface SelectionState {
  selection: Selection | null;
  mode: InteractionMode;
  activeHandle: HandlePosition;
}

// 最小选区尺寸
const MIN_SELECTION_SIZE = 10;

// 手柄检测半径
const HANDLE_DETECTION_RADIUS = 8;

/**
 * 选区管理 Hook
 * 提供选区的创建、移动、调整功能
 */
export function useSelection() {
  const [state, setState] = useState<SelectionState>({
    selection: null,
    mode: 'idle',
    activeHandle: null,
  });

  const startPosRef = useRef({ x: 0, y: 0 });
  const originalSelectionRef = useRef<Selection | null>(null);

  // 开始创建选区
  const startCreating = useCallback((x: number, y: number) => {
    startPosRef.current = { x, y };
    setState({
      selection: { x, y, width: 0, height: 0 },
      mode: 'creating',
      activeHandle: null,
    });
  }, []);

  // 更新创建中的选区
  const updateCreating = useCallback((x: number, y: number) => {
    setState(prev => {
      if (prev.mode !== 'creating' || !prev.selection) return prev;

      const startX = startPosRef.current.x;
      const startY = startPosRef.current.y;

      return {
        ...prev,
        selection: {
          x: Math.min(startX, x),
          y: Math.min(startY, y),
          width: Math.abs(x - startX),
          height: Math.abs(y - startY),
        },
      };
    });
  }, []);

  // 检测点击位置是否在手柄上
  const getHandleAtPosition = useCallback((x: number, y: number, selection: Selection): HandlePosition => {
    const handles: { pos: HandlePosition; x: number; y: number }[] = [
      { pos: 'nw', x: selection.x, y: selection.y },
      { pos: 'n', x: selection.x + selection.width / 2, y: selection.y },
      { pos: 'ne', x: selection.x + selection.width, y: selection.y },
      { pos: 'e', x: selection.x + selection.width, y: selection.y + selection.height / 2 },
      { pos: 'se', x: selection.x + selection.width, y: selection.y + selection.height },
      { pos: 's', x: selection.x + selection.width / 2, y: selection.y + selection.height },
      { pos: 'sw', x: selection.x, y: selection.y + selection.height },
      { pos: 'w', x: selection.x, y: selection.y + selection.height / 2 },
    ];

    for (const handle of handles) {
      const dx = x - handle.x;
      const dy = y - handle.y;
      if (dx * dx + dy * dy <= HANDLE_DETECTION_RADIUS * HANDLE_DETECTION_RADIUS) {
        return handle.pos;
      }
    }

    return null;
  }, []);

  // 检测点击位置是否在选区内
  const isInsideSelection = useCallback((x: number, y: number, selection: Selection): boolean => {
    return (
      x >= selection.x &&
      x <= selection.x + selection.width &&
      y >= selection.y &&
      y <= selection.y + selection.height
    );
  }, []);

  // 开始移动选区
  const startMoving = useCallback((x: number, y: number) => {
    setState(prev => {
      if (!prev.selection) return prev;
      originalSelectionRef.current = { ...prev.selection };
      startPosRef.current = { x, y };
      return { ...prev, mode: 'moving' };
    });
  }, []);

  // 更新移动
  const updateMoving = useCallback((x: number, y: number) => {
    setState(prev => {
      if (prev.mode !== 'moving' || !prev.selection || !originalSelectionRef.current) return prev;

      const dx = x - startPosRef.current.x;
      const dy = y - startPosRef.current.y;

      return {
        ...prev,
        selection: {
          ...prev.selection,
          x: originalSelectionRef.current.x + dx,
          y: originalSelectionRef.current.y + dy,
        },
      };
    });
  }, []);

  // 开始调整大小
  const startResizing = useCallback((x: number, y: number, handle: HandlePosition) => {
    setState(prev => {
      if (!prev.selection) return prev;
      originalSelectionRef.current = { ...prev.selection };
      startPosRef.current = { x, y };
      return { ...prev, mode: 'resizing', activeHandle: handle };
    });
  }, []);

  // 更新调整大小
  const updateResizing = useCallback((x: number, y: number) => {
    setState(prev => {
      if (prev.mode !== 'resizing' || !prev.selection || !originalSelectionRef.current || !prev.activeHandle) {
        return prev;
      }

      const dx = x - startPosRef.current.x;
      const dy = y - startPosRef.current.y;
      const orig = originalSelectionRef.current;
      const handle = prev.activeHandle;

      let newX = orig.x;
      let newY = orig.y;
      let newWidth = orig.width;
      let newHeight = orig.height;

      // 根据手柄位置调整
      if (handle.includes('w')) {
        newX = orig.x + dx;
        newWidth = orig.width - dx;
      }
      if (handle.includes('e')) {
        newWidth = orig.width + dx;
      }
      if (handle.includes('n')) {
        newY = orig.y + dy;
        newHeight = orig.height - dy;
      }
      if (handle.includes('s')) {
        newHeight = orig.height + dy;
      }

      // 确保最小尺寸
      if (newWidth < MIN_SELECTION_SIZE) {
        if (handle.includes('w')) {
          newX = orig.x + orig.width - MIN_SELECTION_SIZE;
        }
        newWidth = MIN_SELECTION_SIZE;
      }
      if (newHeight < MIN_SELECTION_SIZE) {
        if (handle.includes('n')) {
          newY = orig.y + orig.height - MIN_SELECTION_SIZE;
        }
        newHeight = MIN_SELECTION_SIZE;
      }

      return {
        ...prev,
        selection: {
          x: newX,
          y: newY,
          width: newWidth,
          height: newHeight,
        },
      };
    });
  }, []);

  // 完成交互
  const finishInteraction = useCallback(() => {
    setState(prev => ({
      ...prev,
      mode: 'idle',
      activeHandle: null,
    }));
    originalSelectionRef.current = null;
  }, []);

  // 清除选区
  const clearSelection = useCallback(() => {
    setState({
      selection: null,
      mode: 'idle',
      activeHandle: null,
    });
  }, []);

  // 设置选区
  const setSelection = useCallback((selection: Selection | null) => {
    setState(prev => ({
      ...prev,
      selection,
    }));
  }, []);

  return {
    selection: state.selection,
    mode: state.mode,
    activeHandle: state.activeHandle,
    startCreating,
    updateCreating,
    startMoving,
    updateMoving,
    startResizing,
    updateResizing,
    finishInteraction,
    clearSelection,
    setSelection,
    getHandleAtPosition,
    isInsideSelection,
  };
}
