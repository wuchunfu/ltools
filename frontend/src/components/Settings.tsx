import { useState } from 'react';
import { Icon } from './Icon';
import { usePlugins } from '../plugins/usePlugins';
import { useToast } from '../hooks/useToast';
import { PluginState } from '../../bindings/ltools/internal/plugins';
import { getPluginIcon } from '../utils/pluginHelpers';
import { ShortcutEditor } from './ShortcutEditor';

/**
 * 快捷键信息接口
 */
interface ShortcutInfo {
  pluginId: string;
  keyCombo: string;
  displayText: string;
}

/**
 * 插件快捷键项
 */
interface PluginShortcutItem {
  pluginId: string;
  pluginName: string;
  pluginIcon: string;
  shortcut?: ShortcutInfo;
}

interface SettingsProps {
  shortcuts: Record<string, string>;
  onSetShortcut: (pluginId: string, keyCombo: string) => Promise<void>;
  onRemoveShortcut: (keyCombo: string) => Promise<void>;
}

/**
 * 设置页面组件
 */
export function Settings({ shortcuts, onSetShortcut, onRemoveShortcut }: SettingsProps) {
  const { plugins } = usePlugins();
  const { success, error } = useToast();
  const [editingPlugin, setEditingPlugin] = useState<string | null>(null);

  // 获取已启用的插件
  const enabledPlugins = plugins.filter(p => p.state === PluginState.PluginStateEnabled);

  // 构建 pluginId -> keyCombo 的映射
  const pluginToShortcut: Record<string, string> = {};
  Object.entries(shortcuts).forEach(([keyCombo, pluginId]) => {
    pluginToShortcut[pluginId] = keyCombo;
  });

  // 为每个已启用的插件创建快捷键项
  const pluginShortcuts: PluginShortcutItem[] = enabledPlugins.map(plugin => {
    const keyCombo = pluginToShortcut[plugin.id];
    return {
      pluginId: plugin.id,
      pluginName: plugin.name,
      pluginIcon: getPluginIcon(plugin),
      shortcut: keyCombo ? {
        pluginId: plugin.id,
        keyCombo: keyCombo,
        displayText: formatShortcutDisplay(keyCombo),
      } : undefined,
    };
  });

  const handleSetShortcut = async (pluginId: string, keyCombo: string) => {
    try {
      await onSetShortcut(pluginId, keyCombo);
      success(`快捷键已设置: ${keyCombo}`);
    } catch (err: any) {
      error(`设置失败: ${err.message || err}`);
    }
  };

  const handleRemoveShortcut = async (keyCombo: string) => {
    try {
      await onRemoveShortcut(keyCombo);
      success(`快捷键已移除`);
    } catch (err: any) {
      error(`移除失败: ${err.message || err}`);
    }
  };

  return (
    <div className="p-6 max-w-4xl mx-auto">
      {/* 页面标题 */}
      <div className="text-center mb-8">
        <div className="inline-flex items-center justify-center w-20 h-20 rounded-2xl bg-gradient-to-br from-[#7C3AED] to-[#A78BFA] mb-6">
          <Icon name="cog" size={36} color="white" />
        </div>
        <h1 className="text-3xl font-bold mb-2">设置</h1>
        <p className="text-white/50">配置应用和插件快捷键</p>
      </div>

      {/* 快捷键设置 */}
      <div className="glass-light rounded-xl p-6 mb-8">
        <h2 className="text-xl font-semibold text-white mb-4 flex items-center gap-2">
          <Icon name="keyboard" size={20} color="#A78BFA" />
          快捷键设置
        </h2>
        <p className="text-white/50 text-sm mb-6">
          为已启用的插件设置全局快捷键，快速打开对应的插件页面
        </p>

        {pluginShortcuts.length === 0 ? (
          <div className="text-center py-12 text-white/50">
            <Icon name="keyboard" size={32} color="rgba(167, 139, 250, 0.3)" className="mx-auto mb-3" />
            <p>暂无已启用的插件</p>
          </div>
        ) : (
          <div className="space-y-3">
            {pluginShortcuts.map((item) => (
              <ShortcutItem
                key={item.pluginId}
                item={item}
                onEdit={() => setEditingPlugin(item.pluginId)}
                onRemove={item.shortcut ? () => handleRemoveShortcut(item.shortcut!.keyCombo) : undefined}
              />
            ))}
          </div>
        )}
      </div>

      {/* 通用设置（预留） */}
      <div className="glass-light rounded-xl p-6">
        <h2 className="text-xl font-semibold text-white mb-4 flex items-center gap-2">
          <Icon name="cog" size={20} color="#A78BFA" />
          通用设置
        </h2>
        <p className="text-white/50 text-sm mb-6">
          更多设置选项即将推出...
        </p>
        <div className="text-center py-8 text-white/30">
          <Icon name="cog" size={32} color="rgba(167, 139, 250, 0.2)" className="mx-auto mb-3" />
          <p>通用设置功能开发中</p>
        </div>
      </div>

      {/* 快捷键编辑器 */}
      {editingPlugin && (
        <ShortcutEditor
          pluginId={editingPlugin}
          pluginName={pluginShortcuts.find(p => p.pluginId === editingPlugin)?.pluginName || ''}
          currentShortcut={pluginShortcuts.find(p => p.pluginId === editingPlugin)?.shortcut}
          existingShortcuts={shortcuts}
          onSave={(keyCombo) => handleSetShortcut(editingPlugin, keyCombo)}
          onCancel={() => setEditingPlugin(null)}
        />
      )}
    </div>
  );
}

