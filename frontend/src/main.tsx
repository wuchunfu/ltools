import React from 'react'
import ReactDOM from 'react-dom/client'
import App from './App'
import { SearchWindow } from './components/SearchWindow'
import ScreenshotEditor from './components/ScreenshotEditor'
import { ToastProvider } from './contexts/ToastContext'
import './styles.css'

// Get the current path from window.location
const path = window.location.pathname

// Render different components based on the path
const renderApp = () => {
  // Screenshot editor route
  if (path === '/screenshot-editor') {
    return (
      <ToastProvider>
        <ScreenshotEditor />
      </ToastProvider>
    )
  }

  // Search window route
  if (path === '/search') {
    return <SearchWindow />
  }

  // Default/main app route
  return <App />
}

ReactDOM.createRoot(document.getElementById('root') as HTMLElement).render(
  <React.StrictMode>
    {renderApp()}
  </React.StrictMode>,
)
