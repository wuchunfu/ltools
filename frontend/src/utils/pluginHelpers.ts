import { PluginMetadata } from '../../bindings/ltools/internal/plugins';

/**
 * 插件图标映射 - 使用通用 emoji 作为 fallback
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
 * 获取插件显示图标
 * 优先使用插件指定的 icon，否则根据 ID/名称查找，最后使用默认图标
 */
export function getPluginIcon(plugin: PluginMetadata): string {
  // 如果插件有指定图标且不是空字符串，使用它
  if (plugin.icon && plugin.icon.trim() !== '') {
    // 检查是否是一个有效的 emoji（简单检查：长度和字符范围）
    const str = plugin.icon.trim();
    // 基本的 emoji 检查 - 大多数 emoji 是由两个 16 位字符组成的代理对
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
