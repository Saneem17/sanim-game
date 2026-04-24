package handler

// payment.go — Handle proses pembayaran menggunakan Xsolla Store API.
// Flow pembayaran:
//   1. Frontend hantar list item yang nak dibeli + user token
//   2. Kita clear cart lama user dalam Xsolla
//   3. Kita add semua item baru ke cart Xsolla
//   4. Kita checkout cart → Xsolla bagi payment token
//   5. Frontend guna token tu untuk buka Xsolla payment page
//   6. Kita terus simpan purchase dalam DB (tak tunggu webhook sebab local dev)

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"sanim-backend/internal/repository"
)

// LimitedItemSKU — SKU untuk item terhad (dancing-zombie) yang hanya boleh dibeli sekali
const LimitedItemSKU = "dancing-zombie"

// PaymentHandler — handler untuk proses payment
type PaymentHandler struct {
	projectID    string
	purchaseRepo *repository.PurchaseRepository
}

// NewPaymentHandler — buat PaymentHandler dengan project ID dan purchase repository
func NewPaymentHandler(projectID string, purchaseRepo *repository.PurchaseRepository) *PaymentHandler {
	return &PaymentHandler{
		projectID:    projectID,
		purchaseRepo: purchaseRepo,
	}
}

// CartItemRequest — satu item dalam cart yang frontend hantar
type CartItemRequest struct {
	SKU      string `json:"sku"`      // ID unik item (contoh: "pvz-wall-nut")
	Quantity int    `json:"quantity"` // Berapa banyak nak beli
}

// CreatePaymentRequest — data yang frontend hantar untuk mulakan payment
type CreatePaymentRequest struct {
	Items     []CartItemRequest `json:"items"`      // List item yang nak dibeli
	Currency  string            `json:"currency"`   // Matawang (default: USD)
	UserToken string            `json:"user_token"` // Xsolla user token untuk auth ke Xsolla Store API
	UserID    string            `json:"user_id"`    // ID user kita (untuk simpan rekod purchase)
}

// CreatePaymentResponse — apa yang kita hantar balik ke frontend: payment token untuk buka payment page
type CreatePaymentResponse struct {
	Token string `json:"token"` // Token ni frontend guna untuk redirect ke Xsolla payment page
}

