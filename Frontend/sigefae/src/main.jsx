import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import App from './App.jsx'

if (import.meta.env.PROD && 'serviceWorker' in navigator) {
  window.addEventListener('load', () => {
    navigator.serviceWorker
      .register(`${import.meta.env.BASE_URL}sw.js`, {
        scope: import.meta.env.BASE_URL,
      })
      .catch((error) => {
        console.error('No se pudo registrar el Service Worker:', error)
      })
  })
}

createRoot(document.getElementById('root')).render(
  <StrictMode>
    <BrowserRouter basename="/sigefae/Frontend">
      <App />
    </BrowserRouter>
  </StrictMode>,
)