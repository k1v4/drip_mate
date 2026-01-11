package v1

import (
	"net/http"

	"user_service/internal/usecase"
	"user_service/pkg/logger"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func NewRouter(handler *echo.Echo, l logger.Logger, t usecase.ISsoService) {
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
		newSsoRoutes(h, t, l)
	}
}
