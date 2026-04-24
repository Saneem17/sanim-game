package handler

// product.go — Handle request untuk ambil senarai item/produk dari database.
// GET /items atau GET /api/items (protected) — return semua item dalam store.

import (
	"encoding/json"
	"net/http"

	"sanim-backend/internal/service"
)

// ProductHandler — handler untuk product-related endpoints
// Bergantung pada ProductService untuk ambil data dari database
type ProductHandler struct {
	productService *service.ProductService
}

// NewProductHandler — buat ProductHandler dengan service yang diperlukan
func NewProductHandler(productService *service.ProductService) *ProductHandler {
	return &ProductHandler{productService: productService}
}

// GetItems — handle GET /items
// Ambil semua produk dari database dan hantar balik sebagai JSON array
func (h *ProductHandler) GetItems(w http.ResponseWriter, r *http.Request) {
	// Panggil service untuk ambil semua produk dari DB
	products, err := h.productService.GetAllProducts(r.Context())
	if err != nil {
		http.Error(w, "failed to fetch products", http.StatusInternalServerError)
		return
	}

	// Hantar list produk dalam format JSON ke frontend
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(products)
}
