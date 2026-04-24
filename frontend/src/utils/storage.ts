// storage.ts — Utility untuk simpan dan ambil access token dari sessionStorage.
//
// Kenapa sessionStorage dan bukan localStorage?
//   - sessionStorage hilang bila tab browser ditutup → lebih selamat
//   - localStorage kekal selama-lamanya → token lama boleh disalahguna
//
// Nota: Hanya token utama (xsolla_token) diuruskan di sini.
// Data lain seperti user_email, xsolla_user_id dll disimpan terus
// dalam sessionStorage oleh AuthCallbackPage semasa proses login.

const TOKEN_KEY = 'xsolla_token'

// saveToken — simpan JWT access token dalam sessionStorage
export function saveToken(token: string) {
  sessionStorage.setItem(TOKEN_KEY, token)
}

// getToken — ambil token yang tersimpan (return null kalau belum login atau tab baru)
export function getToken() {
  return sessionStorage.getItem(TOKEN_KEY)
}

// removeToken — buang token semasa logout
export function removeToken() {
  sessionStorage.removeItem(TOKEN_KEY)
}
