package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"object_gateway/internal/config"
	"object_gateway/internal/repository"
	"object_gateway/internal/service"
	"object_gateway/internal/transport/grpc"
	"object_gateway/pkg/logger"

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
