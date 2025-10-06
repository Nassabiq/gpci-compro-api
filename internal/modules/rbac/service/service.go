package service

import (
	"context"

	"github.com/Nassabiq/gpci-compro-api/internal/modules/rbac/domain"
)

type Repository interface {
	CreateRole(ctx context.Context, name string) (int64, error)
	CreatePermission(ctx context.Context, key, description string) (int64, error)
	ListRoles(ctx context.Context) ([]domain.Role, error)
	ListPermissions(ctx context.Context) ([]domain.Permission, error)
	AssignPermissionToRole(ctx context.Context, roleName, permKey string) error
	AssignRoleToUserByXID(ctx context.Context, userXID, roleName string) error
	UserHasPermissionByXID(ctx context.Context, userXID, permKey string) (bool, error)
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateRole(ctx context.Context, name string) (int64, error) {
	return s.repo.CreateRole(ctx, name)
}

func (s *Service) CreatePermission(ctx context.Context, key, description string) (int64, error) {
	return s.repo.CreatePermission(ctx, key, description)
}

func (s *Service) ListRoles(ctx context.Context) ([]domain.Role, error) {
	return s.repo.ListRoles(ctx)
}

func (s *Service) ListPermissions(ctx context.Context) ([]domain.Permission, error) {
	return s.repo.ListPermissions(ctx)
}

func (s *Service) AssignPermissionToRole(ctx context.Context, roleName, permKey string) error {
	return s.repo.AssignPermissionToRole(ctx, roleName, permKey)
}

func (s *Service) AssignRoleToUserByXID(ctx context.Context, userXID, roleName string) error {
	return s.repo.AssignRoleToUserByXID(ctx, userXID, roleName)
}

func (s *Service) UserHasPermissionByXID(ctx context.Context, userXID, permKey string) (bool, error) {
	return s.repo.UserHasPermissionByXID(ctx, userXID, permKey)
}
