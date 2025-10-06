package brand

import (
	"database/sql"

	"github.com/Nassabiq/gpci-compro-api/internal/modules/brand/repo/postgres"
	"github.com/Nassabiq/gpci-compro-api/internal/modules/brand/service"
)

type Module struct {
	Repository *postgres.BrandRepository
	Service    *service.BrandService
}

func Provide(db *sql.DB) *Module {
	repo := postgres.NewBrandRepository(db)
	return &Module{
		Repository: repo,
		Service:    service.NewBrandService(repo),
	}
}
