package app

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/Nassabiq/gpci-compro-api/internal/config"
	"github.com/Nassabiq/gpci-compro-api/internal/db"
	"github.com/Nassabiq/gpci-compro-api/internal/utils"
	"github.com/hibiken/asynq"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Container struct {
	Config      *config.Config
	Logger      *slog.Logger
	DB          *sql.DB
	AsynqClient *asynq.Client
	Storage     *minio.Client
}

func New(ctx context.Context, cfg *config.Config) (*Container, func(), error) {
	if ctx == nil {
		ctx = context.Background()
	}
	logger := utils.NewLogger(cfg.App.Env)

	dsn := db.DSN(cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Password, cfg.DB.Name, cfg.DB.SSLMode)
	database, err := db.New(dsn, cfg.DB.MaxOpenConns, cfg.DB.MaxIdleConns, cfg.DB.ConnMaxLifetime)
	if err != nil {
		return nil, nil, fmt.Errorf("init db: %w", err)
	}

	redisOpt := asynq.RedisClientOpt{Addr: cfg.Redis.Addr, Password: cfg.Redis.Password, DB: cfg.Redis.DB}
	asynqClient := asynq.NewClient(redisOpt)

	minioClient, err := minio.New(cfg.Storage.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.Storage.AccessKey, cfg.Storage.SecretKey, ""),
		Secure: cfg.Storage.UseSSL,
		Region: cfg.Storage.Region,
	})
	if err != nil {
		asynqClient.Close()
		database.Close()
		return nil, nil, fmt.Errorf("init storage: %w", err)
	}

	container := &Container{
		Config:      cfg,
		Logger:      logger,
		DB:          database,
		AsynqClient: asynqClient,
		Storage:     minioClient,
	}

	cleanup := func() {
		_ = asynqClient.Close()
		_ = database.Close()
	}

	if err := container.ensureBucket(ctx, cfg.Storage.Bucket, cfg.Storage.Region); err != nil {
		cleanup()
		return nil, nil, err
	}

	return container, cleanup, nil
}

func (c *Container) ensureBucket(ctx context.Context, bucket, region string) error {
	bucketCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := c.Storage.MakeBucket(bucketCtx, bucket, minio.MakeBucketOptions{Region: region}); err != nil {
		exists, errExists := c.Storage.BucketExists(bucketCtx, bucket)
		if errExists != nil {
			return fmt.Errorf("check bucket: %w", errExists)
		}
		if !exists {
			return fmt.Errorf("create bucket: %w", err)
		}
	}
	return nil
}
