package v1

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/k1v4/drip_mate/internal/config"
	"github.com/k1v4/drip_mate/internal/modules/user_service/entity"
	"github.com/k1v4/drip_mate/internal/modules/user_service/usecase"
	middlewareJWT "github.com/k1v4/drip_mate/internal/router/middleware"
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

	// PUT /api/v1/users
	handler.PUT("/users", r.UpdateUserInfo, middlewareJWT.JWTAuth(cfg))

	// DELETE  /api/v1/users
	handler.DELETE("/users", r.DeleteAccount, middlewareJWT.JWTAuth(cfg))
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

	register, accessToken, err := r.t.Register(ctx, u.Email, u.Password)
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

	return c.JSON(http.StatusOK, map[string]any{
		"user_id": register,
	})
}

func (r *containerRoutes) UpdateUserInfo(c echo.Context) error {
	const op = "controller.UpdateUserInfo"

	ctx := c.Request().Context()

	userID := c.Get(middlewareJWT.UserIDKey).(string)

	u := new(entity.UpdateUserRequest)
	if err := c.Bind(u); err != nil {
		r.l.Error(ctx, fmt.Sprintf("%s: %v", op, err))
		return echo.NewHTTPError(http.StatusBadRequest, "bad request").SetInternal(err)
	}

	if len(u.Password) < 10 && len(u.Password) > 0 {
		err := errors.New("bad request")
		r.l.Error(ctx, fmt.Sprintf("%s: invalid password", op))
		return echo.NewHTTPError(http.StatusBadRequest, "bad request").SetInternal(err)
	}

	user, err := r.t.UpdateUserInfo(ctx, userID, u.Email, u.Password, u.Name, u.Surname, u.Username, u.City)
	if err != nil {
		r.l.Error(ctx, fmt.Sprintf("%s: %v", op, err))
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error").SetInternal(err)
	}

	return c.JSON(http.StatusOK, entity.UpdateUserResponse{
		User: user,
	})
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
