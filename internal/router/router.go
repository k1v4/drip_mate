package router

import (
	"net/http"

	"github.com/k1v4/drip_mate/internal/config"
	v1catalog "github.com/k1v4/drip_mate/internal/modules/clothing_catalog/controller/http/v1"
	v1references "github.com/k1v4/drip_mate/internal/modules/reference_module/controller/http/v1"
	v1user "github.com/k1v4/drip_mate/internal/modules/user_service/controller/http/v1"
	"github.com/k1v4/drip_mate/internal/modules/user_service/usecase"
	"github.com/k1v4/drip_mate/pkg/logger"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func NewRouter(
	handler *echo.Echo,
	l logger.Logger,
	t usecase.ISsoService,
	cfg *config.Config,
	cl v1catalog.IClothingUseCase,
	ref v1references.IReferenceUseCase,
) {
	// Middleware
	handler.Use(middleware.RequestLogger())
	handler.Use(middleware.Recover())
	handler.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"http://localhost:3000"},                                                                // Разрешить запросы с этого origin
		AllowMethods:     []string{echo.GET, echo.PUT, echo.POST, echo.DELETE, echo.OPTIONS},                               // Разрешенные методы
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization}, // Разрешенные заголовки
		AllowCredentials: true,                                                                                             // Разрешить передачу кук и заголовков авторизации
	}))

	handler.GET("/api/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	h := handler.Group("/api/v1")
	{
		v1user.NewSsoRoutes(h, t, l, new(cfg.Token))
		v1catalog.NewCatalogRoutes(h, cl, l, new(cfg.Token))
		v1references.NewReferencesRoutes(h, ref, l)
	}
}
