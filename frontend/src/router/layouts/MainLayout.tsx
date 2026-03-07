import { Outlet, useNavigate } from 'react-router-dom'
import { useEffect } from 'react'
import { Events } from '@wailsio/runtime'
import { ToastProvider } from '../../contexts/ToastContext'
import { Sidebar } from '../../components/navigation/Sidebar'
import { useGlobalShortcuts } from '../../hooks/useGlobalShortcuts'
import { UpdateNotification } from '../../widgets/UpdateNotificationWidget'

// Re-export IconName for use in other files
export type { IconName } from '../types'

/**
 * 主布局组件
 * 包含侧边栏和主内容区域
 */
export function MainLayout() {
  const navigate = useNavigate()

  // 初始化全局快捷键监听
  useGlobalShortcuts()

  // 监听来自后端的全局导航事件
  useEffect(() => {
    const unsubscribe = Events.On('navigate:to', (ev: any) => {
      console.log('[MainLayout] Received navigate:to event:', ev)
      const data = ev.data as { path: string; tab?: string }
      if (data.path) {
        // 构建完整的 URL（包含查询参数）
        const url = data.tab ? `${data.path}?tab=${data.tab}` : data.path
        navigate(url, { replace: false })
      }
    })

    return () => {
      if (unsubscribe && typeof unsubscribe === 'function') {
        unsubscribe()
      }
    }
  }, [navigate])

  return (
    <ToastProvider>
      <div className="flex h-screen bg-[#0D0F1A] text-[#FAF5FF]">
        {/* 侧边栏 */}
        <Sidebar />

        {/* 主内容区 - 由子路由填充 */}
        <main className="flex-1 overflow-auto relative">
          <Outlet />
        </main>
      </div>

      {/* 更新通知组件（全局） */}
      <UpdateNotification />
    </ToastProvider>
  )
}
