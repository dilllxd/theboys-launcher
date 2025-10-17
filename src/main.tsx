import React from 'react'
import ReactDOM from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import { Toaster } from 'react-hot-toast'
import App from './App.tsx'
import './styles/index.css'

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <BrowserRouter>
      <App />
      <Toaster
        position="top-right"
        toastOptions={{
          duration: 4000,
          style: {
            background: '#2d3748',
            color: '#ffffff',
            border: '1px solid #4a5568',
          },
          success: {
            style: {
              background: '#38a169',
              border: '1px solid #48bb78',
            },
          },
          error: {
            style: {
              background: '#e53e3e',
              border: '1px solid #f56565',
            },
          },
        }}
      />
    </BrowserRouter>
  </React.StrictMode>,
)