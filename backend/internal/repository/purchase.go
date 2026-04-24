package repository

// purchase.go — Repository layer untuk semua database operations berkaitan purchases.
//
// Layer ini bertanggungjawab HANYA untuk baca/tulis data ke PostgreSQL.
// Dia tak tahu pasal HTTP, JWT, atau business logic — kerja dia buat SQL je.
//
// Siapa yang panggil repository ni:
//   - handler/payment.go  → Save()        — simpan purchase terus selepas buat payment token
//   - handler/webhook.go  → Save()        — simpan purchase bila Xsolla hantar webhook
//   - handler/purchase.go → GetByUser()   — ambil senarai SKU untuk inventory user
//   - handler/payment.go  → HasPurchased() — check limited item sebelum bagi beli lagi

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PurchaseRepository — simpan database connection untuk buat queries berkaitan purchase
type PurchaseRepository struct {
	db *pgxpool.Pool // Connection pool ke PostgreSQL
}

// NewPurchaseRepository — buat PurchaseRepository dengan database pool
func NewPurchaseRepository(db *pgxpool.Pool) *PurchaseRepository {
	return &PurchaseRepository{db: db}
}

// Save — simpan rekod purchase baru (user beli item).
// ON CONFLICT DO NOTHING: kalau rekod yang sama dah ada, skip je — tak bagi error.
// Ini penting untuk idempotency — kalau webhook hantar dua kali, rekod tak berganda.
func (r *PurchaseRepository) Save(ctx context.Context, userID, itemSKU string) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO purchases (user_id, item_sku) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		userID, itemSKU,
	)
	return err
}

// GetByUser — ambil semua SKU item yang user pernah beli.
// Digunakan untuk: (1) papar inventory tab, (2) semak sama ada limited item dah dibeli.
func (r *PurchaseRepository) GetByUser(ctx context.Context, userID string) ([]string, error) {
	rows, err := r.db.Query(ctx,
		`SELECT item_sku FROM purchases WHERE user_id = $1`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	skus := []string{}
	for rows.Next() {
		var sku string
		if err := rows.Scan(&sku); err != nil {
			continue
		}
		skus = append(skus, sku)
	}
	return skus, rows.Err()
}

// HasPurchased — semak sama ada user dah pernah beli item tertentu.
// Digunakan khusus untuk check limited item (dancing-zombie) sebelum benarkan pembelian.
// Return true kalau dah beli, false kalau belum.
func (r *PurchaseRepository) HasPurchased(ctx context.Context, userID, itemSKU string) (bool, error) {
	var count int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM purchases WHERE user_id = $1 AND item_sku = $2`,
		userID, itemSKU,
	).Scan(&count)
	return count > 0, err
}
