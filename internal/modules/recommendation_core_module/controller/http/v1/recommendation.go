package v1

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/k1v4/drip_mate/internal/config"
	"github.com/k1v4/drip_mate/internal/entity"
	middlewareJWT "github.com/k1v4/drip_mate/internal/router/middleware"
	_ "github.com/k1v4/drip_mate/internal/swagger"
	"github.com/k1v4/drip_mate/pkg/logger"
	"github.com/labstack/echo/v4"
)

type IRecommendationUseCase interface {
	GetUserRecommendation(ctx context.Context, formality int, userID uuid.UUID) (*entity.RecommendationsCatalogRequest, error)
}

type recommendationsRoutes struct {
	t   IRecommendationUseCase
	l   logger.Logger
	cfg *config.Token
}

func NewRecommendationRoutes(handler *echo.Group, t IRecommendationUseCase, l logger.Logger, cfg *config.Token) {
	r := &recommendationsRoutes{t, l, cfg}

	handler.POST("/recommendation", r.GetRecommendation, middlewareJWT.JWTAuth(cfg))
}

// GetRecommendation godoc
// @Summary      Get user recommendation
// @Description  Returns personalized catalog recommendations based on user formality level
// @Tags         recommendation
// @Accept       json
// @Produce      json
// @Param        body  body      entity.RecommendationRequest  true  "Recommendation request"
// @Success      200   {object}  entity.RecommendationsCatalogRequest
// @Failure      400   {object}  swagger.ErrorResponse
// @Failure      401   {object}  swagger.ErrorResponse
// @Failure      422   {object}  swagger.ErrorResponse
// @Failure      500   {object}  swagger.ErrorResponse
// @Security     CookieAuth
// @Router       /recommendation [post]
func (r *recommendationsRoutes) GetRecommendation(c echo.Context) error {
	ctx := c.Request().Context()

	userIDStr, ok := c.Get(middlewareJWT.UserIDKey).(string)
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid token").
			SetInternal(errors.New("invalid token"))
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid token").
			SetInternal(errors.New("invalid token"))
	}

	u := new(entity.RecommendationRequest)
	if err = c.Bind(u); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body").
			SetInternal(err)
	}

	if err = c.Validate(u); err != nil {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "missing required field").
			SetInternal(err)
	}

	recommendations, err := r.t.GetUserRecommendation(ctx, u.Formality, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get recommendations").
			SetInternal(err)
	}

	return c.JSON(http.StatusOK, recommendations)
}
