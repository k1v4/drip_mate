package v1

import (
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/k1v4/drip_mate/internal/config"
	"github.com/k1v4/drip_mate/internal/entity"
	middlewareJWT "github.com/k1v4/drip_mate/internal/router/middleware"
	"github.com/k1v4/drip_mate/pkg/DataBase"
	"github.com/k1v4/drip_mate/pkg/logger"
	"github.com/labstack/echo/v4"
)

type IClothingUseCase interface {
	GetItemByID(ctx context.Context, id uuid.UUID) (*entity.Catalog, error)
	DeleteItem(ctx context.Context, id uuid.UUID) error
	UpdateItem(ctx context.Context, req *entity.UpdateCatalogRequest, fileName string, imageData []byte) (*entity.Catalog, error)
	CreateItem(ctx context.Context, req *entity.CreateCatalogRequest, fileName string, imageData []byte) (*entity.Catalog, error)
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
	ctx := c.Request().Context()

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid uuid format").SetInternal(err)
	}

	item, err := r.t.GetItemByID(ctx, id)
	if err != nil {
		if errors.Is(err, DataBase.ErrCatalogItemNotFound) {
			return echo.NewHTTPError(http.StatusBadRequest, "item not found").SetInternal(err)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get item").SetInternal(err)
	}

	return c.JSON(http.StatusOK, item)
}

func (r *containerRoutes) DeleteItem(c echo.Context) error {
	ctx := c.Request().Context()

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid uuid format").SetInternal(err)
	}

	err = r.t.DeleteItem(ctx, id)
	if err != nil {
		if errors.Is(err, DataBase.ErrCatalogItemNotFound) {
			return echo.NewHTTPError(http.StatusBadRequest, "item not found").SetInternal(err)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete item").SetInternal(err)
	}

	return c.NoContent(http.StatusNoContent)
}

func (r *containerRoutes) UpdateItem(c echo.Context) error {
	ctx := c.Request().Context()

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid uuid format").SetInternal(err)
	}

	u := new(entity.UpdateCatalogRequest)
	if err = c.Bind(u); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body").SetInternal(err)
	}

	if err = c.Validate(u); err != nil {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "missing required field").SetInternal(err)
	}
	u.ID = id

	// опционально достаём файл
	var fileName string
	var imageData []byte

	file, err := c.FormFile("image")
	if err == nil { // файл пришёл
		src, err := file.Open()
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to open image").SetInternal(err)
		}
		defer func() {
			_ = src.Close()
		}()

		imageData, err = io.ReadAll(src)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to read image").SetInternal(err)
		}
		fileName = file.Filename
	}

	item, err := r.t.UpdateItem(ctx, u, fileName, imageData)
	if err != nil {
		if errors.Is(err, DataBase.ErrCatalogItemNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "item not found").SetInternal(err)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update item").SetInternal(err)
	}

	return c.JSON(http.StatusOK, item)
}

func (r *containerRoutes) CreateItem(c echo.Context) error {
	ctx := c.Request().Context()

	u := new(entity.CreateCatalogRequest)
	if err := c.Bind(u); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body").SetInternal(err)
	}

	if err := c.Validate(u); err != nil {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "missing required field").SetInternal(err)
	}

	file, err := c.FormFile("image")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "image is required").SetInternal(err)
	}

	src, err := file.Open()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to open image").SetInternal(err)
	}
	defer func() {
		_ = src.Close()
	}()

	imageData, err := io.ReadAll(src)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to read image").SetInternal(err)
	}

	itemUUID, err := r.t.CreateItem(ctx, u, file.Filename, imageData)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create item").SetInternal(err)
	}

	return c.JSON(http.StatusCreated, itemUUID)
}
