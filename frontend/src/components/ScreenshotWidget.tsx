import { useState, useCallback, useEffect } from 'react';
import { Icon } from './Icon';
import { Events } from '@wailsio/runtime';
import * as ScreenshotService from '../../bindings/ltools/plugins/screenshot/screenshotservice';

interface ScreenshotWidgetProps {
  className?: string;
}

export function ScreenshotWidget({ className = '' }: ScreenshotWidgetProps) {
  const [isCapturing, setIsCapturing] = useState(false);
  const [lastCapturePath, setLastCapturePath] = useState<string>('');
  const [lastCaptureImage, setLastCaptureImage] = useState<string>('');
  const [error, setError] = useState<string>('');

  // 监听截图事件
  useEffect(() => {
    const unsubscribeSaved = Events.On('screenshot:saved', (ev: { data: string }) => {
      setLastCapturePath(ev.data);
      setIsCapturing(false);
      setError('');
    });

    const unsubscribeCancelled = Events.On('screenshot:cancelled', () => {
      setIsCapturing(false);
      setLastCaptureImage(''); // 清除预览图片
    });

    const unsubscribeCopied = Events.On('screenshot:copied', () => {
      setIsCapturing(false);
    });

    const unsubscribeError = Events.On('screenshot:error', (ev: { data: string }) => {
      setError(ev.data);
      setIsCapturing(false);
    });

    // 监听截图捕获事件，获取图片预览
    const unsubscribeCaptured = Events.On('screenshot:captured', async (ev: { data: string }) => {
      try {
        // 获取捕获的图片数据
        const imageData = await ScreenshotService.GetCapturedImage();
        setLastCaptureImage(imageData);
        setIsCapturing(false);
      } catch (error) {
        console.error('获取截图失败:', error);
        setError(String(error));
        setIsCapturing(false);
      }
    });

    return () => {
      unsubscribeSaved();
      unsubscribeCancelled();
      unsubscribeCopied();
      unsubscribeError();
      unsubscribeCaptured();
    };
  }, []);

  // 触发截图
  const handleStartCapture = useCallback(async () => {
    console.log('[ScreenshotWidget] 开始截图...');
    setError('');
    setIsCapturing(true);
    setLastCaptureImage(''); // 清除之前的预览

    try {
      console.log('[ScreenshotWidget] 调用 Trigger 方法...');
      await ScreenshotService.Trigger();
      console.log('[ScreenshotWidget] Trigger 调用成功');
    } catch (err) {
      console.error('[ScreenshotWidget] Trigger 调用失败:', err);
      setError(String(err));
      setIsCapturing(false);
    }
  }, []);

  // 获取快捷键信息
  const getShortcutKey = () => {
    const isMac = navigator.platform.toUpperCase().indexOf('MAC') >= 0;
    return isMac ? 'Cmd+Shift+S' : 'Ctrl+Shift+S';
  };

  return (
    <div className={`screenshot-widget ${className}`}>
      {/* 标题 */}
      <div className="widget-header">
        <h3 className="widget-title">
          <Icon name="camera" className="widget-title-icon" />
          截图工具
        </h3>
      </div>

      {/* 内容 */}
      <div className="widget-content">
        {/* 快捷键提示 */}
        <div className="shortcut-hint glass-light rounded-lg p-4 mb-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <Icon name="keyboard" className="text-[#7C3AED]" />
              <span className="text-white/70">全局快捷键</span>
            </div>
            <kbd className="shortcut-key">{getShortcutKey()}</kbd>
          </div>
        </div>

        {/* 截图按钮 */}
        <button
          className="screenshot-button w-full py-4 rounded-xl font-semibold text-lg transition-all duration-200 clickable"
          onClick={handleStartCapture}
          disabled={isCapturing}
        >
          {isCapturing ? (
            <>
              <Icon name="sparkles" className="animate-spin" />
              截图中...
            </>
          ) : (
            <>
              <Icon name="camera" />
              开始截图
            </>
          )}
        </button>

        {/* 状态信息 */}
        {error && (
          <div className="mt-4 p-3 bg-[#EF4444]/10 border border-[#EF4444]/20 rounded-lg flex items-center gap-2 text-[#EF4444]">
            <Icon name="alert-circle" />
            <span className="text-sm">{error}</span>
          </div>
        )}

        {lastCaptureImage && !error && (
          <div className="mt-4 rounded-lg overflow-hidden border border-[#10B981]/20 bg-[#10B981]/5">
            <div className="p-2 bg-[#10B981]/10 border-b border-[#10B981]/20">
              <span className="text-sm text-[#10B981]">📸 截图预览</span>
            </div>
            <img src={lastCaptureImage} alt="截图预览" className="w-full h-auto" />
          </div>
        )}

        {lastCapturePath && !error && (
          <div className="mt-4 p-3 bg-[#10B981]/10 border border-[#10B981]/20 rounded-lg flex items-center gap-2 text-[#10B981]">
            <Icon name="check-circle" />
            <span className="text-sm flex-1 truncate">已保存: {lastCapturePath}</span>
          </div>
        )}

        {/* 使用提示 */}
        <div className="mt-4 text-white/50 text-sm space-y-1">
          <p>💡 使用说明：</p>
          <ul className="list-disc list-inside ml-2 space-y-1">
            <li>点击上方按钮或使用全局快捷键触发截图</li>
            <li>拖拽选择截图范围</li>
            <li>双击或按 Enter 确认选择</li>
            <li>使用工具条添加标注</li>
            <li>按 Ctrl+S 保存或 Enter 复制到剪贴板</li>
          </ul>
        </div>
      </div>
    </div>
  );
}
