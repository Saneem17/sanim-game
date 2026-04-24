package handler

import (
	"encoding/json"
	"net/http"

	"sanim-backend/internal/middleware"
	"sanim-backend/internal/repository"
	"sanim-backend/internal/service"
)

// PurchaseHandler — handler untuk purchase-related endpoints
type PurchaseHandler struct {
	purchaseRepo *repository.PurchaseRepository
}

// NewPurchaseHandler — buat PurchaseHandler dengan purchase repository
func NewPurchaseHandler(purchaseRepo *repository.PurchaseRepository) *PurchaseHandler {
	return &PurchaseHandler{purchaseRepo: purchaseRepo}
}

// PurchasedItemsResponse — format response yang kita hantar balik ke frontend
type PurchasedItemsResponse struct {
	PurchasedSKUs []string `json:"purchased_skus"`
}

// GetPurchases — handle GET /purchases
// User ID diambil dari JWT claims dalam context supaya user tidak boleh query rekod orang lain
func (h *PurchaseHandler) GetPurchases(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims, ok := r.Context().Value(middleware.UserClaimsKey).(*service.XsollaClaims)
	if !ok || claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	skus, err := h.purchaseRepo.GetByUser(r.Context(), claims.Subject)
	if err != nil {
		http.Error(w, "failed to query purchases", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(PurchasedItemsResponse{PurchasedSKUs: skus})
}
