import { useToastContext } from '../contexts/ToastContext';

export function useToast() {
  const context = useToastContext();

  return {
    toasts: context.toasts,
    addToast: context.addToast,
    removeToast: context.removeToast,
    success: context.success,
    error: context.error,
    info: context.info,
    warning: context.warning,
  };
}
