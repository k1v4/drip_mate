package v1

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/k1v4/drip_mate/internal/config"
	"github.com/k1v4/drip_mate/internal/modules/user_service/entity"
	"github.com/k1v4/drip_mate/internal/modules/user_service/usecase"
	middlewareJWT "github.com/k1v4/drip_mate/internal/router/middleware"
	"github.com/k1v4/drip_mate/pkg/DataBase"
	"github.com/k1v4/drip_mate/pkg/logger"
	"github.com/labstack/echo/v4"
)

type containerRoutes struct {
	t   usecase.ISsoService
	l   logger.Logger
	cfg *config.Token
}

func NewSsoRoutes(handler *echo.Group, t usecase.ISsoService, l logger.Logger, cfg *config.Token) {
	r := &containerRoutes{t, l, cfg}

	// POST /api/v1/login
	handler.POST("/login", r.Auth)

	// POST /api/v1/register
	handler.POST("/register", r.Register)

	// DELETE  /api/v1/users
	handler.DELETE("/users", r.DeleteAccount, middlewareJWT.JWTAuth(cfg))

	// POST  /api/v1/users/outfit
	handler.POST("/users/outfit", r.SaveOutfit, middlewareJWT.JWTAuth(cfg))

	// GET  /api/v1/users/outfit
	handler.GET("/users/outfit", r.GetOutfits, middlewareJWT.JWTAuth(cfg))

	// GET  /api/v1/users
	handler.GET("/users", r.GetUserByID, middlewareJWT.JWTAuth(cfg))

	// DELETE  /api/v1/users/outfit
	handler.DELETE("/users/outfit/:id", r.DeleteOutfit, middlewareJWT.JWTAuth(cfg))

	// POST  /api/v1/auth/change-password
	handler.POST("/auth/change-password", r.PassChange, middlewareJWT.JWTAuth(cfg))

	// PATCH  /api/v1/me/profile
	handler.PATCH("/me/profile", r.UpdateUserInfo, middlewareJWT.JWTAuth(cfg))

	// PATCH  /api/v1/me/context
	handler.PATCH("/me/context", r.UpdateUserContext, middlewareJWT.JWTAuth(cfg))
}

func (r *containerRoutes) Auth(c echo.Context) error {
	const op = "controller.Auth"
	ctx := c.Request().Context()

	u := new(entity.LoginRequest)
	if err := c.Bind(u); err != nil {
		r.l.Error(ctx, fmt.Sprintf("%s: %v", op, err))
		return echo.NewHTTPError(http.StatusBadRequest, "bad request").SetInternal(err)
	}

	if len(u.Password) == 0 || len([]rune(u.Email)) == 0 {
		r.l.Error(ctx, fmt.Sprintf("%s: invalid params", op))
		return echo.NewHTTPError(http.StatusBadRequest, "bad request").SetInternal(errors.New("пропущено поле"))
	}

	accessID, accessToken, err := r.t.Login(ctx, u.Email, u.Password)
	if err != nil {
		r.l.Error(ctx, fmt.Sprintf("%s: %v", op, err))

		if errors.Is(err, usecase.ErrInvalidCredentials) {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials").SetInternal(err)
		}
		if errors.Is(err, usecase.ErrNoUser) {
			return echo.NewHTTPError(http.StatusUnauthorized, "no user").SetInternal(err)
		}

		return echo.NewHTTPError(http.StatusInternalServerError, "internal error").SetInternal(err)
	}

	c.SetCookie(&http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   int(r.cfg.TTL.Seconds()),
	})

	return c.JSON(http.StatusOK, entity.LoginResponse{
		AccessId: accessID,
	})
}

func (r *containerRoutes) Register(c echo.Context) error {
	const op = "controller.Register"

	ctx := c.Request().Context()

	u := new(entity.RegisterRequest)
	if err := c.Bind(u); err != nil {
		r.l.Error(ctx, fmt.Sprintf("%s: %v", op, err))
		return echo.NewHTTPError(http.StatusBadRequest, "bad request").SetInternal(err)
	}

	if err := c.Validate(u); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "bad request").SetInternal(err)
	}

	accessID, accessToken, err := r.t.Register(ctx, u.Email, u.Password)
	if err != nil {
		r.l.Error(ctx, fmt.Sprintf("%s: %v", op, err))

		if errors.Is(err, usecase.ErrUserExist) {
			return echo.NewHTTPError(http.StatusConflict, "email or username is exist").SetInternal(err)
		}

		return echo.NewHTTPError(http.StatusInternalServerError, "internal error").SetInternal(err)
	}

	c.SetCookie(&http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   int(r.cfg.TTL.Seconds()),
	})

	return c.JSON(http.StatusOK, entity.LoginResponse{
		AccessId: accessID,
	})
}

func (r *containerRoutes) UpdateUserInfo(c echo.Context) error {
	const op = "controller.UpdateUserInfo"

	ctx := c.Request().Context()

	userID := c.Get(middlewareJWT.UserIDKey).(string)

	u := new(entity.UpdatePersonal)
	if err := c.Bind(u); err != nil {
		r.l.Error(ctx, fmt.Sprintf("%s: %v", op, err))
		return echo.NewHTTPError(http.StatusBadRequest, "bad request").SetInternal(err)
	}

	user, err := r.t.UpdateUserInfo(ctx, userID, u.Name, u.Surname, u.Username, u.Gender)
	if err != nil {
		r.l.Error(ctx, fmt.Sprintf("%s: %v", op, err))
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error").SetInternal(err)
	}

	return c.JSON(http.StatusOK, entity.UpdateUserResponse{
		User: *user,
	})
}

