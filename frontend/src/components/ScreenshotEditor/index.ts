/**
 * 截图编辑器三层架构
 *
 * 导出所有组件以便在其他地方使用
 */

export { MaskLayer } from './MaskLayer';
export { InteractionLayer } from './InteractionLayer';
export { ToolbarLayer } from './ToolbarLayer';
export { default as ScreenshotEditorLayers } from './ScreenshotEditorLayers';

// 导出类型
export type { SelectionRect } from './InteractionLayer';
