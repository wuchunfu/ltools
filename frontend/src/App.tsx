import { useState, useMemo, useEffect } from 'react'
import { PluginMarket } from "./components/PluginMarket";
import { DateTimeWidget, TimestampConverter } from "./components/DateTimeWidget";
import { ClipboardWidget } from "./components/ClipboardWidget";
import { SystemInfoWidget } from "./components/SystemInfoWidget";
import { CalculatorWidget } from "./components/CalculatorWidget";
import { JSONEditorWidget } from "./components/JSONEditorWidget";
import { ProcessManagerWidget } from "./components/ProcessManagerWidget";
import { PasswordGeneratorWidget } from "./components/PasswordGeneratorWidget";
import { ScreenshotWidget } from "./components/ScreenshotWidget";
import { Settings } from "./components/Settings";
import { Icon } from './components/Icon';
import { ToastProvider } from './contexts/ToastContext';
import { usePlugins } from './plugins/usePlugins';
import { PluginState } from '../bindings/ltools/internal/plugins';
import { getPluginIcon } from './utils/pluginHelpers';
import { Events } from '@wailsio/runtime';
import * as ShortcutService from '../bindings/ltools/internal/plugins/shortcutservice';
import * as SearchWindowService from '../bindings/ltools/internal/plugins/searchwindowservice';
import * as ProcessManagerService from '../bindings/ltools/plugins/processmanager/processmanagerservice';
import * as ScreenshotService from '../bindings/ltools/plugins/screenshot/screenshotservice';

/**
 * 导航项配置
 */
interface NavItem {
  id: string;
  label: string;
  icon: IconName;
  pluginId?: string;  // 如果是插件菜单项，存储插件ID
}

const baseNavItems: NavItem[] = [
  { id: 'home', label: '首页', icon: 'home' },
  { id: 'screenshot', label: '截图', icon: 'camera' },
  { id: 'datetime', label: '日期时间', icon: 'clock' },
  { id: 'password', label: '密码生成器', icon: 'key' },
  { id: 'plugins', label: '插件市场', icon: 'puzzle-piece' },
  { id: 'settings', label: '设置', icon: 'cog' },
];

type IconName = 'home' | 'puzzle-piece' | 'clock' | 'cog' | 'key' | 'shield-check' | 'funnel' |
  'check-circle' | 'x-circle' | 'exclamation-circle' | 'information-circle' |
  'search' | 'chevron-down' | 'chevron-right' | 'external-link' | 'refresh' |
  'clipboard' | 'document' | 'copy' | 'sparkles' | 'cube' | 'keyboard' | 'command' |
  'calculator' | 'cpu' | 'memory' | 'disk' | 'network' | 'server' | 'chip' |
  'code' | 'alert-circle' | 'check' | 'download' | 'upload' | 'process' | 'close' |
  'trash' | 'x-mark' | 'camera';