func (r *containerRoutes) GetUserByID(c echo.Context) error {
	ctx := c.Request().Context()
	userID := c.Get(middlewareJWT.UserIDKey).(string)
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid uuid format").SetInternal(err)
	}

	user, err := r.t.GetUserByID(ctx, userUUID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error").SetInternal(err)
	}

	return c.JSON(http.StatusOK, user)
}

func (r *containerRoutes) DeleteAccount(c echo.Context) error {
	const op = "controller.DeleteAccount"
	ctx := c.Request().Context()

	userID := c.Get(middlewareJWT.UserIDKey).(string)

	isSucceed, err := r.t.DeleteAccount(ctx, userID)
	if err != nil {
		r.l.Error(ctx, fmt.Sprintf("%s: %v", op, err))
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error").SetInternal(err)
	}

	c.SetCookie(&http.Cookie{
		Name:     "access_token",
		Value:    "",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   -1,
	})

	return c.JSON(http.StatusOK, entity.DeleteUserResponse{
		IsSuccessfully: isSucceed,
	})
}

func (r *containerRoutes) SaveOutfit(c echo.Context) error {
	const op = "controller.SaveOutfit"

	ctx := c.Request().Context()

	userID := c.Get(middlewareJWT.UserIDKey).(string)
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid uuid format").SetInternal(err)
	}

	u := new(entity.SaveOutfitRequest)
	if err = c.Bind(u); err != nil {
		r.l.Error(ctx, fmt.Sprintf("%s: %v", op, err))
		return echo.NewHTTPError(http.StatusBadRequest, "bad request").SetInternal(err)
	}

	if err = c.Validate(u); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "bad request").SetInternal(err)
	}

	outfitUUID, err := r.t.SaveOutfit(ctx, userUUID, *u)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to save outfit").SetInternal(err)
	}

	return c.JSON(http.StatusOK, map[string]any{
		"outfit_id": outfitUUID,
	})
}

func (r *containerRoutes) GetOutfits(c echo.Context) error {
	const op = "controller.GetOutfits"
	ctx := c.Request().Context()

	userID := c.Get(middlewareJWT.UserIDKey).(string)
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid uuid format").SetInternal(err)
	}

	outfits, err := r.t.GetOutfits(ctx, userUUID)
	if err != nil {
		r.l.Error(ctx, fmt.Sprintf("%s: %v", op, err))
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error").SetInternal(err)
	}

	return c.JSON(http.StatusOK, outfits)
}

func (r *containerRoutes) DeleteOutfit(c echo.Context) error {
	ctx := c.Request().Context()

	userID := c.Get(middlewareJWT.UserIDKey).(string)
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid uuid format").SetInternal(err)
	}

	idParam := c.Param("id")
	outfitUUID, err := uuid.Parse(idParam)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid uuid format").SetInternal(err)
	}

	err = r.t.DeleteOutfit(ctx, userUUID, outfitUUID)
	if err != nil {
		if errors.Is(err, DataBase.ErrOutfitNotFound) {
			return echo.NewHTTPError(http.StatusBadRequest, "outfit not found").SetInternal(err)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error").SetInternal(err)
	}

	return c.NoContent(http.StatusNoContent)
}

func (r *containerRoutes) PassChange(c echo.Context) error {
	const op = "Controller.PassChange"
	ctx := c.Request().Context()

	userID := c.Get(middlewareJWT.UserIDKey).(string)
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid uuid format").SetInternal(err)
	}

	u := new(entity.UpdatePass)
	if err := c.Bind(u); err != nil {
		r.l.Error(ctx, fmt.Sprintf("%s: %v", op, err))
		return echo.NewHTTPError(http.StatusBadRequest, "bad request").SetInternal(err)
	}

	if err = c.Validate(u); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "bad request").SetInternal(err)
	}

	if strings.EqualFold(u.NewPassword, u.CurrPassword) {
		err = errors.New("passwords must be different")
		return echo.NewHTTPError(http.StatusBadRequest, "passwords must be different").SetInternal(err)
	}

	err = r.t.UpdatePassword(ctx, userUUID, u)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidCredentials) {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials").SetInternal(err)
		}

		return echo.NewHTTPError(http.StatusBadRequest, "failed to update password").SetInternal(err)
	}

	return c.NoContent(http.StatusOK)
}

func (r *containerRoutes) UpdateUserContext(c echo.Context) error {
	const op = "controller.UpdateUserContext"
	ctx := c.Request().Context()

	userID := c.Get(middlewareJWT.UserIDKey).(string)
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid uuid format").SetInternal(err)
	}

	u := new(entity.UpdateContext)
	if err := c.Bind(u); err != nil {
		r.l.Error(ctx, fmt.Sprintf("%s: %v", op, err))
		return echo.NewHTTPError(http.StatusBadRequest, "bad request").SetInternal(err)
	}

	err = r.t.UpdateContext(ctx, userUUID, u)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update user").SetInternal(err)
	}

	return c.NoContent(http.StatusOK)
}
