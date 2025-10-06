package users

import (
	"database/sql"

	"github.com/Nassabiq/gpci-compro-api/internal/modules/users/repo/postgres"
	"github.com/Nassabiq/gpci-compro-api/internal/modules/users/service"
)

type Module struct {
	Repository *postgres.UserRepository
	Service    *service.Service
}

func Provide(db *sql.DB) *Module {
	repo := &postgres.UserRepository{DB: db}
	return &Module{
		Repository: repo,
		Service:    service.New(repo),
	}
}
