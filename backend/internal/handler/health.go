package handler

// health.go — Endpoint mudah untuk check sama ada server still hidup ke tak.
// GET /health akan bagi response {"message": "backend is running"}
// Berguna untuk monitoring tools dan deployment checks.

import (
	"encoding/json"
	"net/http"
)

// HealthHandler — handler untuk endpoint /health (takde field sebab takde dependency)
type HealthHandler struct{}

// NewHealthHandler — buat HealthHandler baru
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// ServeHTTP — bila ada request ke /health, terus balas dengan status OK
// Implement interface http.Handler supaya boleh guna dengan mux.Handle()
func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "backend is running",
	})
}
