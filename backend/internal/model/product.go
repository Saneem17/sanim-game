package model

// product.go — Define struktur data untuk satu produk/item dalam game store.
// Struct ni digunakan untuk map data dari database ke Go struct,
// dan untuk serialize ke JSON bila hantar ke frontend.

import "time"

// Product — represent satu item yang boleh dibeli dalam store
// Tag `json:"..."` tentukan nama field dalam JSON response ke frontend
type Product struct {
	ID          int       `json:"id"`          // ID unik dalam database (auto-increment)
	Name        string    `json:"name"`        // Nama item (contoh: "Wall-nut")
	Description string    `json:"description"` // Penerangan item (flavor text dari PvZ)
	Price       float64   `json:"price"`       // Harga dalam cent/unit (contoh: 5 = USD 5)
	Currency    string    `json:"currency"`    // Matawang (contoh: "USD")
	SKU         string    `json:"sku"`         // Stock Keeping Unit — ID unik item untuk Xsolla (contoh: "pvz-wall-nut")
	ImageURL    string    `json:"image_url"`   // URL gambar item untuk display kat frontend
	CreatedAt   time.Time `json:"created_at"`  // Bila item ni ditambah ke database
}
