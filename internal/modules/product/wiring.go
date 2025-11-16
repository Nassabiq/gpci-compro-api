package product

import (
	"database/sql"

	"github.com/Nassabiq/gpci-compro-api/internal/modules/product/repo/postgres"
	"github.com/Nassabiq/gpci-compro-api/internal/modules/product/service"
)

type Module struct {
	Service              *service.ProductService
	CertificationService *service.ProductCertificationService
	ProgramCertService   *service.ProgramCertificateService
	ProductRepository    service.ProductRepository
	CertificationRepo    service.ProductCertificationRepository
}

func Provide(db *sql.DB) *Module {
	productRepo := postgres.NewProductRepository(db)
	certRepo := postgres.NewProductCertificationRepository(db)
	certService := service.NewProductCertificationService(certRepo)

	return &Module{
		ProductRepository:    productRepo,
		CertificationRepo:    certRepo,
		Service:              service.NewProductService(productRepo),
		CertificationService: certService,
		ProgramCertService:   service.NewProgramCertificateService(certRepo, certService),
	}
}
