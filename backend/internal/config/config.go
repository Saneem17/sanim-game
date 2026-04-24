package config

// config.go — Semua setting app kita simpan kat sini.
// Kita baca nilai-nilai ni dari file .env supaya tak payah hardcode dalam code.

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config struct — satu tempat untuk simpan semua configuration yang app kita perlukan
type Config struct {
	AppPort              string // Port yang server run (contoh: 8080)
	FrontendURL          string // URL frontend React/Vite (untuk CORS)
	DatabaseURL          string // Connection string untuk PostgreSQL
	XsollaLoginProjectID string // ID project login Xsolla
	XsollaOAuthClientID  string // Client ID untuk OAuth Xsolla
	XsollaRedirectURI    string // URL yang Xsolla akan redirect lepas login
	XsollaMerchantID      string // ID merchant dalam Xsolla dashboard
	XsollaAPIKey          string // API key untuk panggil Xsolla API
	XsollaProjectID       string // ID project dalam Xsolla store
	XsollaWebhookSecret   string // Secret key untuk verify signature webhook Xsolla
}

// LoadEnv — baca file .env dan load nilai-nilainya ke dalam environment variables
// Kalau file .env tak jumpa, kita guna system environment variables je (okay untuk production)
func LoadEnv(path string) {
	err := godotenv.Load(path)
	if err != nil {
		log.Println(".env file not found, using system environment variables")
	}
}

// Load — ambil semua environment variables dan buat satu Config struct
// Beberapa nilai ada default (contoh AppPort default ke "8080") kalau tak set dalam .env
func Load() Config {
	cfg := Config{
		AppPort:              getEnv("APP_PORT", "8080"),                    // Default port 8080
		FrontendURL:          getEnv("FRONTEND_URL", "http://localhost:5173"), // Default Vite dev server
		DatabaseURL:          os.Getenv("DATABASE_URL"),
		XsollaLoginProjectID: os.Getenv("XSOLLA_LOGIN_PROJECT_ID"),
		XsollaOAuthClientID:  os.Getenv("XSOLLA_OAUTH_CLIENT_ID"),
		XsollaRedirectURI:    os.Getenv("XSOLLA_REDIRECT_URI"),
		XsollaMerchantID:    os.Getenv("XSOLLA_MERCHANT_ID"),
		XsollaAPIKey:        os.Getenv("XSOLLA_API_KEY"),
		XsollaProjectID:     os.Getenv("XSOLLA_PROJECT_ID"),
		XsollaWebhookSecret: os.Getenv("XSOLLA_WEBHOOK_SECRET"),
	}

	// DATABASE_URL wajib ada — kalau takde, app tak boleh jalan langsung
	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	return cfg
}

// getEnv — helper function: ambil nilai environment variable,
// kalau kosong guna fallback value yang dibagi
func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
