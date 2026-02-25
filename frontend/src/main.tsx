import React from 'react'
import ReactDOM from 'react-dom/client'
import App from './App'
import { SearchWindow } from './components/SearchWindow'
import { Screenshot2Overlay } from './components/Screenshot2'
import PinWindow from './components/Screenshot2/PinWindow'
import './styles.css'

// Get the current path from window.location
const path = window.location.pathname

// Render different components based on the path
const renderApp = () => {
  // Screenshot2 overlay route (WeChat-style screenshot)
  if (path === '/screenshot2-overlay') {
    return <Screenshot2Overlay />
  }

  // Pin window route
  if (path === '/pin-window') {
    const params = new URLSearchParams(window.location.search);
    const windowId = parseInt(params.get('id') || '0', 10);
    return <PinWindow windowId={windowId} />
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
