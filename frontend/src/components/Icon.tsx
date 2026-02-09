/**
 * Icon 组件 - 通用 SVG 图标组件
 * 使用 Heroicons 图标集，确保一致的图标风格
 */

export interface IconProps {
  /**
   * 图标名称
   */
  name: IconName;
  /**
   * 图标尺寸
   * @default 24
   */
  size?: number;
  /**
   * 自定义类名
   */
  className?: string;
  /**
   * 图标颜色（使用 Tailwind 类）
   * @default "currentColor"
   */
  color?: string;
}

export type IconName =
  // 导航图标
  | 'home'
  | 'puzzle-piece'
  | 'clock'
  | 'cog'
  | 'key'
  | 'shield-check'
  | 'funnel'
  | 'camera'
  // 状态图标
  | 'check-circle'
  | 'x-circle'
  | 'exclamation-circle'
  | 'information-circle'
  // 操作图标
  | 'search'
  | 'chevron-down'
  | 'chevron-right'
  | 'external-link'
  | 'refresh'
  // 剪贴板相关
  | 'clipboard'
  | 'document'
  | 'copy'
  // 系统信息相关
  | 'cpu'
  | 'memory'
  | 'disk'
  | 'network'
  | 'server'
  | 'chip'
  // 进程管理相关
  | 'process'
  | 'close'
  // 快捷键相关
  | 'keyboard'
  | 'command'
  // 其他
  | 'sparkles'
  | 'cube'
  | 'calculator'
  | 'code'
  | 'alert-circle'
  | 'check'
  | 'download'
  | 'upload'
  | 'trash'
  | 'x-mark';

const iconPaths: Record<IconName, string> = {
  // 导航图标
  home: 'M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6',
  'puzzle-piece': 'M11 4a2 2 0 114 0v1a1 1 0 001 1h3a1 1 0 011 1v3a1 1 0 01-1 1h-1a2 2 0 100 4h1a1 1 0 011 1v3a1 1 0 01-1 1h-3a1 1 0 01-1-1v-1a2 2 0 10-4 0v1a1 1 0 01-1 1H7a1 1 0 01-1-1v-3a1 1 0 00-1-1H4a2 2 0 110-4h1a1 1 0 001-1V7a1 1 0 011-1h3a1 1 0 001-1V4z',
  clock: 'M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z',
  cog: 'M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z M15 12a3 3 0 11-6 0 3 3 0 016 0z',
  key: 'M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z',
  'shield-check': 'M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z',
  funnel: 'M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z',
  camera: 'M3 9a2 2 0 012-2h.93a2 2 0 001.664-.89l.812-1.22A2 2 0 0110.07 4h3.86a2 2 0 011.664.89l.812 1.22A2 2 0 0018.07 7H19a2 2 0 012 2v9a2 2 0 01-2 2H5a2 2 0 01-2-2V9z M15 13a3 3 0 11-6 0 3 3 0 016 0z',

  // 状态图标
  'check-circle': 'M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z',
  'x-circle': 'M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z',
  'exclamation-circle': 'M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z',
  'information-circle': 'M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z',
  'alert-circle': 'M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z',
  'check': 'M5 13l4 4L19 7',
  'download': 'M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4',
  'upload': 'M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12',
  'code': 'M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4',

  // 操作图标
  search: 'M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z',
  'chevron-down': 'M19 9l-7 7-7-7',
  'chevron-right': 'M9 5l7 7-7 7',
  'external-link': 'M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14',
  refresh: 'M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15',
  close: 'M6 18L18 6M6 6l12 12',

  // 剪贴板相关
  clipboard: 'M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2',
  document: 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5l2 2h5a2 2 0 012 2v11a2 2 0 01-2 2z',
  copy: 'M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z',
  trash: 'M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16',
  'x-mark': 'M6 18L18 6M6 6l12 12',

  // 快捷键相关
  keyboard: 'M4 6a2 2 0 012-2h12a2 2 0 012 2v12a2 2 0 01-2 2H6a2 2 0 01-2-2V6zm2 0h12v12H6V6zm3 5h2m-2 4h2m4-4h2m-2 4h2',
  command: 'M4 8V6a2 2 0 012-2h12a2 2 0 012 2v2M4 16v2a2 2 0 002 2h12a2 2 0 002-2v-2M9 9l6 6M15 9l-6 6',

  // 系统信息相关
  cpu: 'M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6m-6 4h6m-6 4h6',
  memory: 'M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10',
  disk: 'M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4',
  network: 'M8.111 16.404a5.5 5.5 0 017.778 0M12 20h.01m-7.08-7.071c3.904-3.905 10.236-3.905 14.141 0M1.394 9.393c5.857-5.857 15.355-5.857 21.213 0',
  server: 'M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01',
  chip: 'M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z',

  // 进程管理相关
  process: 'M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-3 7h3m-3 4h3m-6-4h.01M9 16h.01M9 9h.01M12 9h.01M15 9h.01M9 12h6',

  // 其他
  sparkles: 'M5 3v4M3 5h4M6 17v4m-2-2h4m5-16l2.286 6.857L21 12l-5.714 2.143L13 21l-2.286-6.857L5 12l5.714-2.143L13 3z',
  cube: 'M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4',
  calculator: 'M9 7h6m0 10v-3m-3 3h.01M9 17h.01M9 14h.01M12 14h.01M15 14h.01M9 11h6M5 21h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v14a2 2 0 002 2z',
};

const viewBoxes: Record<IconName, string> = {
  cog: '0 0 24 24',
  sparkles: '0 0 24 24',
  cube: '0 0 24 24',
  clipboard: '0 0 24 24',
  document: '0 0 24 24',
  copy: '0 0 24 24',
  trash: '0 0 24 24',
  'x-mark': '0 0 24 24',
  calculator: '0 0 24 24',
  cpu: '0 0 24 24',
  memory: '0 0 24 24',
  disk: '0 0 24 24',
  network: '0 0 24 24',
  server: '0 0 24 24',
  chip: '0 0 24 24',
  keyboard: '0 0 24 24',
  command: '0 0 24 24',
  process: '0 0 24 24',
  close: '0 0 24 24',
  key: '0 0 24 24',
  'shield-check': '0 0 24 24',
  funnel: '0 0 24 24',
  camera: '0 0 24 24',
  // 默认 viewBox
  home: '0 0 24 24',
  'puzzle-piece': '0 0 24 24',
  clock: '0 0 24 24',
  'check-circle': '0 0 24 24',
  'x-circle': '0 0 24 24',
  'exclamation-circle': '0 0 24 24',
  'information-circle': '0 0 24 24',
  'alert-circle': '0 0 24 24',
  'check': '0 0 24 24',
  'download': '0 0 24 24',
  'upload': '0 0 24 24',
  'code': '0 0 24 24',
  search: '0 0 24 24',
  'chevron-down': '0 0 24 24',
  'chevron-right': '0 0 24 24',
  'external-link': '0 0 24 24',
  refresh: '0 0 24 24',
};

export function Icon({ name, size = 24, className = '', color = 'currentColor' }: IconProps) {
  const path = iconPaths[name];
  const viewBox = viewBoxes[name] || '0 0 24 24';

  return (
    <svg
      width={size}
      height={size}
      viewBox={viewBox}
      fill="none"
      stroke={color}
      strokeWidth={2}
      strokeLinecap="round"
      strokeLinejoin="round"
      className={className}
      xmlns="http://www.w3.org/2000/svg"
      aria-hidden="true"
    >
      <path d={path} />
    </svg>
  );
}
