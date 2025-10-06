package service

import (
	"context"

	"github.com/Nassabiq/gpci-compro-api/internal/modules/brand/domain"
)

type BrandRepository interface {
	ListBrandCategories(ctx context.Context, filter domain.BrandCategoryFilter) ([]domain.BrandCategory, int, error)
	CreateBrandCategory(ctx context.Context, payload domain.BrandCategoryPayload) (domain.BrandCategory, error)
	UpdateBrandCategory(ctx context.Context, id int64, payload domain.BrandCategoryPayload) (domain.BrandCategory, error)
	DeleteBrandCategory(ctx context.Context, id int64) error
	ListBrands(ctx context.Context, filter domain.BrandFilter) ([]domain.Brand, int, error)
	CreateBrand(ctx context.Context, payload domain.BrandPayload) (domain.Brand, error)
	UpdateBrand(ctx context.Context, id int64, payload domain.BrandPayload) (domain.Brand, error)
	DeleteBrand(ctx context.Context, id int64) error
}

type BrandService struct {
	repo BrandRepository
}

func NewBrandService(repo BrandRepository) *BrandService {
	return &BrandService{repo: repo}
}

func (s *BrandService) ListBrandCategories(ctx context.Context, filter domain.BrandCategoryFilter) (domain.BrandCategoryListResponse, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}

	items, total, err := s.repo.ListBrandCategories(ctx, filter)
	if err != nil {
		return domain.BrandCategoryListResponse{}, err
	}

	return domain.BrandCategoryListResponse{
		Items:    items,
		Total:    total,
		Page:     filter.Page,
		PageSize: filter.PageSize,
	}, nil
}

func (s *BrandService) CreateBrandCategory(ctx context.Context, payload domain.BrandCategoryPayload) (domain.BrandCategory, error) {
	return s.repo.CreateBrandCategory(ctx, payload)
}

func (s *BrandService) UpdateBrandCategory(ctx context.Context, id int64, payload domain.BrandCategoryPayload) (domain.BrandCategory, error) {
	return s.repo.UpdateBrandCategory(ctx, id, payload)
}

func (s *BrandService) DeleteBrandCategory(ctx context.Context, id int64) error {
	return s.repo.DeleteBrandCategory(ctx, id)
}

func (s *BrandService) ListBrands(ctx context.Context, filter domain.BrandFilter) (domain.BrandListResponse, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}

	items, total, err := s.repo.ListBrands(ctx, filter)
	if err != nil {
		return domain.BrandListResponse{}, err
	}
	return domain.BrandListResponse{
		Items:    items,
		Total:    total,
		Page:     filter.Page,
		PageSize: filter.PageSize,
	}, nil
}

func (s *BrandService) CreateBrand(ctx context.Context, payload domain.BrandPayload) (domain.Brand, error) {
	return s.repo.CreateBrand(ctx, payload)
}

func (s *BrandService) UpdateBrand(ctx context.Context, id int64, payload domain.BrandPayload) (domain.Brand, error) {
	return s.repo.UpdateBrand(ctx, id, payload)
}

func (s *BrandService) DeleteBrand(ctx context.Context, id int64) error {
	return s.repo.DeleteBrand(ctx, id)
}
