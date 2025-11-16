package service

import (
	"context"

	"github.com/Nassabiq/gpci-compro-api/internal/modules/users/domain"
)

type Repository interface {
	Create(ctx context.Context, name, email, passwordHash string, isActive bool) (int64, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindByID(ctx context.Context, id int64) (*domain.User, error)
	FindByXID(ctx context.Context, xid string) (*domain.User, error)
	List(ctx context.Context) ([]domain.User, error)
	Update(ctx context.Context, xid, name, email string, passwordHash *string, isActive bool) (*domain.User, error)
	Delete(ctx context.Context, xid string) error
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, name, email, passwordHash string, isActive bool) (int64, error) {
	return s.repo.Create(ctx, name, email, passwordHash, isActive)
}

func (s *Service) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	return s.repo.FindByEmail(ctx, email)
}

func (s *Service) FindByID(ctx context.Context, id int64) (*domain.User, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) FindByXID(ctx context.Context, xid string) (*domain.User, error) {
	return s.repo.FindByXID(ctx, xid)
}

func (s *Service) List(ctx context.Context) ([]domain.User, error) {
	return s.repo.List(ctx)
}

func (s *Service) Update(ctx context.Context, xid, name, email string, passwordHash *string, isActive bool) (*domain.User, error) {
	return s.repo.Update(ctx, xid, name, email, passwordHash, isActive)
}

func (s *Service) Delete(ctx context.Context, xid string) error {
	return s.repo.Delete(ctx, xid)
}
