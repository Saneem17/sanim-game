package middleware

// auth.go — Middleware untuk protect routes yang perlukan login.
// Sebelum request sampai ke handler, middleware ni akan:
//   1. Check ada Authorization header ke tak
//   2. Extract Bearer token dari header tu
//   3. Verify token dengan Xsolla (check signature, expiry, issuer)
//   4. Kalau valid, allow request proceed — kalau tak, return 401 Unauthorized

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"sanim-backend/internal/service"
)

// contextKey — custom type untuk kunci dalam request context
// Guna custom type untuk elak collision dengan package lain yang guna string key
type contextKey string

// UserClaimsKey — kunci untuk simpan/ambil user claims dalam request context
// Lepas verify token, kita store claims dalam context supaya handler boleh access
const UserClaimsKey contextKey = "userClaims"

// AuthMiddleware — middleware yang verify JWT token Xsolla
type AuthMiddleware struct {
	authService *service.XsollaAuthService // Service yang buat kerja verify token
}

// NewAuthMiddleware — buat AuthMiddleware dengan auth service
func NewAuthMiddleware(authService *service.XsollaAuthService) *AuthMiddleware {
	return &AuthMiddleware{authService: authService}
}

// RequireAuth — wrapper function yang wrap handler dengan auth check
// Guna macam ni: mux.Handle("/api/items", authMiddleware.RequireAuth(itemsHandler))
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Step 1: Check ada Authorization header dalam request
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeUnauthorized(w, "missing Authorization header")
			return
		}

		// Step 2: Format mesti "Bearer <token>" — split dan validate
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" || parts[1] == "" {
			writeUnauthorized(w, "invalid Authorization header format")
			return
		}

		tokenString := parts[1] // Ambil bahagian token je (tanpa "Bearer ")

		// Step 3: Verify token dengan Xsolla — check signature, expiry, issuer, project ID
		claims, err := m.authService.VerifyToken(r.Context(), tokenString)
		if err != nil {
			writeUnauthorized(w, err.Error())
			return
		}

		// Step 4: Token valid! Simpan user claims dalam request context
		// Handler boleh access claims ni guna: r.Context().Value(UserClaimsKey)
		ctx := context.WithValue(r.Context(), UserClaimsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx)) // Proceed ke handler dengan context yang updated
	})
}

// writeUnauthorized — helper untuk hantar response 401 dengan JSON error message
func writeUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized) // 401 Unauthorized

	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": message,
	})
}
