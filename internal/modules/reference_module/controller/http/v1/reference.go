package v1

import (
	"context"
	"net/http"

	"github.com/k1v4/drip_mate/internal/entity"
	_ "github.com/k1v4/drip_mate/internal/swagger"
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

	g := handler.Group("/references")
	g.GET("/styles", h.GetStyles)
	g.GET("/colors", h.GetColors)
	g.GET("/musics", h.GetMusics)
	g.GET("/categories", h.GetCategories)
	g.GET("/seasons", h.GetSeasons)
}

// GetStyles godoc
// @Summary      Get styles
// @Description  Returns list of available styles
// @Tags         reference
// @Produce      json
// @Success      200  {array}   entity.StyleType
// @Failure      500  {object}  swagger.ErrorResponse
// @Router       /references/styles [get]
func (h *referencesRoutes) GetStyles(c echo.Context) error {
	styles, err := h.uc.GetStyles(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err).SetInternal(err)
	}
	return c.JSON(http.StatusOK, styles)
}

// GetColors godoc
// @Summary      Get colors
// @Description  Returns list of available colors
// @Tags         reference
// @Produce      json
// @Success      200  {array}   entity.ColorType
// @Failure      500  {object}  swagger.ErrorResponse
// @Router       /references/colors [get]
func (h *referencesRoutes) GetColors(c echo.Context) error {
	colors, err := h.uc.GetColors(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err).SetInternal(err)
	}
	return c.JSON(http.StatusOK, colors)
}

// GetMusics godoc
// @Summary      Get music types
// @Description  Returns list of available music preferences
// @Tags         reference
// @Produce      json
// @Success      200  {array}   entity.MusicType
// @Failure      500  {object}  swagger.ErrorResponse
// @Router       /references/musics [get]
func (h *referencesRoutes) GetMusics(c echo.Context) error {
	musics, err := h.uc.GetMusics(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err).SetInternal(err)
	}
	return c.JSON(http.StatusOK, musics)
}

// GetCategories godoc
// @Summary      Get categories
// @Description  Returns list of clothing categories
// @Tags         reference
// @Produce      json
// @Success      200  {array}   entity.Category
// @Failure      500  {object}  swagger.ErrorResponse
// @Router       /references/categories [get]
func (h *referencesRoutes) GetCategories(c echo.Context) error {
	categories, err := h.uc.GetCategories(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err).SetInternal(err)
	}
	return c.JSON(http.StatusOK, categories)
}

// GetSeasons godoc
// @Summary      Get seasons
// @Description  Returns list of seasons
// @Tags         reference
// @Produce      json
// @Success      200  {array}   entity.Season
// @Failure      500  {object}  swagger.ErrorResponse
// @Router       /references/seasons [get]
func (h *referencesRoutes) GetSeasons(c echo.Context) error {
	seasons, err := h.uc.GetSeasons(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err).SetInternal(err)
	}
	return c.JSON(http.StatusOK, seasons)
}
