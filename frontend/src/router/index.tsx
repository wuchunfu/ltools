import { createBrowserRouter, RouterProvider } from 'react-router-dom'
import { mainRoutes } from './routes/mainRoutes'
import { windowRoutes, isWindowPath } from './routes/windowRoutes'

/**
 * 创建路由实例
 * 根据当前路径判断使用窗口路由还是主应用路由
 */
export function createAppRouter() {
  const path = window.location.pathname

  // 如果是独立窗口路径，使用窗口路由
  if (isWindowPath(path)) {
    return createBrowserRouter(windowRoutes)
  }

  // 否则使用主应用路由
  return createBrowserRouter(mainRoutes)
}

/**
 * 路由组件
 */
export function AppRouter() {
  const router = createAppRouter()
  return <RouterProvider router={router} />
}

// 导出类型和工具
export type { RouteConfig, PluginRouteConfig, NavItem, IconName } from './types'
export { registerPluginLifecycle } from './guards/pluginGuard'
