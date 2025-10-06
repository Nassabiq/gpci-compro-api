package catalog

import (
	"database/sql"

	catalogrepo "github.com/Nassabiq/gpci-compro-api/internal/modules/catalog/repo/postgres"
	"github.com/Nassabiq/gpci-compro-api/internal/modules/catalog/service"
)

type Module struct {
	Repository *catalogrepo.Repository
	Service    *service.Service
}

func Provide(db *sql.DB) *Module {
	repo := catalogrepo.NewRepository(db)
	return &Module{
		Repository: repo,
		Service:    service.New(repo),
	}
}