// CreatePayment — main function yang handle POST /create-payment
func (h *PaymentHandler) CreatePayment(w http.ResponseWriter, r *http.Request) {
	// Hanya accept POST request
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Decode JSON body dari frontend
	var req CreatePaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Kena ada sekurang-kurangnya satu item dalam cart
	if len(req.Items) == 0 {
		http.Error(w, "items is required", http.StatusBadRequest)
		return
	}

	// User token wajib ada — kita perlukan untuk panggil Xsolla Store API atas nama user
	if req.UserToken == "" {
		http.Error(w, "user_token is required", http.StatusBadRequest)
		return
	}

	// Default currency USD kalau takde
	if req.Currency == "" {
		req.Currency = "USD"
	}

	// Semak had untuk dancing-zombie — item terhad yang hanya boleh dibeli SEKALI
	for _, item := range req.Items {
		if item.SKU == LimitedItemSKU {
			// Tak boleh beli lebih dari satu unit dalam satu transaksi
			if item.Quantity > 1 {
				http.Error(w, "limited item can only be purchased once", http.StatusForbidden)
				return
			}
			// Semak dalam database — adakah user ni dah pernah beli dancing-zombie sebelum ni?
			if req.UserID != "" {
				already, err := h.purchaseRepo.HasPurchased(r.Context(), req.UserID, item.SKU)
				if err == nil && already {
					http.Error(w, "you already purchased this limited item", http.StatusForbidden)
					return
				}
			}
		}
	}

	// Ambil URL frontend untuk redirect balik lepas payment selesai
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}

	client := &http.Client{}

	// ── STEP 1: Clear cart lama user dalam Xsolla ──
	// Kita nak start fresh — buang item-item lama dalam cart sebelum add yang baru
	clearURL := fmt.Sprintf(
		"https://store.xsolla.com/api/v2/project/%s/cart/clear",
		h.projectID,
	)
	clearReq, err := http.NewRequest(http.MethodPut, clearURL, nil)
	if err != nil {
		http.Error(w, "failed to build clear cart request", http.StatusInternalServerError)
		return
	}
	clearReq.Header.Set("Authorization", "Bearer "+req.UserToken) // Auth sebagai user
	clearResp, err := client.Do(clearReq)
	if err != nil {
		http.Error(w, "failed to clear cart", http.StatusBadGateway)
		return
	}
	clearResp.Body.Close()

	// ── STEP 2: Add setiap item ke cart Xsolla satu persatu ──
	for _, item := range req.Items {
		// Payload untuk add item: just quantity
		itemPayload := map[string]any{
			"quantity": item.Quantity,
		}
		itemBody, _ := json.Marshal(itemPayload)

		// URL format: /cart/item/{sku}
		addURL := fmt.Sprintf(
			"https://store.xsolla.com/api/v2/project/%s/cart/item/%s",
			h.projectID,
			item.SKU,
		)
		addReq, err := http.NewRequest(http.MethodPut, addURL, bytes.NewBuffer(itemBody))
		if err != nil {
			http.Error(w, "failed to build add-to-cart request", http.StatusInternalServerError)
			return
		}
		addReq.Header.Set("Content-Type", "application/json")
		addReq.Header.Set("Authorization", "Bearer "+req.UserToken)

		addResp, err := client.Do(addReq)
		if err != nil {
			http.Error(w, "failed to add item to cart", http.StatusBadGateway)
			return
		}
		addBody, _ := io.ReadAll(addResp.Body)
		addResp.Body.Close()

		// Xsolla return 200 atau 204 bila berjaya
		if addResp.StatusCode != http.StatusOK && addResp.StatusCode != http.StatusNoContent {
			http.Error(w, "failed to add item: "+string(addBody), http.StatusBadGateway)
			return
		}
	}

	// ── STEP 3: Checkout cart → dapat payment token ──
	// Xsolla akan bagi token yang frontend boleh guna untuk buka payment page
	checkoutPayload := map[string]any{
		"currency": req.Currency,
		"sandbox":  true, // Mode sandbox untuk testing — duit tak keluar betul-betul
		"settings": map[string]any{
			"return_url": frontendURL + "/store?payment=success", // Redirect sini lepas bayar
		},
	}

	checkoutBody, err := json.Marshal(checkoutPayload)
	if err != nil {
		http.Error(w, "failed to encode checkout request", http.StatusInternalServerError)
		return
	}

	checkoutURL := fmt.Sprintf(
		"https://store.xsolla.com/api/v2/project/%s/payment/cart",
		h.projectID,
	)

	checkoutReq, err := http.NewRequest(http.MethodPost, checkoutURL, bytes.NewBuffer(checkoutBody))
	if err != nil {
		http.Error(w, "failed to create checkout request", http.StatusInternalServerError)
		return
	}
	checkoutReq.Header.Set("Content-Type", "application/json")
	checkoutReq.Header.Set("Authorization", "Bearer "+req.UserToken)

	checkoutResp, err := client.Do(checkoutReq)
	if err != nil {
		http.Error(w, "failed to contact xsolla checkout", http.StatusBadGateway)
		return
	}
	defer checkoutResp.Body.Close()

	responseBody, err := io.ReadAll(checkoutResp.Body)
	if err != nil {
		http.Error(w, "failed to read xsolla response", http.StatusInternalServerError)
		return
	}

	if checkoutResp.StatusCode != http.StatusOK {
		http.Error(w, "xsolla payment request failed: "+string(responseBody), http.StatusBadGateway)
		return
	}

	// Parse token dari response Xsolla
	var tokenResp CreatePaymentResponse
	if err := json.Unmarshal(responseBody, &tokenResp); err != nil {
		http.Error(w, "failed to parse xsolla payment response", http.StatusInternalServerError)
		return
	}

	if tokenResp.Token == "" {
		http.Error(w, "xsolla token is empty", http.StatusBadGateway)
		return
	}

	// ── STEP 4: Terus simpan purchase dalam DB tanpa tunggu webhook ──
	// Dalam local development, Xsolla tak boleh reach localhost kita untuk hantar webhook.
	// Jadi kita simpan terus selepas payment token berjaya dibuat.
	// Dalam production, webhook akan handle ni secara proper.
	if req.UserID != "" {
		for _, item := range req.Items {
			if err := h.purchaseRepo.Save(r.Context(), req.UserID, item.SKU); err != nil {
				fmt.Printf("warning: failed to save purchase for user %s sku %s: %v\n",
					req.UserID, item.SKU, err)
			}
		}
	}

	// Hantar token balik ke frontend — frontend akan guna ni untuk buka Xsolla payment page
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(tokenResp)
}
