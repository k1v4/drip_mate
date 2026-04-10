package repository

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/minio/minio-go/v7"
)

type UploadRepository struct {
	address     string
	minioClient *minio.Client
	bucketName  string
}

func NewUploadRepository(address string, minioClient *minio.Client, bucketName string) *UploadRepository {
	return &UploadRepository{
		address:     address,
		minioClient: minioClient,
		bucketName:  bucketName,
	}
}

func (ur *UploadRepository) UploadImage(ctx context.Context, fileName string, imageData []byte) (string, error) {
	const op = "repository.UploadImage"

	reader := bytes.NewReader(imageData)

	suffix := strings.Split(fileName, ".")[len(strings.Split(fileName, "."))-1]

	// Загружаем файл в MinIO
	_, err := ur.minioClient.PutObject(ctx, ur.bucketName, fileName, reader, int64(len(imageData)), minio.PutObjectOptions{
		ContentType: fmt.Sprintf("image/%s", suffix),
	})
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return fmt.Sprintf("%s/%s", ur.address, fileName), nil
}

func (ur *UploadRepository) DeleteImage(ctx context.Context, fileName string) error {
	const op = "repository.DeleteImage"

	err := ur.minioClient.RemoveObject(ctx, ur.bucketName, fileName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