/**
 * 单个快捷键项
 */
interface ShortcutItemProps {
  item: PluginShortcutItem;
  onEdit: () => void;
  onRemove?: () => void;
}

function ShortcutItem({ item, onEdit, onRemove }: ShortcutItemProps) {
  return (
    <div className="flex items-center justify-between p-4 bg-[#0D0F1A]/50 rounded-lg border border-white/10 hover:border-white/20 transition-all duration-200">
      {/* 插件信息 */}
      <div className="flex items-center gap-3 flex-1 min-w-0">
        <div className="w-10 h-10 rounded-lg bg-gradient-to-br from-[#7C3AED]/20 to-[#A78BFA]/20 flex items-center justify-center text-lg flex-shrink-0">
          {item.pluginIcon}
        </div>
        <div className="min-w-0">
          <h3 className="text-white font-medium truncate">{item.pluginName}</h3>
          {item.shortcut ? (
            <div className="flex items-center gap-2 mt-1">
              <kbd className="px-2 py-1 text-xs font-mono bg-white/10 rounded border border-white/20 text-[#A78BFA]">
                {item.shortcut.displayText}
              </kbd>
            </div>
          ) : (
            <p className="text-sm text-white/40 mt-1">未设置快捷键</p>
          )}
        </div>
      </div>

      {/* 操作按钮 */}
      <div className="flex items-center gap-2 flex-shrink-0 ml-4">
        {item.shortcut ? (
          <>
            <button
              className="p-2 rounded-lg hover:bg-white/10 text-white/60 hover:text-white transition-all duration-200 clickable"
              onClick={onEdit}
              title="编辑快捷键"
            >
              <Icon name="refresh" size={16} />
            </button>
            <button
              className="p-2 rounded-lg hover:bg-[#EF4444]/10 text-white/40 hover:text-[#EF4444] transition-all duration-200 clickable"
              onClick={onRemove}
              title="移除快捷键"
            >
              <Icon name="x-circle" size={16} />
            </button>
          </>
        ) : (
          <button
            className="px-4 py-2 bg-[#7C3AED]/20 hover:bg-[#7C3AED]/30 text-[#A78BFA] rounded-lg transition-all duration-200 clickable text-sm font-medium"
            onClick={onEdit}
          >
            设置快捷键
          </button>
        )}
      </div>
    </div>
  );
}

/**
 * 格式化快捷键用于显示
 */
function formatShortcutDisplay(keyCombo: string): string {
  const platform = navigator.platform.toLowerCase();
  const isMac = platform.includes('mac');

  const parts = keyCombo.split('+');
  const modifiers: string[] = [];
  let mainKey = '';

  parts.forEach(part => {
    switch (part) {
      case 'ctrl':
        modifiers.push(isMac ? '⌘' : 'Ctrl');
        break;
      case 'cmd':
        modifiers.push(isMac ? '⌘' : 'Win');
        break;
      case 'shift':
        modifiers.push(isMac ? '⇧' : 'Shift');
        break;
      case 'alt':
        modifiers.push(isMac ? '⌥' : 'Alt');
        break;
      default:
        mainKey = part.toUpperCase();
    }
  });

  return [...modifiers, mainKey].filter(Boolean).join('+');
}
