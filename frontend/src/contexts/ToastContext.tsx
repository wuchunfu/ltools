import { createContext, useContext, useState, useCallback, ReactNode } from 'react';
import { Toast, ToastType } from '../components/Toast';

let toastIdCounter = 0;

interface ToastContextValue {
  toasts: Toast[];
  addToast: (type: ToastType, message: string, duration?: number) => string;
  removeToast: (id: string) => void;
  success: (message: string, duration?: number) => string;
  error: (message: string, duration?: number) => string;
  info: (message: string, duration?: number) => string;
  warning: (message: string, duration?: number) => string;
}

const ToastContext = createContext<ToastContextValue | undefined>(undefined);

export function useToastContext(): ToastContextValue {
  const context = useContext(ToastContext);
  if (!context) {
    throw new Error('useToastContext must be used within a ToastProvider');
  }
  return context;
}

export function ToastProvider({ children }: { children: ReactNode }) {
  const [toasts, setToasts] = useState<Toast[]>([]);

  const addToast = useCallback((type: ToastType, message: string, duration?: number): string => {
    const id = `toast-${++toastIdCounter}`;
    const newToast: Toast = { id, type, message, duration };

    setToasts((prev) => [...prev, newToast]);

    return id;
  }, []);

  const removeToast = useCallback((id: string) => {
    setToasts((prev) => prev.filter((t) => t.id !== id));
  }, []);

  const success = useCallback((message: string, duration?: number) => {
    return addToast('success', message, duration);
  }, [addToast]);

  const error = useCallback((message: string, duration?: number) => {
    return addToast('error', message, duration);
  }, [addToast]);

  const info = useCallback((message: string, duration?: number) => {
    return addToast('info', message, duration);
  }, [addToast]);

  const warning = useCallback((message: string, duration?: number) => {
    return addToast('warning', message, duration);
  }, [addToast]);

  return (
    <ToastContext.Provider value={{ toasts, addToast, removeToast, success, error, info, warning }}>
      {children}
      <ToastContainer toasts={toasts} onRemove={removeToast} />
    </ToastContext.Provider>
  );
}

// 创建一个辅助函数来获取 context 值（用于非组件场景）
export function createToastContext() {
  // 这个函数只是导出接口，实际使用需要通过 Provider
  return null;
}

// ToastContainer 组件（从 Toast.tsx 移植过来）
import { Icon } from '../components/Icon';

interface ToastItemProps {
  toast: Toast;
  onRemove: (id: string) => void;
}

function ToastItem({ toast, onRemove }: ToastItemProps): JSX.Element {
  const getStyles = () => {
    switch (toast.type) {
      case 'success':
        return 'bg-[#22C55E]/10 border-[#22C55E]/20 text-[#22C55E]';
      case 'error':
        return 'bg-[#EF4444]/10 border-[#EF4444]/20 text-[#EF4444]';
      case 'warning':
        return 'bg-[#F59E0B]/10 border-[#F59E0B]/20 text-[#F59E0B]';
      case 'info':
      default:
        return 'bg-[#3B82F6]/10 border-[#3B82F6]/20 text-[#60A5FA]';
    }
  };

  const getIcon = () => {
    switch (toast.type) {
      case 'success':
        return 'check-circle';
      case 'error':
        return 'x-circle';
      case 'warning':
        return 'exclamation-circle';
      case 'info':
      default:
        return 'information-circle';
    }
  };

  return (
    <div
      className={`glass-light rounded-lg p-4 flex items-center gap-3 border ${getStyles()} animate-slide-in-right min-w-[300px] max-w-md`}
      role="alert"
    >
      <Icon name={getIcon() as any} size={20} />
      <p className="flex-1 text-sm text-white/90">{toast.message}</p>
      <button
        className="p-1 rounded hover:bg-white/5 transition-colors clickable"
        onClick={() => onRemove(toast.id)}
        aria-label="关闭"
      >
        <Icon name="x-circle" size={16} />
      </button>
    </div>
  );
}

interface ToastContainerProps {
  toasts: Toast[];
  onRemove: (id: string) => void;
}

function ToastContainer({ toasts, onRemove }: ToastContainerProps): JSX.Element | null {
  if (toasts.length === 0) return null;

  return (
    <div className="fixed top-4 right-4 z-50 flex flex-col gap-3">
      {toasts.map((toast) => (
        <ToastItem key={toast.id} toast={toast} onRemove={onRemove} />
      ))}
    </div>
  );
}
