package faq

import (
	"database/sql"

	"github.com/Nassabiq/gpci-compro-api/internal/modules/faq/repo/postgres"
	"github.com/Nassabiq/gpci-compro-api/internal/modules/faq/service"
)

type Module struct {
	Repository *postgres.FAQRepository
	Service    *service.Service
}

func Provide(db *sql.DB) *Module {
	repo := postgres.NewFAQRepository(db)
	return &Module{
		Repository: repo,
		Service:    service.NewService(repo),
	}
}
