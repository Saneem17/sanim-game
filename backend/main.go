package main

// main.go — Ini titik mula (entry point) untuk backend server kita.
// Semua benda start dari sini — load config, connect database, setup handler, then run server.

import (
	"context"
	"log"
	"net/http"
	"time"

	"sanim-backend/internal/config"
	"sanim-backend/internal/handler"
	"sanim-backend/internal/middleware"
	"sanim-backend/internal/repository"
	"sanim-backend/internal/router"
	"sanim-backend/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// Step 1: Load semua setting dari file .env (port, database URL, Xsolla keys, etc.)
	config.LoadEnv(".env")
	cfg := config.Load()

	// Kalau Xsolla config tak lengkap, kita stop terus — tak boleh proceed tanpa ni
	if cfg.XsollaMerchantID == "" || cfg.XsollaAPIKey == "" || cfg.XsollaProjectID == "" {
		log.Fatal("xsolla config missing (MERCHANT_ID / API_KEY / PROJECT_ID)")
	}
	if cfg.XsollaWebhookSecret == "" {
		log.Fatal("XSOLLA_WEBHOOK_SECRET is required")
	}

	// Step 2: Connect ke PostgreSQL database menggunakan connection pool (pgxpool)
	// Connection pool bermakna kita boleh handle banyak request serentak tanpa buat connection baru tiap kali
	dbPool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatal("failed to connect database:", err)
	}
	defer dbPool.Close() // Tutup connection pool bila server shutdown

	// Ping database untuk pastikan connection betul-betul berjaya (timeout 5 saat)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := dbPool.Ping(ctx); err != nil {
		log.Fatal("database ping failed:", err)
	}

	log.Println("database connected")

	// Step 3: Setup Repository — layer yang buat SQL queries ke database
	productRepo := repository.NewProductRepository(dbPool)
	purchaseRepo := repository.NewPurchaseRepository(dbPool)

	// Step 4: Setup Services — layer business logic antara handler dan repository
	productService := service.NewProductService(productRepo)
	// xsollaAuthService untuk verify token via Xsolla /users/me API (tanpa JWKS)
	xsollaAuthService := service.NewXsollaAuthService(cfg.XsollaLoginProjectID)

	// Step 5: Setup Handlers — yang receive HTTP request dan bagi response
	healthHandler := handler.NewHealthHandler()
	productHandler := handler.NewProductHandler(productService)

	// authHandler handle callback bila user login via Xsolla OAuth
	authHandler := handler.NewAuthHandler(
		cfg.XsollaOAuthClientID,
		cfg.XsollaRedirectURI,
	)

	// paymentHandler handle bila user nak buat bayaran (beli item dalam game)
	paymentHandler := handler.NewPaymentHandler(
		cfg.XsollaProjectID,
		purchaseRepo,
	)

	// Step 6: Setup Middleware — code yang jalan SEBELUM handler
	authMiddleware := middleware.NewAuthMiddleware(xsollaAuthService)       // Check token valid ke tak
	corsMiddleware := middleware.NewCORSMiddleware(cfg.FrontendURL)         // Allow frontend request masuk
	loggingMiddleware := middleware.NewLoggingMiddleware()                  // Log setiap request

	// Step 7: Setup Router — tentukan URL mana pergi ke handler mana
	appRouter := router.SetupRouter(
		healthHandler,
		productHandler,
		authHandler,
		paymentHandler,
		authMiddleware,
		corsMiddleware,
		loggingMiddleware,
		purchaseRepo,
		cfg.XsollaWebhookSecret,
	)

	// Step 8: Start HTTP server
	server := &http.Server{
		Addr:         ":" + cfg.AppPort,   // Port yang server akan listen (default 8080)
		Handler:      appRouter,
		ReadTimeout:  10 * time.Second,    // Maksimum masa untuk baca request
		WriteTimeout: 10 * time.Second,    // Maksimum masa untuk hantar response
	}

	log.Printf("server running at http://localhost:%s", cfg.AppPort)
	log.Fatal(server.ListenAndServe()) // Start server — kalau ada error, log dan exit
}