function App() {
  const [activeTab, setActiveTab] = useState<string>('home');
  const [previousActiveTab, setPreviousActiveTab] = useState<string>('');

  // 获取已启用的插件用于动态菜单
  const { plugins } = usePlugins();
  const enabledPlugins = useMemo(() => {
    return plugins.filter(p => p.state === PluginState.PluginStateEnabled);
  }, [plugins]);

  // 从后端加载快捷键配置
  const [shortcuts, setShortcuts] = useState<Record<string, string>>({});

  // 加载快捷键配置
  useEffect(() => {
    const loadShortcuts = async () => {
      try {
        console.log('[App] Loading shortcuts from backend...');
        const data = await ShortcutService.GetAllShortcuts();
        console.log('[App] Raw shortcuts data:', data);

        const shortcutMap: Record<string, string> = {};

        data.forEach(s => {
          if (s.enabled) {
            shortcutMap[s.keyCombo] = s.pluginId;
          }
        });

        setShortcuts(shortcutMap);
        console.log('[App] Shortcuts loaded successfully:', shortcutMap);
        console.log('[App] Total shortcuts:', Object.keys(shortcutMap).length);
      } catch (error) {
        console.error('[App] Failed to load shortcuts:', error);
      }
    };

    loadShortcuts();
  }, []);

  // 监听后端快捷键事件（真正的全局快捷键）
  useEffect(() => {
    console.log('[App] =======================================');
    console.log('[App] Setting up BACKEND shortcut event listener');
    console.log('[App] This listens for global shortcuts registered by Wails KeyBinding API');

    const unsubscribe = Events.On('shortcut:triggered', (ev: { data: string }) => {
      const pluginId = ev.data;
      console.log('[App] *** BACKEND SHORTCUT TRIGGERED ***');
      console.log('[App] Plugin ID:', pluginId);
      console.log('[App] Full event:', ev);
      console.log('[App] Current active tab:', activeTab);

      // 跳过搜索窗口快捷键（由单独的监听器处理）
      if (pluginId === 'search.window.builtin') {
        console.log('[App] Skipping search window shortcut (handled separately)');
        return;
      }

      // 跳过截图窗口快捷键（由独立窗口处理）
      if (pluginId === 'screenshot.window.builtin') {
        console.log('[App] Screenshot window shortcut triggered, calling Trigger()...');
        // 触发截图，ScreenshotEditor 组件会自己处理全屏请求
        ScreenshotService.Trigger()
          .catch((error: any) => {
            console.error('[App] Failed to trigger screenshot:', error);
          });
        return;
      }

      // 特殊处理 datetime 插件
      if (pluginId === 'datetime.builtin') {
        console.log('[App] Switching to datetime tab');
        setActiveTab('datetime');
      } else {
        const newTab = `plugin-${pluginId}`;
        console.log('[App] Switching to plugin tab:', newTab);
        setActiveTab(newTab);
      }
    });

    console.log('[App] Backend shortcut event listener REGISTERED successfully');
    console.log('[App] =======================================');

    return () => {
      unsubscribe();
      console.log('[App] Backend shortcut event listener removed');
    };
  }, [activeTab]);

  // 监听搜索窗口快捷键事件
  useEffect(() => {
    console.log('[App] Setting up search window shortcut listener');

    const unsubscribe = Events.On('shortcut:triggered', (ev: { data: string }) => {
      const pluginId = ev.data;

      // 检查是否是搜索窗口的快捷键
      if (pluginId === 'search.window.builtin') {
        console.log('[App] Search window shortcut triggered');
        SearchWindowService.Toggle().catch((error: any) => {
          console.error('[App] Failed to toggle search window:', error);
        });
      }
    });

    console.log('[App] Search window shortcut listener registered');

    return () => {
      unsubscribe();
      console.log('[App] Search window shortcut listener removed');
    };
  }, []);

  // 检测插件视图切换，通知后端
  useEffect(() => {
    if (activeTab === previousActiveTab) return;

    // 处理离开上一个插件
    if (previousActiveTab && previousActiveTab.startsWith('plugin-')) {
      const pluginId = previousActiveTab.replace('plugin-', '');
      handlePluginLeave(pluginId);
    }

    // 处理进入新插件
    if (activeTab && activeTab.startsWith('plugin-')) {
      const pluginId = activeTab.replace('plugin-', '');
      handlePluginEnter(pluginId);
    }

    setPreviousActiveTab(activeTab);
  }, [activeTab, previousActiveTab]);

  // 重新加载快捷键配置的辅助函数
  const reloadShortcuts = async () => {
    try {
      console.log('[App] Reloading shortcuts from backend...');
      const data = await ShortcutService.GetAllShortcuts();
      console.log('[App] Reloaded raw data:', data);

      const shortcutMap: Record<string, string> = {};

      data.forEach(s => {
        if (s.enabled) {
          shortcutMap[s.keyCombo] = s.pluginId;
        }
      });

      setShortcuts(shortcutMap);
      console.log('[App] Shortcuts reloaded successfully:', shortcutMap);
    } catch (error) {
      console.error('[App] Failed to reload shortcuts:', error);
    }
  };

  // 添加本地键盘事件调试（仅用于调试，可以看到所有按键事件）
  useEffect(() => {
    console.log('[App] Adding LOCAL keyboard event listener for DEBUG');

    const handleKeyDown = (e: KeyboardEvent) => {
      // 只记录有修饰键的组合
      if (e.metaKey || e.ctrlKey || e.altKey) {
        const parts: string[] = [];
        if (e.metaKey) parts.push('meta');
        if (e.ctrlKey) parts.push('ctrl');
        if (e.altKey) parts.push('alt');
        if (e.shiftKey) parts.push('shift');
        parts.push(e.key);

        const keyCombo = parts.join('+');
        console.log('[App] LOCAL keyboard event:', keyCombo);
        console.log('[App]  -> metaKey:', e.metaKey, 'ctrlKey:', e.ctrlKey, 'altKey:', e.altKey, 'shiftKey:', e.shiftKey, 'key:', e.key);
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    console.log('[App] LOCAL keyboard listener registered');

    return () => {
      window.removeEventListener('keydown', handleKeyDown);
      console.log('[App] LOCAL keyboard listener removed');
    };
  }, []);

  // 根据插件 ID 获取对应的图标
  const getPluginIcon = (pluginId: string): IconName => {
    const iconMap: Record<string, IconName> = {
      'screenshot.builtin': 'camera',
      'datetime.builtin': 'clock',
      'calculator.builtin': 'calculator',
      'clipboard.builtin': 'clipboard',
      'jsoneditor.builtin': 'code',
      'processmanager.builtin': 'process',
      'sysinfo.builtin': 'cpu',
      'search.window.builtin': 'search',
    };
    return iconMap[pluginId] || 'puzzle-piece';
  };

  // 动态生成菜单项（基础菜单 + 已启用的插件）
  const navItems = useMemo(() => {
    const items: NavItem[] = [...baseNavItems];
    // 添加已启用的插件到菜单，排除 showInMenu: false 的插件
    enabledPlugins.forEach(plugin => {
      // 排除有专门菜单项的插件（datetime, screenshot）
      // 排除 showInMenu: false 的插件
      if (plugin.id !== 'datetime.builtin' &&
          plugin.id !== 'screenshot.builtin' &&
          plugin.showInMenu !== false) {
        items.push({
          id: `plugin-${plugin.id}`,
          label: plugin.name,
          icon: getPluginIcon(plugin.id),
          pluginId: plugin.id,
        });
      }
    });
    return items;
  }, [enabledPlugins]);

  // 插件视图切换辅助函数
  const handlePluginEnter = async (pluginId: string) => {
    try {
      switch (pluginId) {
        case 'processmanager.builtin':
          await ProcessManagerService.EnterView();
          break;
        // 未来可扩展其他插件
      }
    } catch (err) {
      console.error(`Failed to enter plugin ${pluginId}:`, err);
    }
  };

  const handlePluginLeave = async (pluginId: string) => {
    try {
      switch (pluginId) {
        case 'processmanager.builtin':
          await ProcessManagerService.LeaveView();
          break;
        // 未来可扩展其他插件
      }
    } catch (err) {
      console.error(`Failed to leave plugin ${pluginId}:`, err);
    }
  };

  return (
    <ToastProvider>
      <div className="flex h-screen bg-[#0D0F1A] text-[#FAF5FF]">
        {/* 侧边栏 */}
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
        <nav className="flex-1 p-3 space-y-1">
          {navItems.map((item) => (
            <button
              key={item.id}
              className={`w-full flex items-center gap-3 px-4 py-3 rounded-lg transition-all duration-200 clickable ${
                activeTab === item.id
                  ? 'bg-[#7C3AED]/20 text-[#A78BFA] border border-[#7C3AED]/30'
                  : 'text-white/60 hover:bg-white/5 hover:text-white/90'
              }`}
              onClick={() => setActiveTab(item.id)}
            >
              <Icon name={item.icon} size={20} />
              <span className="font-medium">{item.label}</span>
            </button>
          ))}
        </nav>
      </aside>

      {/* 主内容区 */}
      <main className="flex-1 overflow-auto">
        {activeTab === 'home' && (
          <div className="p-8">
            <div className="max-w-2xl mx-auto">
              {/* Hero 区域 */}
              <div className="text-center mb-12 animate-fade-in">
                <div className="inline-flex items-center justify-center w-20 h-20 rounded-2xl bg-gradient-to-br from-[#7C3AED] to-[#A78BFA] mb-6">
                  <Icon name="sparkles" size={36} color="white" />
                </div>
                <h1 className="text-4xl font-bold mb-4 bg-gradient-to-r from-[#A78BFA] to-[#C4B5FD] bg-clip-text text-transparent">
                  LTools
                </h1>
                <p className="text-lg text-white/60 mb-2">
                  插件式工具箱
                </p>
                <p className="text-sm text-white/40">
                  基于 Wails v3 和 React 的跨平台桌面应用
                </p>
              </div>

              {/* 技术栈信息 */}
              <div className="mt-12 text-center text-white/30 text-sm">
                <p>使用 Wails v3 + React + TypeScript 构建</p>
              </div>
            </div>
          </div>
        )}

        {activeTab === 'plugins' && (
          <div className="p-6">
            <PluginMarket />
          </div>
        )}

        {activeTab === 'screenshot' && (
          <div className="p-8">
            <div className="max-w-2xl mx-auto">
              <div className="text-center mb-8">
                <div className="inline-flex items-center justify-center w-16 h-16 rounded-xl bg-gradient-to-br from-[#7C3AED] to-[#A78BFA] mb-4">
                  <Icon name="camera" size={28} color="white" />
                </div>
                <h1 className="text-3xl font-bold mb-2">截图工具</h1>
                <p className="text-white/50">微信风格的屏幕截图和标注工具</p>
              </div>
              <div className="space-y-8">
                <ScreenshotWidget />
              </div>
            </div>
          </div>
        )}

        {activeTab === 'datetime' && (
          <div className="p-8">
            <div className="max-w-2xl mx-auto">
              <div className="text-center mb-8">
                <div className="inline-flex items-center justify-center w-16 h-16 rounded-xl bg-gradient-to-br from-[#7C3AED] to-[#A78BFA] mb-4">
                  <Icon name="clock" size={28} color="white" />
                </div>
                <h1 className="text-3xl font-bold mb-2">日期时间插件</h1>
                <p className="text-white/50">实时显示当前时间和日期</p>
              </div>
              <div className="space-y-8">
                <DateTimeWidget />
                <TimestampConverter />
              </div>
            </div>
          </div>
        )}

        {activeTab === 'password' && (
          <div className="p-8">
            <div className="max-w-6xl mx-auto">
              <div className="text-center mb-8">
                <div className="inline-flex items-center justify-center w-20 h-20 rounded-2xl bg-gradient-to-br from-[#7C3AED] to-[#A78BFA] mb-6">
                  <Icon name="key" size={36} color="white" />
                </div>
                <h1 className="text-3xl font-bold mb-2">随机密码生成器</h1>
                <p className="text-white/50">生成安全的随机密码，支持自定义选项</p>
                <p className="text-sm text-white/30 mt-2">使用安全的随机数生成算法</p>
              </div>
              <PasswordGeneratorWidget />
            </div>
          </div>
        )}

        {activeTab === 'settings' && (
          <Settings
            shortcuts={shortcuts}
            onSetShortcut={async (pluginId, keyCombo) => {
              console.log('[App] =======================================');
              console.log('[App] SetShortcut called');
              console.log('[App]  -> pluginId:', pluginId);
              console.log('[App]  -> keyCombo:', keyCombo);
              try {
                console.log('[App] Calling ShortcutService.SetShortcut...');
                await ShortcutService.SetShortcut(keyCombo, pluginId);
                console.log('[App] SetShortcut SUCCESS, reloading shortcuts...');
                await reloadShortcuts();
                console.log('[App] =======================================');
              } catch (error: any) {
                console.error('[App] Failed to set shortcut:', error);
                throw error;
              }
            }}
            onRemoveShortcut={async (keyCombo) => {
              console.log('[App] =======================================');
              console.log('[App] RemoveShortcut called');
              console.log('[App]  -> keyCombo:', keyCombo);
              try {
                console.log('[App] Calling ShortcutService.RemoveShortcut...');
                await ShortcutService.RemoveShortcut(keyCombo);
                console.log('[App] RemoveShortcut SUCCESS, reloading shortcuts...');
                await reloadShortcuts();
                console.log('[App] =======================================');
              } catch (error: any) {
                console.error('[App] Failed to remove shortcut:', error);
                throw error;
              }
            }}
          />
        )}

        {/* 动态插件视图 - 使用缓存组件保持状态 */}
        <div className={activeTab.startsWith('plugin-') ? '' : 'hidden'}>
          <PluginView pluginId={activeTab.replace('plugin-', '')} plugins={plugins} />
        </div>
      </main>
    </div>
    </ToastProvider>
  )
}

/**
 * 插件视图组件 - 缓存所有插件组件以保持状态
 */
interface PluginViewProps {
  pluginId: string;
  plugins: Array<any>;
}

function PluginView({ pluginId, plugins }: PluginViewProps) {
  const plugin = plugins.find(p => p.id === pluginId);

  if (!plugin) {
    return (
      <div className="p-8 text-center">
        <div className="inline-flex items-center justify-center w-16 h-16 rounded-full bg-[#EF4444]/10 mb-4">
          <Icon name="exclamation-circle" size={32} color="#EF4444" />
        </div>
        <h2 className="text-xl font-semibold text-white mb-2">插件未找到</h2>
        <p className="text-white/60">插件 "{pluginId}" 不存在或未启用</p>
      </div>
    );
  }

  // 处理 hasPage: false 的插件
  if (plugin.hasPage === false) {
    return (
      <div className="p-8">
        <div className="max-w-2xl mx-auto text-center">
          <div className="inline-flex items-center justify-center w-20 h-20 rounded-2xl bg-gradient-to-br from-[#7C3AED] to-[#A78BFA] mb-6">
            <Icon name="cube" size={36} color="white" />
          </div>
          <h1 className="text-3xl font-bold mb-2">{plugin.name}</h1>
          <p className="text-white/50 mb-4">{plugin.description}</p>
          <div className="glass-light rounded-xl p-6">
            <p className="text-white/60">
              此插件通过快捷键或其他方式调用，无需独立页面。
            </p>
            {plugin.keywords && plugin.keywords.length > 0 && (
              <div className="mt-4">
                <p className="text-sm text-white/40 mb-2">关键词：</p>
                <div className="flex flex-wrap justify-center gap-2">
                  {plugin.keywords.map((kw: string, i: number) => (
                    <span key={i} className="px-2 py-1 rounded bg-[#7C3AED]/10 text-[#A78BFA] text-xs border border-[#7C3AED]/20">
                      {kw}
                    </span>
                  ))}
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    );
  }

  const pluginIcon = getPluginIcon(plugin);

  // 渲染所有插件组件，但只显示当前活动的插件
  // 这样可以保持所有组件的状态（如编辑器内容、光标位置等）
  return (
    <>
      {/* 剪贴板插件 */}
      {pluginId === 'clipboard.builtin' && plugins.some(p => p.id === 'clipboard.builtin') && (
        <div className="p-6">
          <div className="max-w-4xl mx-auto">
            <div className="mb-6">
              <div className="flex items-center gap-4">
                <div className="w-16 h-16 rounded-2xl bg-gradient-to-br from-[#7C3AED] to-[#A78BFA] flex items-center justify-center">
                  <Icon name="clipboard" size={28} color="white" />
                </div>
                <div>
                  <h1 className="text-3xl font-bold mb-1">剪贴板管理器</h1>
                  <p className="text-white/50">管理系统剪贴板历史记录</p>
                  <p className="text-sm text-white/30 mt-1">v1.0.0 · by LTools</p>
                </div>
              </div>
            </div>
            <ClipboardWidget />
          </div>
        </div>
      )}

      {/* 系统信息插件 */}
      {pluginId === 'sysinfo.builtin' && plugins.some(p => p.id === 'sysinfo.builtin') && (
        <div className="p-6">
          <div className="max-w-4xl mx-auto">
            <div className="text-center mb-8">
              <div className="inline-flex items-center justify-center w-20 h-20 rounded-2xl bg-gradient-to-br from-[#7C3AED] to-[#A78BFA] mb-6">
                <Icon name="information-circle" size={36} color="white" />
              </div>
              <h1 className="text-3xl font-bold mb-2">系统信息</h1>
              <p className="text-white/50">查看系统硬件和运行信息</p>
              <p className="text-sm text-white/30 mt-2">v1.0.0 · by LTools</p>
            </div>
            <SystemInfoWidget />
          </div>
        </div>
      )}

      {/* 计算器插件 */}
      {pluginId === 'calculator.builtin' && plugins.some(p => p.id === 'calculator.builtin') && (
        <div className="p-6">
          <div className="max-w-6xl mx-auto">
            <div className="text-center mb-8">
              <div className="inline-flex items-center justify-center w-20 h-20 rounded-2xl bg-gradient-to-br from-[#7C3AED] to-[#A78BFA] mb-6">
                <Icon name="calculator" size={36} color="white" />
              </div>
              <h1 className="text-3xl font-bold mb-2">计算器</h1>
              <p className="text-white/50">支持基本运算和科学计算</p>
              <p className="text-sm text-white/30 mt-2">v1.0.0 · by LTools</p>
            </div>
            <CalculatorWidget />
          </div>
        </div>
      )}

      {/* JSON 编辑器插件 */}
      {pluginId === 'jsoneditor.builtin' && plugins.some(p => p.id === 'jsoneditor.builtin') && (
        <div className="p-6">
          <div className="max-w-6xl mx-auto">
            <div className="text-center mb-8">
              <div className="inline-flex items-center justify-center w-20 h-20 rounded-2xl bg-gradient-to-br from-[#7C3AED] to-[#A78BFA] mb-6">
                <Icon name="code" size={36} color="white" />
              </div>
              <h1 className="text-3xl font-bold mb-2">JSON 编辑器</h1>
              <p className="text-white/50">格式化、验证和编辑 JSON 数据</p>
              <p className="text-sm text-white/30 mt-2">v1.0.0 · by LTools</p>
            </div>
            <JSONEditorWidget />
          </div>
        </div>
      )}

      {/* 进程管理器插件 */}
      {pluginId === 'processmanager.builtin' && plugins.some(p => p.id === 'processmanager.builtin') && (
        <div className="p-6">
          <div className="max-w-6xl mx-auto">
            <div className="text-center mb-8">
              <div className="inline-flex items-center justify-center w-20 h-20 rounded-2xl bg-gradient-to-br from-[#7C3AED] to-[#A78BFA] mb-6">
                <Icon name="process" size={36} color="white" />
              </div>
              <h1 className="text-3xl font-bold mb-2">进程管理器</h1>
              <p className="text-white/50">查看和管理系统运行中的进程</p>
              <p className="text-sm text-white/30 mt-2">v1.0.0 · by LTools</p>
            </div>
            <ProcessManagerWidget />
          </div>
        </div>
      )}

      {/* 默认插件界面 */}
      {!pluginId.startsWith('clipboard') && !pluginId.startsWith('sysinfo') && !pluginId.startsWith('calculator') && !pluginId.startsWith('jsoneditor') && !pluginId.startsWith('processmanager') && (
        <div className="p-8">
          <div className="max-w-4xl mx-auto">
            {/* 插件页头 */}
            <div className="text-center mb-8">
              <div className="inline-flex items-center justify-center w-20 h-20 rounded-2xl bg-gradient-to-br from-[#7C3AED] to-[#A78BFA] mb-6">
                <span className="text-4xl">{pluginIcon}</span>
              </div>
              <h1 className="text-3xl font-bold mb-2">{plugin.name}</h1>
              <p className="text-white/50">{plugin.description}</p>
              <p className="text-sm text-white/30 mt-2">
                v{plugin.version} · by {plugin.author}
              </p>
            </div>

            {/* 插件内容区域 */}
            <div className="glass-light rounded-xl p-8">
              <div className="text-center py-12">
                <Icon name="cube" size={48} color="rgba(167, 139, 250, 0.3)" />
                <p className="text-white/40 mt-4">插件功能正在开发中...</p>
                <p className="text-white/30 text-sm mt-2">
                  此插件 ({plugin.id}) 已成功启用，但尚未实现用户界面。
                </p>
              </div>
            </div>

            {/* 插件信息 */}
            <div className="mt-6 grid grid-cols-1 md:grid-cols-2 gap-4">
              {plugin.permissions && plugin.permissions.length > 0 && (
                <div className="glass-light rounded-xl p-4">
                  <h3 className="text-sm font-medium text-white/60 mb-2">所需权限</h3>
                  <div className="flex flex-wrap gap-2">
                    {plugin.permissions.map((perm: string, i: number) => (
                      <span key={i} className="px-2 py-1 rounded bg-[#F59E0B]/10 text-[#F59E0B] text-xs border border-[#F59E0B]/20">
                        {perm}
                      </span>
                    ))}
                  </div>
                </div>
              )}
              {plugin.keywords && plugin.keywords.length > 0 && (
                <div className="glass-light rounded-xl p-4">
                  <h3 className="text-sm font-medium text-white/60 mb-2">关键词</h3>
                  <div className="flex flex-wrap gap-2">
                    {plugin.keywords.map((kw: string, i: number) => (
                      <span key={i} className="px-2 py-1 rounded bg-[#7C3AED]/10 text-[#A78BFA] text-xs border border-[#7C3AED]/20">
                        {kw}
                      </span>
                    ))}
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>
      )}
    </>
  );
}

export default App
