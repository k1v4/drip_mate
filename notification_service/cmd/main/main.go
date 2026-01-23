package main

import (
	"context"
	"fmt"
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

	emailNotificationUseCase := usecase.NewEmailNotificationUseCase()

	handler := echo.New()

	v1.NewRouter(handler, notificationLogger, emailNotificationUseCase)

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
