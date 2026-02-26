import React from 'react'
import ReactDOM from 'react-dom/client'
import { AppRouter } from './router'
import './styles.css'

/**
 * 应用入口
 * 使用 React Router v6 统一管理路由
 */
ReactDOM.createRoot(document.getElementById('root') as HTMLElement).render(
  <React.StrictMode>
    <AppRouter />
  </React.StrictMode>,
)
