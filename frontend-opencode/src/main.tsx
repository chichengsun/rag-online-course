import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import App from './App.tsx'

// 路由由 App 内 RouterProvider（createBrowserRouter）统一提供，此处不可再包 BrowserRouter，否则会嵌套两个 Router 导致白屏。
createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>,
)
