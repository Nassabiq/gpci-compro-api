package minio

import (
	"context"
	"io"

	"github.com/minio/minio-go/v7"
)

type Repository struct {
	client *minio.Client
}

func New(client *minio.Client) *Repository {
	return &Repository{client: client}
}

func (r *Repository) PutObject(ctx context.Context, bucket, objectName string, reader io.ReadSeeker, size int64, contentType string) error {
	_, err := r.client.PutObject(
		ctx,
		bucket,
		objectName,
		reader,
		size,
		minio.PutObjectOptions{ContentType: contentType},
	)
	return err
}
