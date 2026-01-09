package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/k1v4/drip_mate/file_upload_service/internal/config"
	"github.com/k1v4/drip_mate/file_upload_service/internal/repository"
	"github.com/k1v4/drip_mate/file_upload_service/internal/service"
	"github.com/k1v4/drip_mate/file_upload_service/internal/transport/grpc"
	"github.com/k1v4/drip_mate/file_upload_service/pkg/logger"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func main() {

	ctx := context.Background()

	uploadLogger := logger.NewLogger()

	ctx = context.WithValue(ctx, logger.LoggerKey, uploadLogger)

	cfg := config.MustLoadConfig()
	if cfg == nil {
		panic("load config fail")
	}

	uploadLogger.Info(ctx, "successfully read config")

	minioClient, err := minio.New(cfg.EndPoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyId, cfg.SecretAccessKey, ""),
		Secure: true,
	})
	if err != nil {
		uploadLogger.Info(ctx, fmt.Sprintf("create minio client error: %v", err))

		return
	}

	uploadLogger.Info(ctx, "successfully create minioClient")

	uploadRepo := repository.NewUploadRepository(cfg.Address, minioClient, cfg.BucketName)

	uploadService := service.NewUploadService(uploadRepo)

	grpcServer, err := grpc.NewServer(ctx, cfg.GRPCServerPort, uploadService)
	if err != nil {
		uploadLogger.Error(ctx, fmt.Sprintf("create grpc server error: %v", err))
		return
	}

	graceCh := make(chan os.Signal, 1)
	signal.Notify(graceCh, syscall.SIGINT, syscall.SIGTERM)

	// запуск сервера
	go func() {
		if err = grpcServer.Start(ctx); err != nil {
			uploadLogger.Error(ctx, err.Error())
		}
	}()

	<-graceCh

	err = grpcServer.Stop(ctx)
	if err != nil {
		uploadLogger.Error(ctx, err.Error())
	}

	uploadLogger.Info(ctx, "Server stopped")
}

func createBucket(minioClient *minio.Client, bucketName string, location string) {
	ctx := context.Background()

	// Проверяем, существует ли бакет
	exists, err := minioClient.BucketExists(ctx, bucketName)
	if err != nil {
		log.Fatalln("Ошибка при проверке существования бакета:", err)
	}

	if exists {
		fmt.Println("Бакет уже существует:", bucketName)
		return
	}

	// Создаем бакет
	err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: location})
	if err != nil {
		log.Fatalln("Ошибка при создании бакета:", err)
	}

	fmt.Println("Бакет успешно создан:", bucketName)
}

func uploadObject(minioClient *minio.Client, bucketName string, objectName string, filePath string, address string) string {
	ctx := context.Background()

	// Загружаем файл
	_, err := minioClient.FPutObject(ctx, bucketName, objectName, filePath, minio.PutObjectOptions{})
	if err != nil {
		log.Fatalln("Ошибка при загрузке файла:", err)
	}

	fmt.Println("Файл успешно загружен:", objectName)

	return fmt.Sprintf("%s/%s", address, objectName)
}

func downloadObject(minioClient *minio.Client, bucketName string, objectName string, filePath string) {
	ctx := context.Background()

	// Скачиваем файл
	err := minioClient.FGetObject(ctx, bucketName, objectName, filePath, minio.GetObjectOptions{})
	if err != nil {
		log.Fatalln("Ошибка при скачивании файла:", err)
	}

	fmt.Println("Файл успешно скачан:", objectName)
}

func deleteObject(minioClient *minio.Client, bucketName string, objectName string) {
	ctx := context.Background()

	// Удаляем объект
	err := minioClient.RemoveObject(ctx, bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		log.Fatalln("Ошибка при удалении файла:", err)
	}

	fmt.Println("Файл успешно удален:", objectName)
}
