// LoginPage.tsx — Halaman pertama yang user nampak bila buka app.
//
// Logik mudah:
//   - Kalau user DAH login (ada token dalam sessionStorage) → redirect terus ke /store
//   - Kalau BELUM login → papar butang login Xsolla
//
// Component yang dipapar: LoginWidget (butang "Login" yang buka Xsolla popup)

import { Navigate } from 'react-router-dom'
import { useAuthContext } from '../context/AuthContext'
import LoginWidget from '../components/auth/LoginWidget'
import '../styles/login.css'

export default function LoginPage() {
  // Semak status login dari AuthContext
  const { isAuthenticated } = useAuthContext()

  // Kalau dah login, redirect terus ke store — jangan tunjuk halaman login lagi
  if (isAuthenticated) return <Navigate to="/store" />

  // Kalau belum login, papar halaman dengan login widget Xsolla
  return (
    <div className="login">
      <div className="login__card">
        <h1> Garden War </h1>
        <p>Login to continue</p>
        {/* LoginWidget render butang "Login" yang buka Xsolla popup bila diklik */}
        <LoginWidget />
      </div>
    </div>
  )
}
