package handler

// webhook.go — Handle payment webhook dari Xsolla.
//
// Flow webhook:
//   1. User bayar dalam Xsolla Pay Station
//   2. Xsolla hantar POST request ke /webhook/xsolla (server kita)
//   3. Kita verify signature untuk pastikan request betul-betul dari Xsolla
//   4. Kita extract user ID dan item SKU dari payload
//   5. Kita simpan rekod purchase dalam database
//   6. Return {"status":"ok"} ke Xsolla
//
// Kenapa guna signature dan bukan JWT?
//   Xsolla webhook tidak bawa JWT — dia guna SHA1 signature untuk authentication.
//   Format: Authorization: Signature <sha1(body + secret_key)>

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"sanim-backend/internal/repository"
)

// WebhookHandler — handler untuk Xsolla webhook notifications
type WebhookHandler struct {
	purchaseRepo  *repository.PurchaseRepository // Untuk simpan purchase ke DB
	webhookSecret string                          // Secret key dari Xsolla dashboard
}

// NewWebhookHandler — buat WebhookHandler dengan purchase repository dan webhook secret
func NewWebhookHandler(purchaseRepo *repository.PurchaseRepository, webhookSecret string) *WebhookHandler {
	return &WebhookHandler{purchaseRepo: purchaseRepo, webhookSecret: webhookSecret}
}

// verifyXsollaSignature — sahkan bahawa request datang dari Xsolla, bukan orang lain.
// Xsolla sign setiap webhook dengan: SHA1(body + secret_key)
// Nilai ni dihantar dalam Authorization header dengan format: "Signature <hex_value>"
func verifyXsollaSignature(body []byte, authHeader, secret string) bool {
	// Split "Signature <hex_value>" jadi dua bahagian
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Signature" {
		return false
	}

	// Kira SHA1(body + secret) dan bandingkan dengan nilai dalam header
	h := sha1.New()
	h.Write(body)
	h.Write([]byte(secret))
	return hex.EncodeToString(h.Sum(nil)) == parts[1]
}

// HandleWebhook — handle POST /webhook/xsolla
func (h *WebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Baca body request sepenuhnya — perlu untuk verify signature
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// SECURITY: Verify signature dulu sebelum proses apa-apa
	// Kalau signature tak match, reject request terus
	if !verifyXsollaSignature(body, r.Header.Get("Authorization"), h.webhookSecret) {
		log.Println("webhook: invalid signature, rejecting request")
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}

	log.Println("Xsolla webhook received")

	// Parse JSON payload dari Xsolla
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "invalid json payload", http.StatusBadRequest)
		return
	}

	// Semak jenis notification — kita hanya proses "payment" dan "order_paid"
	notificationType, _ := payload["notification_type"].(string)
	log.Println("notification_type:", notificationType)

	if notificationType == "payment" || notificationType == "order_paid" {
		// Extract user ID dan item SKU dari payload Xsolla
		userID := extractUserID(payload)
		itemSKU := extractItemSKU(payload)

		log.Println("user_id:", userID)
		log.Println("item_sku:", itemSKU)

		// Simpan purchase ke database kalau kedua-dua nilai ada
		// Repository: internal/repository/purchase.go → Save()
		if userID != "" && itemSKU != "" {
			if err := h.purchaseRepo.Save(context.Background(), userID, itemSKU); err != nil {
				log.Println("failed to save purchase:", err)
				http.Error(w, "failed to save purchase", http.StatusInternalServerError)
				return
			}
			log.Println("purchase saved successfully")
		}
	}

	// Xsolla expect response 200 dengan JSON {"status":"ok"} untuk confirm webhook diterima
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

// extractUserID — ambil user ID dari payload webhook Xsolla.
// Xsolla letak user info dalam field "user" → "id".
// Fallback ke "external_id" kalau format berbeza.
func extractUserID(payload map[string]any) string {
	if user, ok := payload["user"].(map[string]any); ok {
		if id, ok := user["id"].(string); ok {
			return id
		}
	}
	if external, ok := payload["external_id"].(string); ok {
		return external
	}
	return ""
}

// extractItemSKU — ambil SKU item dari payload webhook Xsolla.
// Cuba dua tempat: "custom_parameters.item_sku" atau "purchase.items[0].sku"
func extractItemSKU(payload map[string]any) string {
	// Cuba custom_parameters dulu (kalau kita set masa create payment)
	if custom, ok := payload["custom_parameters"].(map[string]any); ok {
		if sku, ok := custom["item_sku"].(string); ok {
			return sku
		}
	}
	// Cuba dalam purchase.items array (standard Xsolla format)
	if purchase, ok := payload["purchase"].(map[string]any); ok {
		if items, ok := purchase["items"].([]any); ok && len(items) > 0 {
			if firstItem, ok := items[0].(map[string]any); ok {
				if sku, ok := firstItem["sku"].(string); ok {
					return sku
				}
			}
		}
	}
	return ""
}
