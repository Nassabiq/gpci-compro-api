package uploads

import (
	"github.com/minio/minio-go/v7"

	miniorepo "github.com/Nassabiq/gpci-compro-api/internal/modules/uploads/repo/minio"
	"github.com/Nassabiq/gpci-compro-api/internal/modules/uploads/service"
)

type Module struct {
	Repository *miniorepo.Repository
	Service    *service.Service
}

func Provide(client *minio.Client, bucket, basePath string, modulePaths map[string]string) *Module {
	repo := miniorepo.New(client)
	return &Module{
		Repository: repo,
		Service:    service.New(repo, bucket, basePath, modulePaths),
	}
}
