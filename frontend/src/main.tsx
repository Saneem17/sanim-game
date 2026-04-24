// main.tsx — Titik mula (entry point) React app.
// React render bermula dari sini — cari <div id="root"> dalam index.html
// dan pasang keseluruhan app ke dalam element tu.
//
// Susunan bungkus (dari luar ke dalam):
//   StrictMode → AuthProvider → App → Router → Pages

import React from 'react'
import ReactDOM from 'react-dom/client'
import App from './App'
import { AuthProvider } from './context/AuthContext'
import './styles/globals.css'

ReactDOM.createRoot(document.getElementById('root')!).render(
  // StrictMode: dalam development, render komponen 2x untuk detect side effects
  <React.StrictMode>
    {/* AuthProvider: bagi semua page dan komponen dalam app akses kepada token login */}
    <AuthProvider>
      <App />
    </AuthProvider>
  </React.StrictMode>
)
