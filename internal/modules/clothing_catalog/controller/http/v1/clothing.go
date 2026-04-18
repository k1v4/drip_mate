package v1

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/k1v4/drip_mate/internal/config"
	"github.com/k1v4/drip_mate/internal/entity"
	middlewareJWT "github.com/k1v4/drip_mate/internal/router/middleware"
	"github.com/k1v4/drip_mate/pkg/logger"
	"github.com/labstack/echo/v4"
)

type IClothingUseCase interface {
	GetItemByID(ctx context.Context, id uuid.UUID) (*entity.Catalog, error)
	DeleteItem(ctx context.Context, id uuid.UUID) error
	UpdateItem(ctx context.Context, item *entity.Catalog) (uuid.UUID, error)
	CreateItem(ctx context.Context, item *entity.Catalog) (uuid.UUID, error)
}

type containerRoutes struct {
	t   IClothingUseCase
	l   logger.Logger
	cfg *config.Token
}

func NewCatalogRoutes(handler *echo.Group, t IClothingUseCase, l logger.Logger, cfg *config.Token) {
	r := &containerRoutes{t, l, cfg}

	// Группа для каталога: /api/v1/catalog
	catalogGroup := handler.Group("/catalog")

	catalogGroup.GET("/:id", r.GetItem, middlewareJWT.JWTAuth(cfg))

	adminGroup := catalogGroup.Group("")
	adminGroup.Use(middlewareJWT.JWTAuth(cfg))
	adminGroup.Use(middlewareJWT.AdminOnly())

	adminGroup.POST("", r.CreateItem)
	adminGroup.PUT("/:id", r.UpdateItem)
	adminGroup.DELETE("/:id", r.DeleteItem)
}

func (r *containerRoutes) GetItem(c echo.Context) error {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid uuid format")
	}

	item, err := r.t.GetItemByID(c.Request().Context(), id)
	if err != nil {
		// Тут твоя логика обработки ошибок (например, проверка на sql.ErrNoRows)
		return err
	}

	return c.JSON(http.StatusOK, item)
}

func (r *containerRoutes) DeleteItem(c echo.Context) error {
	return c.NoContent(http.StatusNoContent)
}

func (r *containerRoutes) UpdateItem(c echo.Context) error {
	return c.NoContent(http.StatusNoContent)
}

func (r *containerRoutes) CreateItem(c echo.Context) error {
	return c.NoContent(http.StatusNoContent)
}
