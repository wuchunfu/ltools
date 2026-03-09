import { PluginMetadata } from '../../bindings/ltools/internal/plugins';
import type { IconName } from '../components/Icon';

/**
 * 插件 ID 到图标名称的映射
 * 优先使用专业的 SVG 图标，fallback 到 emoji
 */
export const PLUGIN_ICON_MAP: Record<string, { icon?: IconName; emoji: string }> = {
  // 内置插件
  'datetime.builtin': { icon: 'clock', emoji: '🕐' },
  'calculator.builtin': { icon: 'calculator', emoji: '🧮' },
  'clipboard.builtin': { icon: 'clipboard', emoji: '📋' },
  'jsoneditor.builtin': { icon: 'code-bracket', emoji: '📝' },
  'processmanager.builtin': { icon: 'process', emoji: '⚙️' },
  'screenshot.builtin': { icon: 'camera', emoji: '📷' },
  'screenshot2.builtin': { icon: 'camera', emoji: '📷' },
  'sysinfo.builtin': { icon: 'server', emoji: '💻' },
  'applauncher.builtin': { icon: 'grid', emoji: '🚀' },
  'bookmark.builtin': { icon: 'bookmark', emoji: '🔖' },
  'qrcode.builtin': { icon: 'qrcode', emoji: '📱' },
  'hosts.builtin': { icon: 'network', emoji: '🌐' },
  'tunnel.builtin': { icon: 'network', emoji: '🔧' },
  'kanban.builtin': { icon: 'kanban', emoji: '📋' },
  'markdown.builtin': { icon: 'document', emoji: '📄' },
  'imagebed.builtin': { icon: 'photo', emoji: '🖼️' },
  'sticky.builtin': { icon: 'pin', emoji: '📌' },
  'musicplayer.builtin': { icon: 'heart', emoji: '🎵' },
  'vault.builtin': { icon: 'shield-check', emoji: '🔐' },
  'ipinfo.builtin': { icon: 'globe', emoji: '🌍' },
  'localtranslate.builtin': { icon: 'language', emoji: '🌐' },
  'password.builtin': { icon: 'key', emoji: '🔑' },
};

/**
 * 默认 emoji 图标映射 - 用于 fallback
 */
export const PLUGIN_ICONS: Record<string, string> = {
  // 默认图标
  'default': '📦',
  // 计算器相关
  'calculator': '🧮',
  'calc': '🧮',
  'math': '🔢',
  // 时间相关
  'datetime': '🕐',
  'time': '⏰',
  'date': '📅',
  'clock': '🕰️',
  // 系统相关
  'sysinfo': '💻',
  'system': '⚙️',
  'info': 'ℹ️',
  // 剪贴板相关
  'clipboard': '📋',
  'copy': '📄',
  // 通用工具
  'tool': '🔧',
  'utility': '🛠️',
  'plugin': '🔌',
  // 应用相关
  'app': '📱',
  'extension': '➕',
};

/**
 * 获取插件显示图标 (emoji)
 * 优先使用插件指定的 icon，否则根据 ID/名称查找，最后使用默认图标
 */
export function getPluginIcon(plugin: PluginMetadata): string {
  // 首先检查映射表
  const mapping = PLUGIN_ICON_MAP[plugin.id];
  if (mapping) {
    return mapping.emoji;
  }

  // 如果插件有指定图标且不是空字符串，使用它
  if (plugin.icon && plugin.icon.trim() !== '') {
    const str = plugin.icon.trim();
    // 基本的 emoji 检查
    if (str.length <= 4 && /^[\p{Emoji}\p{Emoji_Component}]+$/u.test(str)) {
      return str;
    }
  }

  // 根据插件 ID 查找映射
  const pluginIdLower = plugin.id.toLowerCase();
  for (const [key, icon] of Object.entries(PLUGIN_ICONS)) {
    if (key !== 'default' && pluginIdLower.includes(key)) {
      return icon;
    }
  }

  // 根据插件名称查找映射
  const pluginNameLower = plugin.name.toLowerCase();
  for (const [key, icon] of Object.entries(PLUGIN_ICONS)) {
    if (key !== 'default' && pluginNameLower.includes(key)) {
      return icon;
    }
  }

  // 使用默认图标
  return PLUGIN_ICONS['default'];
}

/**
 * 获取插件的 Icon 组件图标名称
 * 返回 IconName 如果有对应的专业图标，否则返回 null（需要使用 emoji fallback）
 */
export function getPluginIconName(plugin: PluginMetadata): IconName | null {
  // 首先检查映射表
  const mapping = PLUGIN_ICON_MAP[plugin.id];
  if (mapping?.icon) {
    return mapping.icon;
  }

  // 如果插件的 icon 字段是一个图标名称而不是 emoji，尝试使用它
  if (plugin.icon && plugin.icon.trim() !== '') {
    const str = plugin.icon.trim();
    // 如果不是 emoji，可能是一个图标名称
    if (!/^[\p{Emoji}\p{Emoji_Component}]+$/u.test(str)) {
      return str as IconName;
    }
  }

  return null;
}
