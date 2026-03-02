import { Icon, IconName } from './Icon';

/**
 * 设置分类类型
 */
export type SettingsCategory = 'general' | 'shortcuts' | 'sync' | 'plugins' | 'about';

/**
 * 导航项配置
 */
interface NavItem {
  id: SettingsCategory;
  label: string;
  icon: IconName;
  description: string;
}

/**
 * 所有设置导航项
 */
const navItems: NavItem[] = [
  {
    id: 'general',
    label: '通用',
    icon: 'cog',
    description: '语言、主题、启动行为',
  },
  {
    id: 'shortcuts',
    label: '快捷键',
    icon: 'keyboard',
    description: '全局快捷键配置',
  },
  {
    id: 'sync',
    label: '同步',
    icon: 'cloud-arrow-up',
    description: '数据同步设置',
  },
  {
    id: 'plugins',
    label: '插件',
    icon: 'puzzle-piece',
    description: '插件启用/禁用、权限管理',
  },
  {
    id: 'about',
    label: '关于',
    icon: 'information-circle',
    description: '版本信息、更新检查',
  },
];

interface SettingsNavProps {
  activeCategory: SettingsCategory;
  onCategoryChange: (category: SettingsCategory) => void;
}

/**
 * 设置页面左侧导航组件
 */
export function SettingsNav({ activeCategory, onCategoryChange }: SettingsNavProps) {
  return (
    <nav className="w-56 flex-shrink-0">
      <div className="glass-light rounded-xl p-3">
        <ul className="space-y-1">
          {navItems.map((item) => (
            <li key={item.id}>
              <button
                onClick={() => onCategoryChange(item.id)}
                className={`
                  w-full flex items-center gap-3 px-3 py-2.5 rounded-lg transition-all duration-200 text-left
                  ${activeCategory === item.id
                    ? 'bg-[#7C3AED]/20 text-white border border-[#7C3AED]/30'
                    : 'text-white/60 hover:text-white hover:bg-white/5 border border-transparent'
                  }
                `}
              >
                <Icon
                  name={item.icon}
                  size={18}
                  color={activeCategory === item.id ? '#A78BFA' : 'currentColor'}
                />
                <span className="font-medium text-sm">{item.label}</span>
              </button>
            </li>
          ))}
        </ul>
      </div>
    </nav>
  );
}
