// LoginWidget.tsx — Komponen yang render Xsolla Login Widget.
//
// Flow login penuh:
//   1. User klik butang "Login" → handleOpenWidget() dipanggil
//   2. Widget Xsolla dimount ke dalam div#xl_auth dan dibuka (popup)
//   3. User log masuk dengan email / Google / Steam dalam widget tu
//   4. Xsolla redirect browser ke VITE_XSOLLA_CALLBACK_URL (/auth/callback)
//      dengan "code" dan "state" dalam URL
//   5. AuthCallbackPage ambil code tu dan tukar dengan backend untuk dapat token

import { useRef } from 'react'
import { Widget } from '@xsolla/login-sdk'

export default function LoginWidget() {
  // widgetRef — simpan instance Widget supaya boleh re-open tanpa buat widget baru lagi
  const widgetRef = useRef<Widget | null>(null)
  // mountedRef — track sama ada widget dah di-mount ke DOM (elak mount dua kali)
  const mountedRef = useRef(false)

  const handleOpenWidget = () => {
    // Generate random UUID untuk CSRF protection
    // Xsolla akan return nilai "state" ni balik dalam callback URL
    const state = crypto.randomUUID()
    sessionStorage.setItem('xsolla_state', state)

    // Cari container div dalam DOM dan pastikan ia visible
    const container = document.getElementById('xl_auth')
    if (!container) return
    container.style.display = 'block'

    // Mount widget sekali je — kalau dah mount sebelum ni, just open semula
    if (!mountedRef.current) {
      widgetRef.current = new Widget({
        projectId: import.meta.env.VITE_XSOLLA_LOGIN_PROJECT_ID,        // ID project login Xsolla
        preferredLocale: import.meta.env.VITE_XSOLLA_LOCALE || 'en_US', // Bahasa widget (en_US, ms_MY, dll)
        clientId: import.meta.env.VITE_XSOLLA_CLIENT_ID,                // OAuth 2.0 client ID
        responseType: 'code',  // Minta authorization code (bukan token terus — lebih selamat)
        state,                 // CSRF token untuk verify callback nanti
        redirectUri: import.meta.env.VITE_XSOLLA_CALLBACK_URL,          // URL callback selepas login
        scope: import.meta.env.VITE_XSOLLA_SCOPE || 'offline email',    // Permissions: email + refresh token
      } as never)

      widgetRef.current.mount('xl_auth') // Pasang widget ke dalam div#xl_auth
      mountedRef.current = true
    }

    // Buka widget popup (mungkin dah mount sebelum ni, just buka semula)
    widgetRef.current?.open()
  }

  return (
    <div className="login-widget">
      {/* Container untuk Xsolla widget — SDK akan inject HTML widget ke dalam div ni */}
      <div id="xl_auth" className="login-widget__container"></div>

      {/* Butang yang user klik untuk buka popup login */}
      <button
        type="button"
        className="login-widget__button"
        onClick={handleOpenWidget}
      >
        Login
      </button>
    </div>
  )
}
