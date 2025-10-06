package service

import (
	"context"

	"github.com/Nassabiq/gpci-compro-api/internal/modules/catalog/domain"
)

type Repository interface {
	ListPrograms(ctx context.Context) ([]domain.Program, error)
	CreateProgram(ctx context.Context, payload domain.ProgramPayload) (domain.Program, error)
	UpdateProgram(ctx context.Context, id int16, payload domain.ProgramPayload) (domain.Program, error)
	DeleteProgram(ctx context.Context, id int16) error
	ListStatuses(ctx context.Context) ([]domain.CertificationStatus, error)
	CreateStatus(ctx context.Context, payload domain.StatusPayload) (domain.CertificationStatus, error)
	UpdateStatus(ctx context.Context, id int16, payload domain.StatusPayload) (domain.CertificationStatus, error)
	DeleteStatus(ctx context.Context, id int16) error
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListPrograms(ctx context.Context) ([]domain.Program, error) {
	return s.repo.ListPrograms(ctx)
}

func (s *Service) CreateProgram(ctx context.Context, payload domain.ProgramPayload) (domain.Program, error) {
	return s.repo.CreateProgram(ctx, payload)
}

func (s *Service) UpdateProgram(ctx context.Context, id int16, payload domain.ProgramPayload) (domain.Program, error) {
	return s.repo.UpdateProgram(ctx, id, payload)
}

func (s *Service) DeleteProgram(ctx context.Context, id int16) error {
	return s.repo.DeleteProgram(ctx, id)
}

func (s *Service) ListStatuses(ctx context.Context) ([]domain.CertificationStatus, error) {
	return s.repo.ListStatuses(ctx)
}

func (s *Service) CreateStatus(ctx context.Context, payload domain.StatusPayload) (domain.CertificationStatus, error) {
	return s.repo.CreateStatus(ctx, payload)
}

func (s *Service) UpdateStatus(ctx context.Context, id int16, payload domain.StatusPayload) (domain.CertificationStatus, error) {
	return s.repo.UpdateStatus(ctx, id, payload)
}

func (s *Service) DeleteStatus(ctx context.Context, id int16) error {
	return s.repo.DeleteStatus(ctx, id)
}
