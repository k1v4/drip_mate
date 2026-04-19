package v1

import (
	"context"
	"net/http"

	"github.com/k1v4/drip_mate/internal/entity"
	"github.com/k1v4/drip_mate/pkg/logger"
	"github.com/labstack/echo/v4"
)

type IReferenceUseCase interface {
	GetStyles(ctx context.Context) ([]entity.StyleType, error)
	GetColors(ctx context.Context) ([]entity.ColorType, error)
	GetMusics(ctx context.Context) ([]entity.MusicType, error)
	GetCategories(ctx context.Context) ([]entity.Category, error)
	GetSeasons(ctx context.Context) ([]entity.Season, error)
}

type referencesRoutes struct {
	uc IReferenceUseCase
	l  logger.Logger
}

func NewReferencesRoutes(handler *echo.Group, t IReferenceUseCase, l logger.Logger) {
	h := &referencesRoutes{t, l}

	g := handler.Group("/reference")
	g.GET("/styles", h.GetStyles)
	g.GET("/colors", h.GetColors)
	g.GET("/musics", h.GetMusics)
	g.GET("/categories", h.GetCategories)
	g.GET("/seasons", h.GetSeasons)
}

func (h *referencesRoutes) GetStyles(c echo.Context) error {
	styles, err := h.uc.GetStyles(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err).SetInternal(err)
	}
	return c.JSON(http.StatusOK, styles)
}

func (h *referencesRoutes) GetColors(c echo.Context) error {
	colors, err := h.uc.GetColors(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err).SetInternal(err)
	}
	return c.JSON(http.StatusOK, colors)
}

func (h *referencesRoutes) GetMusics(c echo.Context) error {
	musics, err := h.uc.GetMusics(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err).SetInternal(err)
	}
	return c.JSON(http.StatusOK, musics)
}

func (h *referencesRoutes) GetCategories(c echo.Context) error {
	categories, err := h.uc.GetCategories(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err).SetInternal(err)
	}
	return c.JSON(http.StatusOK, categories)
}

func (h *referencesRoutes) GetSeasons(c echo.Context) error {
	seasons, err := h.uc.GetSeasons(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err).SetInternal(err)
	}
	return c.JSON(http.StatusOK, seasons)
}
