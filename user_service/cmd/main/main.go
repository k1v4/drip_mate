package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"user_service/internal/config"
	v1 "user_service/internal/controller/http/v1"
	"user_service/internal/usecase"
	"user_service/internal/usecase/repository"
	"user_service/pkg/DataBase/postgres"
	"user_service/pkg/httpserver"
	"user_service/pkg/logger"

	"github.com/labstack/echo/v4"
)

func main() {
	ctx := context.Background()

	userLogger := logger.NewLogger()
	ctx = context.WithValue(ctx, logger.LoggerKey, userLogger)

	cfg := config.MustLoadConfig()
	if cfg == nil {
		panic("load config fail")
	}

	userLogger.Info(ctx, "read config successfully")

	url := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.UserName,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DbName,
	)

	pg, err := postgres.New(url, postgres.MaxPoolSize(cfg.PoolMax))
	if err != nil {
		userLogger.Error(ctx, fmt.Sprintf("app - Run - postgres.New: %s", err))
	}
	defer pg.Close()

	userLogger.Info(ctx, "connected to database successfully")

	authRepo := repository.NewAuthRepository(pg)

	authUseCase := usecase.NewAuthUseCase(authRepo, cfg.TokenTTL, cfg.RefreshTokenTTL)

	handler := echo.New()

	v1.NewRouter(handler, userLogger, authUseCase)

	httpServer := httpserver.New(handler, httpserver.Port(strconv.Itoa(cfg.RestServerPort)))

	// signal for graceful shutdown
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		userLogger.Info(ctx, "app-Run-signal: "+s.String())
	case err = <-httpServer.Notify():
		userLogger.Error(ctx, fmt.Sprintf("app-Run-httpServer.Notify: %s", err))
	}

	// shutdown
	err = httpServer.Shutdown()
	if err != nil {
		userLogger.Error(ctx, fmt.Sprintf("app-Run-httpServer.Shutdown: %s", err))
	}

}
