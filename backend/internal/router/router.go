package router

import (
	"net/http"

	"sanim-backend/internal/handler"
	"sanim-backend/internal/middleware"
	"sanim-backend/internal/repository"
)

// SetupRouter — setup semua routes dan return satu http.Handler yang siap guna
func SetupRouter(
	healthHandler *handler.HealthHandler,
	productHandler *handler.ProductHandler,
	authHandler *handler.AuthHandler,
	paymentHandler *handler.PaymentHandler,
	authMiddleware *middleware.AuthMiddleware,
	corsMiddleware *middleware.CORSMiddleware,
	loggingMiddleware *middleware.LoggingMiddleware,
	purchaseRepo *repository.PurchaseRepository,
	webhookSecret string,
) http.Handler {
	mux := http.NewServeMux()

	// GET /health — public, takde auth
	mux.Handle("/health", healthHandler)

	// GET /items — public, sesiapa boleh tengok item dalam store
	mux.HandleFunc("/items", productHandler.GetItems)

	// POST /auth/xsolla/callback — tukar OAuth code jadi access token (public by design)
	mux.HandleFunc("/auth/xsolla/callback", authHandler.HandleXsollaCallback)

	// POST /webhook/xsolla — Xsolla call ni bila payment berjaya (protected by signature, bukan JWT)
	webhookHandler := handler.NewWebhookHandler(purchaseRepo, webhookSecret)
	mux.HandleFunc("/webhook/xsolla", webhookHandler.HandleWebhook)

	// POST /create-payment — buat payment session (kena login)
	mux.Handle("/create-payment", authMiddleware.RequireAuth(http.HandlerFunc(paymentHandler.CreatePayment)))

	// GET /purchases — ambil senarai item yang user dah beli (kena login)
	purchaseHandler := handler.NewPurchaseHandler(purchaseRepo)
	mux.Handle("/purchases", authMiddleware.RequireAuth(http.HandlerFunc(purchaseHandler.GetPurchases)))

	// GET /api/items — sama macam /items tapi protected
	mux.Handle("/api/items", authMiddleware.RequireAuth(http.HandlerFunc(productHandler.GetItems)))

	var finalHandler http.Handler = mux
	finalHandler = corsMiddleware.Handle(finalHandler)
	finalHandler = loggingMiddleware.Handle(finalHandler)

	return finalHandler
}
