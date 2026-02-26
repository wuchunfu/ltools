import { useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { Events } from '@wailsio/runtime'
import * as SearchWindowService from '../../bindings/ltools/internal/plugins/searchwindowservice'
import * as Screenshot2Service from '../../bindings/ltools/plugins/screenshot2/screenshot2service'

/**
 * 全局快捷键处理 Hook
 * 监听后端发送的快捷键事件并进行相应导航
 */
export function useGlobalShortcuts() {
  const navigate = useNavigate()

  useEffect(() => {
    console.log('[useGlobalShortcuts] Setting up backend shortcut event listener')

    const unsubscribe = Events.On('shortcut:triggered', (ev: { data: string }) => {
      const pluginId = ev.data
      console.log('[useGlobalShortcuts] Shortcut triggered:', pluginId)

      // 跳过搜索窗口快捷键（由单独的监听器处理）
      if (pluginId === 'search.window.builtin') {
        console.log('[useGlobalShortcuts] Handling search window shortcut')
        SearchWindowService.Toggle().catch((error: any) => {
          console.error('[useGlobalShortcuts] Failed to toggle search window:', error)
        })
        return
      }

      // 截图快捷键（由独立窗口处理）
      if (pluginId === 'screenshot2.window.builtin') {
        console.log('[useGlobalShortcuts] Handling screenshot2 shortcut')
        Screenshot2Service.StartCapture().catch((error: any) => {
          console.error('[useGlobalShortcuts] Failed to trigger screenshot2:', error)
        })
        return
      }

      // 导航到插件页面
      console.log('[useGlobalShortcuts] Navigating to plugin:', pluginId)
      navigate(`/plugins/${pluginId}`)
    })

    console.log('[useGlobalShortcuts] Event listener registered')

    return () => {
      unsubscribe()
      console.log('[useGlobalShortcuts] Event listener removed')
    }
  }, [navigate])

  // 本地键盘事件调试（仅用于调试）
  useEffect(() => {
    console.log('[useGlobalShortcuts] Adding local keyboard debug listener')

    const handleKeyDown = (e: KeyboardEvent) => {
      // 只记录有修饰键的组合
      if (e.metaKey || e.ctrlKey || e.altKey) {
        const parts: string[] = []
        if (e.metaKey) parts.push('meta')
        if (e.ctrlKey) parts.push('ctrl')
        if (e.altKey) parts.push('alt')
        if (e.shiftKey) parts.push('shift')
        parts.push(e.key)

        const keyCombo = parts.join('+')
        console.log('[useGlobalShortcuts] Local keyboard event:', keyCombo)
      }
    }

    window.addEventListener('keydown', handleKeyDown)

    return () => {
      window.removeEventListener('keydown', handleKeyDown)
    }
  }, [])
}
