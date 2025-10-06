package rbac

import (
	"database/sql"

	"github.com/Nassabiq/gpci-compro-api/internal/modules/rbac/repo/postgres"
	"github.com/Nassabiq/gpci-compro-api/internal/modules/rbac/service"
)

type Module struct {
	Repository *postgres.RBACRepository
	Service    *service.Service
}

func Provide(db *sql.DB) *Module {
	repo := &postgres.RBACRepository{DB: db}
	return &Module{
		Repository: repo,
		Service:    service.New(repo),
	}
}
