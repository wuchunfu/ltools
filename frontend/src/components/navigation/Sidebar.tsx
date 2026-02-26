import { useMemo } from 'react'
import { useNavigate, useLocation } from 'react-router-dom'
import { Icon } from '../Icon'
import { usePlugins } from '../../plugins/usePlugins'
import { PluginState } from '../../../bindings/ltools/internal/plugins'
import type { IconName, NavItem } from '../../router/types'

/**
 * 基础导航项配置
 */
const baseNavItems: NavItem[] = [
  { id: 'home', label: '首页', icon: 'home', path: '/' },
  { id: 'plugins', label: '插件市场', icon: 'puzzle-piece', path: '/plugins' },
  { id: 'settings', label: '设置', icon: 'cog', path: '/settings' },
]

/**
 * 根据插件 ID 获取对应的图标
 */
const getPluginIconName = (pluginId: string): IconName => {
  const iconMap: Record<string, IconName> = {
    'calculator.builtin': 'calculator',
    'clipboard.builtin': 'clipboard',
    'jsoneditor.builtin': 'code',
    'processmanager.builtin': 'process',
    'sysinfo.builtin': 'cpu',
    'qrcode.builtin': 'qrcode',
    'hosts.builtin': 'server',
    'tunnel.builtin': 'network',
    'datetime.builtin': 'clock',
    'screenshot2.builtin': 'camera',
    'password.builtin': 'key',
    'kanban.builtin': 'view-columns',
  }
  return iconMap[pluginId] || 'puzzle-piece'
}

/**
 * 侧边栏组件
 */
export function Sidebar() {
  const navigate = useNavigate()
  const location = useLocation()

  // 获取已启用的插件用于动态菜单
  const { plugins } = usePlugins()
  const enabledPlugins = useMemo(() => {
    return plugins.filter(p => p.state === PluginState.PluginStateEnabled)
  }, [plugins])

  // 动态生成菜单项（基础菜单 + 已启用的插件）
  const navItems = useMemo(() => {
    const items: NavItem[] = [...baseNavItems]
    // 添加已启用的插件到菜单
    enabledPlugins.forEach(plugin => {
      const shouldShow = plugin.hasPage !== false && plugin.showInMenu !== false
      if (shouldShow) {
        items.push({
          id: `plugin-${plugin.id}`,
          label: plugin.name,
          icon: getPluginIconName(plugin.id),
          path: `/plugins/${plugin.id}`,
          pluginId: plugin.id,
        })
      }
    })
    return items
  }, [enabledPlugins])

  // 根据当前路径确定活动的导航项
  const getActiveId = (): string => {
    const path = location.pathname
    if (path === '/') return 'home'
    if (path === '/plugins') return 'plugins'
    if (path === '/settings') return 'settings'
    if (path.startsWith('/plugins/')) {
      const pluginId = path.replace('/plugins/', '')
      return `plugin-${pluginId}`
    }
    return 'home'
  }

  const activeId = getActiveId()

  return (
    <aside className="w-64 glass border-r border-white/10 flex flex-col">
      {/* Logo 区域 */}
      <div className="p-6 border-b border-white/10">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-[#7C3AED] to-[#A78BFA] flex items-center justify-center">
            <Icon name="cube" size={20} color="white" />
          </div>
          <div>
            <h1 className="text-xl font-bold text-white">LTools</h1>
            <p className="text-xs text-white/50">插件式工具箱</p>
          </div>
        </div>
      </div>

      {/* 导航菜单 */}
      <nav className="flex-1 p-3 space-y-1 overflow-y-auto scrollbar-hide">
        {navItems.map((item) => (
          <button
            key={item.id}
            className={`w-full flex items-center gap-3 px-4 py-3 rounded-lg transition-all duration-200 clickable ${
              activeId === item.id
                ? 'bg-[#7C3AED]/20 text-[#A78BFA] border border-[#7C3AED]/30'
                : 'text-white/60 hover:bg-white/5 hover:text-white/90'
            }`}
            onClick={() => navigate(item.path)}
          >
            <Icon name={item.icon} size={20} />
            <span className="font-medium">{item.label}</span>
          </button>
        ))}
      </nav>
    </aside>
  )
}
