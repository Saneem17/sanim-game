package middleware

// cors.go — Middleware untuk handle Cross-Origin Resource Sharing (CORS).
// CORS ni penting sebab browser akan block request dari domain lain secara default.
// Contoh: frontend kita kat http://localhost:5173 tapi API kat http://localhost:8080
// — tanpa CORS headers, browser akan reject request tu.

import "net/http"

// CORSMiddleware — middleware yang add CORS headers ke setiap response
type CORSMiddleware struct {
	allowedOrigin string // URL frontend yang dibenarkan (contoh: "http://localhost:5173")
}

// NewCORSMiddleware — buat CORSMiddleware dengan origin frontend yang dibenarkan
func NewCORSMiddleware(allowedOrigin string) *CORSMiddleware {
	return &CORSMiddleware{allowedOrigin: allowedOrigin}
}

// Handle — wrap handler dengan CORS headers
func (m *CORSMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow request dari frontend URL kita je (bukan semua origin)
		w.Header().Set("Access-Control-Allow-Origin", m.allowedOrigin)

		// Allow header-header ni dalam request dari frontend
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Allow method-method HTTP ni
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")

		// Bila browser buat "preflight" request (method OPTIONS) sebelum actual request,
		// kita terus balas OK — ini normal behaviour browser untuk check CORS
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Proceed ke handler seterusnya
		next.ServeHTTP(w, r)
	})
}
