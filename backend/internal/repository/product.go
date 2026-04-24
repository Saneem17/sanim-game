package repository

// product.go — Repository layer untuk produk.
// Repository adalah layer yang bertanggungjawab buat SQL queries ke database.
// Dia tak tahu pasal business logic — kerja dia just baca/tulis data.
// Handler → Service → Repository → Database

import (
	"context"

	"sanim-backend/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ProductRepository — simpan database connection untuk buat queries
type ProductRepository struct {
	db *pgxpool.Pool // Connection pool ke PostgreSQL
}

// NewProductRepository — buat ProductRepository dengan database pool
func NewProductRepository(db *pgxpool.Pool) *ProductRepository {
	return &ProductRepository{db: db}
}

// GetAll — ambil SEMUA produk dari table `items` dalam database, tersusun by ID
func (r *ProductRepository) GetAll(ctx context.Context) ([]model.Product, error) {
	// SQL query untuk ambil semua column dari table items, sorted by ID
	query := `
		SELECT id, name, description, price, currency, sku, image_url, created_at
		FROM items
		ORDER BY id ASC
	`

	// Jalankan query — rows adalah result set (macam cursor)
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close() // Pastikan rows ditutup bila selesai untuk release resources

	var products []model.Product

	// Iterate through setiap row dalam result
	for rows.Next() {
		var p model.Product
		// Scan — assign nilai dari setiap column ke field dalam struct Product
		// Order mesti sama dengan order dalam SELECT statement!
		err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.Description,
			&p.Price,
			&p.Currency,
			&p.SKU,
			&p.ImageURL,
			&p.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		products = append(products, p) // Tambah product ke list
	}

	// Check kalau ada error semasa iterate rows (selain error row-by-row)
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return products, nil // Return list semua produk
}
