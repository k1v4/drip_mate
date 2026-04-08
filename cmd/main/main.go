package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/k1v4/drip_mate/internal/config"
	"github.com/k1v4/drip_mate/pkg/DataBase/postgres"
	"github.com/k1v4/drip_mate/pkg/httpserver"
	"github.com/k1v4/drip_mate/pkg/logger"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func main() {
	ctx := context.Background()

	serviceLogger := logger.NewLogger()
	ctx = context.WithValue(ctx, logger.LoggerKey, serviceLogger)

	cfg := config.MustLoadConfig()
	if cfg == nil {
		panic("load config fail")
	}

	serviceLogger.Info(ctx, "read config successfully")

	pg, err := postgres.New(cfg.URL, postgres.MaxPoolSize(cfg.PoolMax))
	if err != nil {
		serviceLogger.Error(ctx, fmt.Sprintf("app - Run - postgres.New: %s", err))
	}
	defer pg.Close()

	serviceLogger.Info(ctx, "connected to database successfully")

	e := echo.New()
	e.HideBanner = true

	e.HTTPErrorHandler = func(err error, c echo.Context) {
		var he *echo.HTTPError
		if !errors.As(err, &he) {
			he = echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
		}

		if he.Internal != nil {
			ctx = c.Request().Context()
			serviceLogger.Error(ctx, "http error",
				zap.Int("status", he.Code),
				zap.String("reason", he.Internal.Error()),
				zap.String("path", c.Request().URL.Path),
			)
		}

		if !c.Response().Committed {
			_ = c.JSON(he.Code, echo.Map{"error": he.Message})
		}
	}

	httpServer := httpserver.New(e, httpserver.Port(strconv.Itoa(cfg.RestServerPort)))

	// signal for graceful shutdown
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		serviceLogger.Info(ctx, "app-Run-signal: "+s.String())
	case err = <-httpServer.Notify():
		serviceLogger.Error(ctx, fmt.Sprintf("app-Run-httpServer.Notify: %s", err))
	}

	// shutdown
	err = httpServer.Shutdown()
	if err != nil {
		serviceLogger.Error(ctx, fmt.Sprintf("app-Run-httpServer.Shutdown: %s", err))
	}

}
