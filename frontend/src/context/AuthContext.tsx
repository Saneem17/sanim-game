// AuthContext.tsx — Global state management untuk authentication.
//
// Menggunakan React Context supaya token boleh diakses dari mana-mana komponen
// tanpa perlu pass props dari parent ke child (prop drilling).
//
// Cara guna dalam komponen lain:
//   const { isAuthenticated, login, logout } = useAuthContext()
//
// Flow:
//   1. App mula → AuthProvider baca token dari sessionStorage
//   2. User login → login() dipanggil → token disimpan
//   3. User logout → logout() dipanggil → token dibuang
//   4. Mana-mana komponen boleh semak isAuthenticated untuk tahu status login

import { createContext, useContext, useState } from 'react'
import type { ReactNode } from 'react'
import { getToken, saveToken, removeToken } from '../utils/storage'

// Interface yang define apa yang AuthContext sediakan ke komponen lain
interface AuthContextType {
  token: string | null            // Token JWT semasa (null kalau belum login)
  isAuthenticated: boolean        // true kalau ada token, false kalau takde
  login: (token: string) => void  // Simpan token baru (panggil selepas dapat token dari Xsolla)
  logout: () => void              // Buang token (clear semua auth state)
}

// Buat context dengan null sebagai nilai default
// null bermakna context belum di-provide — guna untuk detect kalau guna luar Provider
const AuthContext = createContext<AuthContextType | null>(null)

// AuthProvider — bungkus app dengan context ni supaya semua child dapat akses token
export function AuthProvider({ children }: { children: ReactNode }) {
  // Ambil token dari sessionStorage semasa app mula — supaya login kekal bila refresh page
  const [token, setToken] = useState<string | null>(getToken())

  // login — simpan token dalam sessionStorage DAN update React state
  const login = (newToken: string) => {
    saveToken(newToken)  // Simpan dalam sessionStorage (persist merentasi refresh)
    setToken(newToken)   // Update state (trigger re-render semua komponen yang guna context ni)
  }

  // logout — buang token dari sessionStorage DAN clear React state
  const logout = () => {
    removeToken()  // Buang dari sessionStorage
    setToken(null) // Clear state (trigger re-render → redirect ke login)
  }

  return (
    <AuthContext.Provider value={{ token, isAuthenticated: !!token, login, logout }}>
      {children}
    </AuthContext.Provider>
  )
}

// useAuthContext — custom hook untuk guna context ini dengan mudah dalam mana-mana komponen
// Throw error kalau dipanggil di luar AuthProvider — untuk tangkap bug awal
export function useAuthContext() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('AuthContext missing')
  return ctx
}
