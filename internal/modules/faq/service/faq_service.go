package service

import (
	"context"

	"github.com/Nassabiq/gpci-compro-api/internal/modules/faq/domain"
)

type Repository interface {
	ListFAQs(ctx context.Context, filter domain.FAQFilter) ([]domain.FAQ, int, error)
	CreateFAQ(ctx context.Context, payload domain.FAQPayload) (domain.FAQ, error)
	GetFAQByID(ctx context.Context, id int64) (*domain.FAQ, error)
	UpdateFAQ(ctx context.Context, id int64, payload domain.FAQPayload) (domain.FAQ, error)
	DeleteFAQ(ctx context.Context, id int64) error
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListFAQs(ctx context.Context, filter domain.FAQFilter) (domain.FAQListResponse, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}

	items, total, err := s.repo.ListFAQs(ctx, filter)
	if err != nil {
		return domain.FAQListResponse{}, err
	}

	return domain.FAQListResponse{
		Items:    items,
		Total:    total,
		Page:     filter.Page,
		PageSize: filter.PageSize,
	}, nil
}

func (s *Service) CreateFAQ(ctx context.Context, payload domain.FAQPayload) (domain.FAQ, error) {
	return s.repo.CreateFAQ(ctx, payload)
}

func (s *Service) GetFAQByID(ctx context.Context, id int64) (*domain.FAQ, error) {
	return s.repo.GetFAQByID(ctx, id)
}

func (s *Service) UpdateFAQ(ctx context.Context, id int64, payload domain.FAQPayload) (domain.FAQ, error) {
	return s.repo.UpdateFAQ(ctx, id, payload)
}

func (s *Service) DeleteFAQ(ctx context.Context, id int64) error {
	return s.repo.DeleteFAQ(ctx, id)
}
