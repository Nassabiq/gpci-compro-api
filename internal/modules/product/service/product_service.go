package service

import (
	"context"

	"github.com/Nassabiq/gpci-compro-api/internal/modules/product/domain"
)

type ProductRepository interface {
	ListProducts(ctx context.Context, filter domain.ProductFilter) ([]domain.Product, int, error)
	CreateProduct(ctx context.Context, payload domain.ProductPayload) (domain.Product, error)
	GetProductBySlug(ctx context.Context, slug string) (*domain.Product, error)
	UpdateProduct(ctx context.Context, slug string, payload domain.ProductPayload) (domain.Product, error)
	DeleteProduct(ctx context.Context, slug string) error
}

type ProductService struct {
	repo ProductRepository
}

func NewProductService(repo ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

func (s *ProductService) ListProducts(ctx context.Context, filter domain.ProductFilter) (domain.ProductListResponse, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}

	items, total, err := s.repo.ListProducts(ctx, filter)
	if err != nil {
		return domain.ProductListResponse{}, err
	}
	return domain.ProductListResponse{
		Items:    items,
		Total:    total,
		Page:     filter.Page,
		PageSize: filter.PageSize,
	}, nil
}

func (s *ProductService) CreateProduct(ctx context.Context, payload domain.ProductPayload) (domain.Product, error) {
	return s.repo.CreateProduct(ctx, payload)
}

func (s *ProductService) GetProductBySlug(ctx context.Context, slug string) (*domain.Product, error) {
	return s.repo.GetProductBySlug(ctx, slug)
}

func (s *ProductService) UpdateProduct(ctx context.Context, slug string, payload domain.ProductPayload) (domain.Product, error) {
	return s.repo.UpdateProduct(ctx, slug, payload)
}

func (s *ProductService) DeleteProduct(ctx context.Context, slug string) error {
	return s.repo.DeleteProduct(ctx, slug)
}
