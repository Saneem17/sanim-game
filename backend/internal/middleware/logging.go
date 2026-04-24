package middleware

// logging.go — Middleware untuk log setiap HTTP request yang masuk ke server.
// Bila ada request, kita log: method, URL path, dan berapa lama masa nak proses.
// Berguna untuk debugging dan monitor performance server.

import (
	"log"
	"net/http"
	"time"
)

// LoggingMiddleware — middleware untuk log requests (takde field sebab takde config)
type LoggingMiddleware struct{}

// NewLoggingMiddleware — buat LoggingMiddleware baru
func NewLoggingMiddleware() *LoggingMiddleware {
	return &LoggingMiddleware{}
}

// Handle — wrap handler dengan logging
func (m *LoggingMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Catat masa mula request diterima
		start := time.Now()

		// Proses request (pergi ke handler seterusnya)
		next.ServeHTTP(w, r)

		// Lepas handler selesai, log maklumat request:
		// Contoh output: "GET /items 1.234ms"
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}
