import React, { useEffect, useState, useCallback } from 'react';
import * as ScreenshotService from '../../bindings/ltools/plugins/screenshot/screenshotservice';
import { Events } from '@wailsio/runtime';
import ScreenshotSelector from './ScreenshotSelector';
import './ScreenshotEditor.css';

interface ScreenshotEditorProps {
  imageData?: string;
  onClose?: () => void;
}

type EditorMode = 'selecting' | 'editing';

const ScreenshotEditorWithDebug: React.FC<ScreenshotEditorProps> = ({ imageData: propImageData, onClose }) => {
  const [mode, setMode] = useState<EditorMode>('selecting');
  const [originalImage, setOriginalImage] = useState<string>('');
  const [debugLogs, setDebugLogs] = useState<string[]>([]);

  // 添加调试日志的函数
  const addDebugLog = (message: string) => {
    const timestamp = new Date().toLocaleTimeString();
    setDebugLogs(prev => [...prev, `[${timestamp}] ${message}`]);
    console.log(`[Debug] ${message}`);
  };

  // 组件挂载时初始化
  useEffect(() => {
    addDebugLog('组件挂载');
    addDebugLog(`当前 URL: ${window.location.href}`);

    // 确保窗口获得焦点
    window.focus();
    addDebugLog('已调用 window.focus()');

    // 设置窗口标题
    document.title = 'Screenshot Editor - 按 ESC 取消';

    // 测试键盘事件监听器
    const testKeyListener = (e: KeyboardEvent) => {
      addDebugLog(`键盘事件测试: ${e.key}`);
    };

    window.addEventListener('keydown', testKeyListener);
    addDebugLog('测试键盘监听器已添加');

    return () => {
      window.removeEventListener('keydown', testKeyListener);
      addDebugLog('测试键盘监听器已移除');
    };
  }, []);

  // 从事件系统接收图片数据
  useEffect(() => {
    addDebugLog('设置图片数据事件监听器');

    const unsubscribe = Events.On('screenshot:image-data', (event) => {
      const data = event.data as string;
      addDebugLog(`收到图片数据事件，长度: ${data?.length || 0}`);
      if (data) {
        addDebugLog('设置原始图片');
        setOriginalImage(data);
      } else {
        addDebugLog('警告：图片数据为空！');
      }
    });

    // 同时尝试从 URL 参数获取（用于兼容）
    const urlParams = new URLSearchParams(window.location.search);
    const urlData = urlParams.get('data');
    if (urlData) {
      addDebugLog(`从 URL 获取图片数据，长度: ${urlData.length}`);
      setOriginalImage(urlData);
    } else if (propImageData) {
      addDebugLog('使用传入的图片数据');
      setOriginalImage(propImageData);
    }

    return () => {
      unsubscribe();
      addDebugLog('图片数据事件监听器已移除');
    };
  }, [propImageData]);

  // 处理区域选择完成
  const handleSelectionComplete = (selectedImageData: string) => {
    addDebugLog('区域选择完成');
    setMode('editing');
  };

  // 创建一个调试日志回调函数，传递给子组件
  const debugLogCallback = useCallback((message: string) => {
    addDebugLog(message);
  }, []);

  // 处理取消
  const handleCancel = async () => {
    addDebugLog('用户请求取消');
    try {
      await ScreenshotService.CancelCapture();
      addDebugLog('取消成功');
      if (onClose) {
        onClose();
      } else {
        window.close();
      }
    } catch (error) {
      addDebugLog(`取消失败: ${error}`);
    }
  };

  // 如果在选择模式，显示区域选择器
  if (mode === 'selecting') {
    return (
      <div style={{ position: 'relative', width: '100%', height: '100%' }}>
        {/* 调试面板 */}
        <div style={{
          position: 'absolute',
          top: '10px',
          left: '10px',
          right: '10px',
          background: 'rgba(0, 0, 0, 0.85)',
          color: '#4CD964',
          padding: '15px',
          borderRadius: '8px',
          fontFamily: 'monospace',
          fontSize: '11px',
          zIndex: 9999,  // 降低层级，确保不覆盖Canvas
          maxHeight: '400px',  // 增加最大高度
          overflowY: 'auto',  // 确保可以滚动
          overflowX: 'hidden',  // 隐藏横向滚动条
          border: '2px solid #4CD964',
          pointerEvents: 'none',  // 确保调试面板不拦截鼠标事件
          // WebkitOverflowScrolling: 'touch'
        }}>
          <div style={{
            fontSize: '14px',
            fontWeight: 'bold',
            marginBottom: '10px',
            borderBottom: '1px solid #4CD964',
            paddingBottom: '5px'
          }}>
            🔍 调试信息
          </div>
          {debugLogs.length === 0 ? (
            <div>等待日志...</div>
          ) : (
            debugLogs.map((log, index) => (
              <div key={index} style={{ marginBottom: '2px' }}>
                {log}
              </div>
            ))
          )}
          <div style={{
            marginTop: '10px',
            paddingTop: '10px',
            borderTop: '1px solid #4CD964',
            fontSize: '11px',
            opacity: 0.8
          }}>
            💡 按 ESC 键测试取消功能
          </div>
        </div>

        <ScreenshotSelector
          imageData={originalImage}
          onSelectionComplete={handleSelectionComplete}
          onCancel={handleCancel}
          onDebugLog={debugLogCallback}
        />
      </div>
    );
  }

  // 编辑模式
  return (
    <div className="screenshot-editor">
      <div style={{
        position: 'fixed',
        top: '10px',
        right: '10px',
        background: 'rgba(0, 0, 0, 0.8)',
        color: '#4CD964',
        padding: '15px',
        borderRadius: '8px',
        fontFamily: 'monospace',
        fontSize: '12px',
        zIndex: 10000,
        maxWidth: '400px',
        maxHeight: '300px',
        overflow: 'auto',
        border: '2px solid #4CD964'
      }}>
        <div style={{
          fontSize: '14px',
          fontWeight: 'bold',
          marginBottom: '10px',
          borderBottom: '1px solid #4CD964',
          paddingBottom: '5px'
        }}>
          🔍 调试信息 (编辑模式)
        </div>
        {debugLogs.map((log, index) => (
          <div key={index} style={{ marginBottom: '2px' }}>
            {log}
          </div>
        ))}
      </div>

      <div style={{
        position: 'fixed',
        top: '50%',
        left: '50%',
        transform: 'translate(-50%, -50%)',
        background: 'rgba(0, 0, 0, 0.9)',
        color: 'white',
        padding: '30px',
        borderRadius: '12px',
        textAlign: 'center',
        zIndex: 9999
      }}>
        <h2>✓ 区域选择完成</h2>
        <p>编辑功能开发中...</p>
        <button
          onClick={handleCancel}
          style={{
            marginTop: '20px',
            padding: '10px 20px',
            background: '#FF3B30',
            color: 'white',
            border: 'none',
            borderRadius: '6px',
            cursor: 'pointer',
            fontSize: '16px'
          }}
        >
          按 ESC 或点击此处取消
        </button>
      </div>
    </div>
  );
};

export default ScreenshotEditorWithDebug;
