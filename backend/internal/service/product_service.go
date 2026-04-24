package service

// product_service.go — Service layer untuk produk.
// Service adalah layer antara Handler dan Repository.
// Tempat business logic duduk — contoh: kalau nak filter, sort, atau process data
// sebelum hantar ke handler atau simpan ke database.
// Dalam kes ni simple je — terus delegate ke repository.

import (
	"context"

	"sanim-backend/internal/model"
	"sanim-backend/internal/repository"
)

// ProductService — service untuk handle business logic berkaitan produk
type ProductService struct {
	repo *repository.ProductRepository // Guna repo untuk akses database
}

// NewProductService — buat ProductService dengan repository yang diperlukan
func NewProductService(repo *repository.ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

// GetAllProducts — ambil semua produk (delegate terus ke repository)
// Kalau ada business logic macam filter atau transform, buat kat sini
func (s *ProductService) GetAllProducts(ctx context.Context) ([]model.Product, error) {
	return s.repo.GetAll(ctx)
}
