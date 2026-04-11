package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/labstack/echo/v4"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/k1v4/drip_mate/internal/config"
	"github.com/k1v4/drip_mate/internal/modules/notification_service"
	controllerNotif "github.com/k1v4/drip_mate/internal/modules/notification_service/controller/http/v1"
	serviceNotif "github.com/k1v4/drip_mate/internal/modules/notification_service/usecase"
	repositoryObject "github.com/k1v4/drip_mate/internal/modules/object_gateway/repository"
	serviceObject "github.com/k1v4/drip_mate/internal/modules/object_gateway/service"
	grpcTransport "github.com/k1v4/drip_mate/internal/modules/object_gateway/transport/grpc"
	serviceUser "github.com/k1v4/drip_mate/internal/modules/user_service/usecase"
	repositoryUser "github.com/k1v4/drip_mate/internal/modules/user_service/usecase/repository"
	"github.com/k1v4/drip_mate/internal/router"
	"github.com/k1v4/drip_mate/pkg/DataBase/postgres"
	"github.com/k1v4/drip_mate/pkg/adapter"
	"github.com/k1v4/drip_mate/pkg/httpserver"
	"github.com/k1v4/drip_mate/pkg/kafkaPkg"
	"github.com/k1v4/drip_mate/pkg/logger"
)

func Run() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serviceLogger := logger.NewLogger()
	ctx = context.WithValue(ctx, logger.LoggerKey, serviceLogger)

	cfg := config.MustLoadConfig()
	if cfg == nil {
		panic("load config fail")
	}
	serviceLogger.Info(ctx, "read config successfully")

	pg, err := postgres.New(cfg.DB.URL, postgres.MaxPoolSize(cfg.DB.PoolMax))
	if err != nil {
		serviceLogger.Error(ctx, fmt.Sprintf("postgres.New: %s", err))
		return
	}
	defer pg.Close()
	serviceLogger.Info(ctx, "connected to database successfully")

	err = makeMigrate(cfg.DB.URL)
	if err != nil {
		serviceLogger.Error(ctx, fmt.Sprintf("migrate: %s", err))
		return
	}

	serviceLogger.Info(ctx, "migrations applied successfully")

	minioClient, err := minio.New(cfg.ObjectStorage.EndPoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.ObjectStorage.AccessKeyID, cfg.ObjectStorage.SecretAccessKey, ""),
		Secure: true,
	})
	if err != nil {
		serviceLogger.Error(ctx, fmt.Sprintf("create minio client error: %v", err))
		return
	}
	serviceLogger.Info(ctx, "minio client created successfully")

	templates, err := notification_service.NewTemplates()
	if err != nil {
		serviceLogger.Error(ctx, fmt.Sprintf("parse templates: %v", err))
		return
	}

	kafkaProducer := kafkaPkg.NewProducer([]string{cfg.Kafka.Brokers}, cfg.Kafka.Topic)
	if kafkaProducer == nil {
		serviceLogger.Error(ctx, fmt.Sprintf("create kafka producer error: %v", err))
		return
	}
	defer func() {
		err := kafkaProducer.Close()
		if err != nil {
			serviceLogger.Error(ctx, fmt.Sprintf("close kafka producer error: %v", err))
		}
	}()

	email := adapter.NewGoMailClient(cfg.SMTP)

	authRepo := repositoryUser.NewAuthRepository(pg)
	uploadRepo := repositoryObject.NewUploadRepository(cfg.ObjectStorage.Address, minioClient, cfg.ObjectStorage.BucketName)

	authUseCase := serviceUser.NewAuthUseCase(authRepo, serviceLogger, kafkaProducer, new(cfg.Token))
	uploadService := serviceObject.NewUploadService(uploadRepo)
	notifUseCase := serviceNotif.NewEmailNotificationUseCase(email, serviceLogger, templates)

	notifController := controllerNotif.NewEmailController(notifUseCase)

	e := echo.New()
	e.HideBanner = true
	e.HTTPErrorHandler = makeHTTPErrorHandler(serviceLogger)
	router.NewRouter(e, serviceLogger, authUseCase, cfg)

	httpServer := httpserver.New(e, httpserver.Port(strconv.Itoa(cfg.Server.RestPort)))
	grpcServer, err := grpcTransport.NewServer(ctx, cfg.Server.GRPCPort, uploadService)
	if err != nil {
		serviceLogger.Error(ctx, fmt.Sprintf("create grpc server error: %v", err))
		return
	}
	kafkaReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{cfg.Kafka.Brokers},
		Topic:          cfg.Kafka.Topic,
		GroupID:        cfg.Kafka.GroupID,
		CommitInterval: 0,
	})

	// создаём consumer
	kafkaConsumer := kafkaPkg.NewConsumer(kafkaReader, notifController, kafkaProducer, serviceLogger)
	if kafkaConsumer == nil {
		serviceLogger.Error(ctx, fmt.Sprintf("create kafka consumer error: %v", err))
		return
	}

	// Запускаем оба сервера параллельно через errgroup
	eg, egCtx := errgroup.WithContext(ctx)

	// запускаем в errgroup
	eg.Go(func() error {
		serviceLogger.Info(ctx, "kafka consumer starting")
		if err := kafkaConsumer.Run(egCtx); err != nil {
			return fmt.Errorf("kafka consumer: %w", err)
		}
		return nil
	})

	eg.Go(func() error {
		serviceLogger.Info(ctx, fmt.Sprintf("http server starting on :%d", cfg.Server.RestPort))
		select {
		case err = <-httpServer.Notify():
			return fmt.Errorf("http server: %w", err)
		case <-egCtx.Done():
			return httpServer.Shutdown()
		}
	})

	eg.Go(func() error {
		serviceLogger.Info(ctx, fmt.Sprintf("grpc server starting on :%d", cfg.Server.GRPCPort))
		if err = grpcServer.Start(egCtx); err != nil {
			return fmt.Errorf("grpc server: %w", err)
		}
		return nil
	})

	eg.Go(func() error {
		<-egCtx.Done()
		serviceLogger.Info(ctx, "shutdown signal received")
		return grpcServer.Stop(ctx)
	})

	if err = eg.Wait(); err != nil {
		serviceLogger.Error(ctx, fmt.Sprintf("server error: %v", err))
	}

	serviceLogger.Info(ctx, "server stopped gracefully")
}

func makeMigrate(url string) error {
	m, err := migrate.New("file://migrations", url)
	if err != nil {
		return fmt.Errorf("migration setup failed: %v", err)
	}
	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migration failed: %v", err)
	}

	return nil
}

func makeHTTPErrorHandler(l logger.Logger) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		var he *echo.HTTPError
		if !errors.As(err, &he) {
			he = echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
		}

		if he.Internal != nil {
			l.Error(c.Request().Context(), "http error",
				zap.Int("status", he.Code),
				zap.String("reason", he.Internal.Error()),
				zap.String("path", c.Request().URL.Path),
			)
		}

		if !c.Response().Committed {
			_ = c.JSON(he.Code, echo.Map{"error": he.Message})
		}
	}
}
