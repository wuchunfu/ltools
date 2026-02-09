// 插件类型定义
export enum PluginType {
  BuiltIn = 'builtin',
  Web = 'web',
  Native = 'native',
}

// 插件状态定义
export enum PluginState {
  Installed = 'installed',
  Enabled = 'enabled',
  Disabled = 'disabled',
  Error = 'error',
}

// 权限定义
export enum Permission {
  FileSystem = 'filesystem',
  Network = 'network',
  Clipboard = 'clipboard',
  Notification = 'notification',
  Process = 'process',
}

// 插件元数据接口
export interface PluginMetadata {
  id: string;
  name: string;
  version: string;
  author: string;
  description: string;
  icon?: string;
  type: PluginType;
  state: PluginState;
  permissions?: Permission[];
  keywords?: string[];
  homepage?: string;
  repository?: string;
  license?: string;
  enabled?: boolean;
  // 控制插件是否显示在侧边栏菜单中，默认 true
  showInMenu?: boolean;
  // 控制插件是否有独立的页面视图，默认 true
  hasPage?: boolean;
}

// 插件组件接口
export interface PluginComponent {
  id: string;
  metadata: PluginMetadata;
  Component: React.ComponentType;
}

// 插件事件类型
export interface PluginEvent {
  name: string;
  data: unknown;
  timestamp: number;
}
