package service

import (
	"context"

	"github.com/Nassabiq/gpci-compro-api/internal/modules/product/domain"
)

type ProductCertificationRepository interface {
	ListProductCertifications(ctx context.Context, productSlug string, filter domain.ProductCertificationFilter) ([]domain.ProductCertification, int, error)
	CreateProductCertification(ctx context.Context, productSlug string, payload domain.ProductCertificationPayload) (domain.ProductCertification, error)
	UpdateProductCertification(ctx context.Context, productSlug string, certificationID int64, payload domain.ProductCertificationPayload) (domain.ProductCertification, error)
	DeleteProductCertification(ctx context.Context, productSlug string, certificationID int64) error
}

type ProductCertificationService struct {
	repository ProductCertificationRepository
}

func NewProductCertificationService(repository ProductCertificationRepository) *ProductCertificationService {
	return &ProductCertificationService{repository: repository}
}

func (service *ProductCertificationService) ListProductCertifications(ctx context.Context, productSlug string, filter domain.ProductCertificationFilter) (domain.ProductCertificationListResponse, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}

	items, total, err := service.repository.ListProductCertifications(ctx, productSlug, filter)
	if err != nil {
		return domain.ProductCertificationListResponse{}, err
	}

	return domain.ProductCertificationListResponse{
		Items:    items,
		Total:    total,
		Page:     filter.Page,
		PageSize: filter.PageSize,
	}, nil
}

func (service *ProductCertificationService) CreateProductCertification(
	ctx context.Context,
	productSlug string,
	payload domain.ProductCertificationPayload,
) (domain.ProductCertification, error) {
	return service.repository.CreateProductCertification(ctx, productSlug, payload)
}

func (service *ProductCertificationService) UpdateProductCertification(
	ctx context.Context,
	productSlug string,
	certificationID int64,
	payload domain.ProductCertificationPayload,
) (domain.ProductCertification, error) {
	return service.repository.UpdateProductCertification(ctx, productSlug, certificationID, payload)
}

func (service *ProductCertificationService) DeleteProductCertification(
	ctx context.Context,
	productSlug string,
	certificationID int64,
) error {
	return service.repository.DeleteProductCertification(ctx, productSlug, certificationID)
}
