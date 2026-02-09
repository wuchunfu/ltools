import { useState, useEffect, useCallback } from 'react';
import { Icon } from './Icon';

/**
 * 快捷键信息接口
 */
interface ShortcutInfo {
  pluginId: string;
  keyCombo: string;
  displayText: string;
}

/**
 * ShortcutEditor 组件属性
 */
interface ShortcutEditorProps {
  pluginId: string;
  pluginName: string;
  currentShortcut?: ShortcutInfo;
  existingShortcuts: Record<string, string>;
  onSave: (keyCombo: string) => void;
  onCancel: () => void;
}

/**
 * 快捷键编辑器组件（纯前端实现）
 */
export function ShortcutEditor({ pluginId, pluginName, currentShortcut, existingShortcuts, onSave, onCancel }: ShortcutEditorProps) {
  const [isRecording, setIsRecording] = useState(false);
  const [recordedKeys, setRecordedKeys] = useState<string[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);
  const [displayShortcut, setDisplayShortcut] = useState<string>('');

  // 初始化显示当前快捷键
  useEffect(() => {
    if (currentShortcut) {
      setDisplayShortcut(currentShortcut.displayText);
      parseKeyCombo(currentShortcut.keyCombo);
    }
  }, [currentShortcut]);

  /**
   * 解析快捷键组合
   */
  const parseKeyCombo = (keyCombo: string) => {
    const normalized = keyCombo.toLowerCase();
    const parts = normalized.split('+');
    setRecordedKeys(parts);
  };

  /**
   * 格式化按键用于显示
   */
  const formatKeyForDisplay = (key: string): string => {
    const platform = navigator.platform.toLowerCase();
    const isMac = platform.includes('mac');

    switch (key.toLowerCase()) {
      case 'ctrl':
      case 'control':
        return isMac ? '⌘' : 'Ctrl';
      case 'cmd':
      case 'command':
      case 'meta':
        return isMac ? '⌘' : 'Win';
      case 'shift':
        return isMac ? '⇧' : 'Shift';
      case 'alt':
      case 'option':
        return isMac ? '⌥' : 'Alt';
      default:
        return key.toUpperCase();
    }
  };

  /**
   * 获取显示的快捷键文本
   */
  const getDisplayText = useCallback((): string => {
    if (recordedKeys.length === 0) {
      return '按下快捷键组合...';
    }

    // 分离修饰键和主键
    const modifiers: string[] = [];
    let mainKey = '';

    recordedKeys.forEach(key => {
      const lowerKey = key.toLowerCase();
      if (['ctrl', 'control', 'cmd', 'command', 'meta', 'shift', 'alt', 'option'].includes(lowerKey)) {
        modifiers.push(key);
      } else {
        mainKey = key;
      }
    });

    // 格式化显示
    const formattedModifiers = modifiers.map(formatKeyForDisplay);
    const formattedMainKey = mainKey ? formatKeyForDisplay(mainKey) : '';

    return [...formattedModifiers, formattedMainKey].filter(Boolean).join('+');
  }, [recordedKeys]);

  useEffect(() => {
    setDisplayShortcut(getDisplayText());
  }, [recordedKeys, getDisplayText]);

  /**
   * 处理键盘按下事件
   */
  const handleKeyDown = useCallback((e: KeyboardEvent) => {
    e.preventDefault();
    e.stopPropagation();

    const keys: string[] = [];

    // 收集修饰键
    if (e.ctrlKey) keys.push('ctrl');
    if (e.metaKey) keys.push('cmd');
    if (e.shiftKey) keys.push('shift');
    if (e.altKey) keys.push('alt');

    // 收集主键（排除修饰键）
    const mainKey = e.key.toLowerCase();
    if (!['control', 'meta', 'shift', 'alt'].includes(mainKey) && mainKey !== ' ') {
      keys.push(mainKey);
    }

    // 至少需要一个主键
    const hasMainKey = keys.some(k => !['ctrl', 'cmd', 'shift', 'alt'].includes(k.toLowerCase()));
    if (!hasMainKey) {
      return;
    }

    setRecordedKeys(keys);
    setError(null);
  }, []);

  /**
   * 处理键盘抬起事件 - 完成录制
   */
  const handleKeyUp = useCallback((e: KeyboardEvent) => {
    e.preventDefault();
    e.stopPropagation();

    if (recordedKeys.length > 0) {
      setIsRecording(false);
    }
  }, [recordedKeys.length]);

  /**
   * 开始录制快捷键
   */
  const startRecording = () => {
    setIsRecording(true);
    setRecordedKeys([]);
    setError(null);
  };

  /**
   * 停止录制
   */
  const stopRecording = () => {
    setIsRecording(false);
    setRecordedKeys([]);
    setError(null);
  };

  /**
   * 注册/注销键盘事件监听
   */
  useEffect(() => {
    if (isRecording) {
      window.addEventListener('keydown', handleKeyDown, { capture: true });
      window.addEventListener('keyup', handleKeyUp, { capture: true });
    }

    return () => {
      window.removeEventListener('keydown', handleKeyDown, { capture: true });
      window.removeEventListener('keyup', handleKeyUp, { capture: true });
    };
  }, [isRecording, handleKeyDown, handleKeyUp]);

  /**
   * 保存快捷键
   */
  const handleSave = async () => {
    if (recordedKeys.length === 0) {
      setError('请先录制快捷键');
      return;
    }

    try {
      setSaving(true);
      setError(null);

      // 构建快捷键字符串
      const keyCombo = recordedKeys.join('+');

      // 检查冲突（排除当前插件）
      const conflictingPlugin = existingShortcuts[keyCombo];
      if (conflictingPlugin && conflictingPlugin !== pluginId) {
        setError(`此快捷键已被其他插件使用`);
        return;
      }

      onSave(keyCombo);
    } catch (err: any) {
      setError(err.message || '保存失败');
    } finally {
      setSaving(false);
    }
  };

  /**
   * 清除当前快捷键
   */
  const handleClear = () => {
    setRecordedKeys([]);
    setError(null);
  };

  return (
    <div className="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50 p-4">
      <div className="glass rounded-2xl p-6 w-full max-w-md animate-fade-in">
        {/* 标题 */}
        <div className="flex items-center justify-between mb-6">
          <h2 className="text-xl font-semibold text-white">
            设置快捷键
          </h2>
          <button
            className="p-2 hover:bg-white/10 rounded-lg transition-all duration-200 clickable"
            onClick={onCancel}
            disabled={saving}
          >
            <Icon name="x-circle" size={20} color="rgba(255,255,255,0.6)" />
          </button>
        </div>

        {/* 插件名称 */}
        <div className="mb-6">
          <p className="text-white/60 text-sm">插件</p>
          <p className="text-white font-medium">{pluginName}</p>
        </div>

        {/* 快捷键录制区域 */}
        <div className="mb-6">
          <p className="text-white/60 text-sm mb-3">快捷键组合</p>
          <div
            className={`
              relative p-4 rounded-lg border-2 border-dashed transition-all duration-200
              ${isRecording
                ? 'border-[#7C3AED] bg-[#7C3AED]/10'
                : 'border-white/20 bg-[#0D0F1A]/50 hover:border-white/30'
              }
              ${saving ? 'opacity-50 pointer-events-none' : 'clickable'}
            `}
            onClick={isRecording ? undefined : startRecording}
          >
            {isRecording ? (
              <div className="text-center">
                <div className="flex items-center justify-center gap-2 mb-2">
                  <div className="w-2 h-2 rounded-full bg-[#EF4444] animate-pulse" />
                  <span className="text-white/80">录制中...</span>
                </div>
                <p className="text-white/50 text-sm">按下快捷键组合，松开完成</p>
              </div>
            ) : (
              <div className="text-center">
                {displayShortcut ? (
                  <div className="flex items-center justify-center gap-3">
                    <Icon name="keyboard" size={20} color="#A78BFA" />
                    <span className="text-[#A78BFA] font-mono text-lg">{displayShortcut}</span>
                  </div>
                ) : (
                  <div className="text-white/40">
                    点击开始录制快捷键
                  </div>
                )}
              </div>
            )}

            {isRecording && (
              <button
                className="mt-3 w-full py-2 bg-white/10 hover:bg-white/20 rounded-lg text-white/80 text-sm transition-all duration-200 clickable"
                onClick={stopRecording}
              >
                取消录制
              </button>
            )}
          </div>
        </div>

        {/* 错误提示 */}
        {error && (
          <div className="mb-6 p-3 bg-[#EF4444]/10 border border-[#EF4444]/20 rounded-lg">
            <div className="flex items-center gap-2 text-[#EF4444] text-sm">
              <Icon name="exclamation-circle" size={16} />
              <span>{error}</span>
            </div>
          </div>
        )}

        {/* 操作按钮 */}
        <div className="flex gap-3">
          <button
            className="flex-1 py-3 bg-white/5 hover:bg-white/10 text-white/80 rounded-lg transition-all duration-200 clickable font-medium"
            onClick={onCancel}
            disabled={saving}
          >
            取消
          </button>
          {recordedKeys.length > 0 && (
            <button
              className="px-4 py-3 bg-[#EF4444]/20 hover:bg-[#EF4444]/30 text-[#EF4444] rounded-lg transition-all duration-200 clickable font-medium"
              onClick={handleClear}
              disabled={saving}
            >
              <Icon name="x-circle" size={16} />
            </button>
          )}
          <button
            className={`
              flex-1 py-3 rounded-lg transition-all duration-200 clickable font-medium
              ${recordedKeys.length > 0
                ? 'bg-[#7C3AED] hover:bg-[#6D28D9] text-white'
                : 'bg-white/5 text-white/40 cursor-not-allowed'
              }
            `}
            onClick={handleSave}
            disabled={saving || recordedKeys.length === 0}
          >
            {saving ? '保存中...' : '保存'}
          </button>
        </div>

        {/* 提示信息 */}
        <div className="mt-6 p-3 bg-white/5 rounded-lg">
          <p className="text-white/40 text-xs">
            💡 提示：可以使用 Ctrl、Shift、Alt、Cmd (macOS) 等修饰键组合。例如：
          </p>
          <div className="mt-2 space-y-1">
            <p className="text-white/30 text-xs font-mono">• Cmd+Shift+D (macOS)</p>
            <p className="text-white/30 text-xs font-mono">• Ctrl+Shift+D (Windows/Linux)</p>
            <p className="text-white/30 text-xs font-mono">• Alt+Space</p>
          </div>
        </div>
      </div>
    </div>
  );
}
