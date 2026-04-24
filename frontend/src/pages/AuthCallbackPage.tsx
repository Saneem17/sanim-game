// AuthCallbackPage.tsx — Halaman yang dipapar selepas Xsolla redirect balik ke app kita.
//
// Flow penuh (sambungan dari LoginWidget.tsx):
//   1. Xsolla redirect ke /auth/callback?code=xxxx&state=yyyy
//   2. Halaman ni ambil "code" dari URL
//   3. Hantar code ke backend POST /auth/xsolla/callback
//   4. Backend tukar code dengan Xsolla untuk dapat access_token
//   5. Backend decode JWT dan return maklumat user (email, name, user_id, dll)
//   6. Kita simpan semua maklumat tu dalam sessionStorage
//   7. Redirect ke /store
//
// File backend yang handle request ni: internal/handler/auth.go → HandleXsollaCallback()

import { useEffect } from 'react'
import { useNavigate } from 'react-router-dom'

// Shape data yang kita terima dari backend selepas login berjaya
type LoginResponse = {
  email?: string
  name?: string
  access_token?: string
  refresh_token?: string
  provider?: string
  user_id?: string
  picture?: string
}

// getUserIdFromToken — cuba extract user ID dari JWT payload sebagai fallback
// Digunakan kalau backend tak return user_id dalam response
function getUserIdFromToken(token: string): string {
  try {
    const parts = token.split('.')
    if (parts.length < 2) return ''

    // Bahagian kedua JWT (sebelum titik ketiga) adalah payload dalam base64
    const payload = JSON.parse(atob(parts[1]))

    // Cuba beberapa field yang mungkin ada user ID dalam JWT Xsolla
    return (
      payload.sub ||
      payload.user_id ||
      payload.external_account_id ||
      payload.id ||
      ''
    )
  } catch (error) {
    console.error('Failed to decode token payload:', error)
    return ''
  }
}

export default function AuthCallbackPage() {
  const navigate = useNavigate()

  useEffect(() => {
    // Ambil "code" dari query string URL (?code=xxxx)
    const params = new URLSearchParams(window.location.search)
    const code = params.get('code')

    const loginWithXsolla = async () => {
      // Kalau takde code dalam URL, ada masalah — redirect balik ke login
      if (!code) {
        navigate('/', { replace: true })
        return
      }

      try {
        // Hantar code ke backend untuk ditukar dengan token
        // Backend: internal/handler/auth.go → HandleXsollaCallback()
        const res = await fetch(`${import.meta.env.VITE_API_BASE_URL}/auth/xsolla/callback`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({ code }),
        })

        if (!res.ok) {
          const text = await res.text()
          console.error('Backend response:', text)
          throw new Error('Failed to process login')
        }

        const data: LoginResponse = await res.json()

        // Tentukan user ID — guna dari response dulu, fallback ke decode JWT
        const accessToken = data.access_token || ''
        const fallbackUserId = accessToken ? getUserIdFromToken(accessToken) : ''
        const finalUserId = data.user_id || fallbackUserId || ''

        // Simpan semua maklumat user dalam sessionStorage untuk guna dalam StorePage
        sessionStorage.setItem('xsolla_code', code)
        sessionStorage.setItem('user_email', data.email || 'unknown@gardenwars.com')
        sessionStorage.setItem('user_name', data.name || data.email || 'Garden Player')
        sessionStorage.setItem('xsolla_access_token', accessToken)   // Token utama untuk API calls
        sessionStorage.setItem('xsolla_refresh_token', data.refresh_token || '')
        sessionStorage.setItem('xsolla_provider', data.provider || '')
        sessionStorage.setItem('xsolla_user_id', finalUserId)         // ID untuk track purchases
        sessionStorage.setItem('xsolla_picture', data.picture || '')

        // Login berjaya — pergi ke store
        navigate('/store', { replace: true })
      } catch (error) {
        console.error('Error:', error)
        navigate('/', { replace: true }) // Kalau gagal, balik ke login
      }
    }

    loginWithXsolla()
  }, [navigate])

  // Papar mesej ringkas semasa proses login berlaku (background)
  return <div>Processing login...</div>
}
