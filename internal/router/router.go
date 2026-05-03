package router

import (
	"net/http"

	"github.com/k1v4/drip_mate/internal/config"
	v1catalog "github.com/k1v4/drip_mate/internal/modules/clothing_catalog/controller/http/v1"
	v1recommendation "github.com/k1v4/drip_mate/internal/modules/recommendation_core_module/controller/http/v1"
	v1references "github.com/k1v4/drip_mate/internal/modules/reference_module/controller/http/v1"
	v1user "github.com/k1v4/drip_mate/internal/modules/user_service/controller/http/v1"
	"github.com/k1v4/drip_mate/internal/modules/user_service/usecase"
	"github.com/k1v4/drip_mate/pkg/logger"
	echoSwagger "github.com/swaggo/echo-swagger"

	_ "github.com/k1v4/drip_mate/docs"
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
	rec v1recommendation.IRecommendationUseCase,
) {
	// Middleware
	handler.Use(middleware.RequestLogger())
	handler.Use(middleware.Recover())

	handler.GET("/api/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})
	token := new(cfg.Token)
	handler.GET("/api/swagger/*", echoSwagger.WrapHandler)

	h := handler.Group("/api/v1")
	{
		v1user.NewSsoRoutes(h, t, l, token)
		v1catalog.NewCatalogRoutes(h, cl, l, token)
		v1references.NewReferencesRoutes(h, ref, l)
		v1recommendation.NewRecommendationRoutes(h, rec, l, token)
	}
}
