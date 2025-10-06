package service

import (
	"context"

	"github.com/Nassabiq/gpci-compro-api/internal/modules/users/domain"
)

type Repository interface {
	Create(ctx context.Context, name, email, passwordHash string) (int64, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, name, email, passwordHash string) (int64, error) {
	return s.repo.Create(ctx, name, email, passwordHash)
}

func (s *Service) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	return s.repo.FindByEmail(ctx, email)
}
