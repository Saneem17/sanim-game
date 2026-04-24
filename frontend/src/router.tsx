// router.tsx — Definisi semua URL routes dalam app.
// Setiap path dipetakan ke page component yang berkaitan.
//
// Flow navigation:
//   '/'              → LoginPage       — halaman utama, user kena login dulu
//   '/auth/callback' → AuthCallbackPage — Xsolla redirect ke sini selepas login berjaya
//   '/store'         → StorePage       — game store, hanya boleh akses selepas login

import { createBrowserRouter } from 'react-router-dom'
import LoginPage from './pages/LoginPage'
import AuthCallbackPage from './pages/AuthCallbackPage'
import StorePage from './pages/StorePage'

export const router = createBrowserRouter([
  { path: '/', element: <LoginPage /> },                    // Halaman login
  { path: '/auth/callback', element: <AuthCallbackPage /> }, // OAuth callback dari Xsolla
  { path: '/store', element: <StorePage /> },               // Game store (selepas login)
])
