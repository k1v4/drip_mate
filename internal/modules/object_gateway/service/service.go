package service

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type IUploadRepository interface {
	UploadImage(ctx context.Context, fileName string, imageData []byte) (string, error)
	DeleteImage(ctx context.Context, fileName string) error
}

type UploadServer struct {
	upRepo IUploadRepository
}

func NewUploadService(upRepo IUploadRepository) *UploadServer {
	return &UploadServer{
		upRepo: upRepo,
	}
}

func (s *UploadServer) UploadImage(ctx context.Context, fileName string, imageData []byte) (string, error) {
	const op = "service.GetShoes"

	currentTime := time.Now()

	formattedTime := currentTime.Format("2006-01-02 15:04:05")
	formattedTime = strings.ReplaceAll(formattedTime, "-", "")
	formattedTime = strings.ReplaceAll(formattedTime, ":", "-")
	formattedTime = strings.ReplaceAll(formattedTime, " ", "_")

	arr := strings.Split(fileName, ".")
	imageType := arr[len(arr)-1]
	arr = arr[:len(arr)-1]

	urlImage, err := s.upRepo.UploadImage(ctx,
		fmt.Sprintf("%s_%s.%s", strings.Join(arr, "."),
			formattedTime,
			imageType,
		),
		imageData,
	)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return urlImage, nil
}

func (s *UploadServer) DeleteImage(ctx context.Context, url string) (bool, error) {
	const op = "service.DeleteImage"

	urlArr := strings.Split(url, "/")
	fileName := urlArr[len(urlArr)-1]

	err := s.upRepo.DeleteImage(ctx, fileName)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return true, nil
}
