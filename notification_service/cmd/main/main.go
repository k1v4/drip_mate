package main

import (
	"context"
	"fmt"
	"notification_service/pkg/adapter"
	kafkaPkg "notification_service/pkg/kafka"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"notification_service/internal/config"
	v1 "notification_service/internal/controller/http/v1"
	"notification_service/internal/usecase"
	"notification_service/pkg/httpserver"
	"notification_service/pkg/logger"

	"github.com/labstack/echo/v4"
	"github.com/segmentio/kafka-go"
)

func main() {
	ctx := context.Background()

	notificationLogger := logger.NewLogger()
	ctx = context.WithValue(ctx, logger.LoggerKey, notificationLogger)

	cfg := config.MustLoadConfig()
	if cfg == nil {
		panic("load config fail")
	}

	notificationLogger.Info(ctx, "read config successfully")

	// TODO запихать в конфиг ключ
	client := adapter.NewSendGridClient("")
	if client == nil {
		notificationLogger.Error(ctx, "create email notification service fail")
		return
	}

	emailNotificationUseCase := usecase.NewEmailNotificationUseCase(client)

	controller := v1.NewEmailController(emailNotificationUseCase)
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "my-topic",
		GroupID: "my-groupID",
	})
	defer func() {
		if err := reader.Close(); err != nil {
			notificationLogger.Error(ctx, fmt.Sprintf("reader close error: %v", err))
		}
	}()

	consumer := kafkaPkg.NewConsumer(reader, controller, notificationLogger)
	if consumer == nil {
		notificationLogger.Error(ctx, "create email notification service fail")
	}

	handler := echo.New()
	v1.NewRouter(handler, notificationLogger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := consumer.Run(ctx); err != nil {
			notificationLogger.Error(ctx, fmt.Sprintf("consumer stopped with error: %v", err))
			cancel()
		}
	}()

	httpServer := httpserver.New(handler, httpserver.Port(strconv.Itoa(cfg.RestServerPort)))

	// signal for graceful shutdown
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	var err error
	select {
	case s := <-interrupt:
		notificationLogger.Info(ctx, "app-Run-signal: "+s.String())
	case err = <-httpServer.Notify():
		notificationLogger.Error(ctx, fmt.Sprintf("app-Run-httpServer.Notify: %s", err))
	}

	// shutdown
	err = httpServer.Shutdown()
	if err != nil {
		notificationLogger.Error(ctx, fmt.Sprintf("app-Run-httpServer.Shutdown: %s", err))
	}

}
