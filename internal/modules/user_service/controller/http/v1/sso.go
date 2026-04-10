package v1

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/k1v4/drip_mate/internal/modules/user_service/entity"
	"github.com/k1v4/drip_mate/internal/modules/user_service/usecase"
	jwtPkg "github.com/k1v4/drip_mate/pkg/jwtpkg"
	"github.com/k1v4/drip_mate/pkg/logger"

	"github.com/labstack/echo/v4"
)

type containerRoutes struct {
	t usecase.ISsoService
	l logger.Logger
}

func NewSsoRoutes(handler *echo.Group, t usecase.ISsoService, l logger.Logger) {
	r := &containerRoutes{t, l}

	// POST /api/v1/login
	handler.POST("/login", r.Auth)

	// POST /api/v1/register
	handler.POST("/register", r.Register)

	// PUT /api/v1/users
	handler.PUT("/users", r.UpdateUserInfo)

	// DELETE  /api/v1/users
	handler.DELETE("/users", r.DeleteAccount)

	// POST /api/v1/refresh
	handler.POST("/refresh", r.RefreshToken)
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

	accessID, accessToken, refreshToken, err := r.t.Login(ctx, u.Email, u.Password)
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

	// TODO устанавливать стразу в куки
	return c.JSON(http.StatusOK, entity.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		AccessId:     accessID,
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

	if len([]rune(u.Password)) < 10 {
		err := errors.New("password must be equal or longer than 10")
		r.l.Error(ctx, fmt.Sprintf("%s: invalid password", op))
		return echo.NewHTTPError(http.StatusBadRequest, "password must be equal or longer than 10").SetInternal(err)
	}

	if len([]rune(u.Email)) == 0 {
		err := errors.New("email is required")
		r.l.Error(ctx, fmt.Sprintf("%s: invalid email", op))
		return echo.NewHTTPError(http.StatusBadRequest, "email is required").SetInternal(err)
	}

	register, err := r.t.Register(ctx, u.Email, u.Password)
	if err != nil {
		r.l.Error(ctx, fmt.Sprintf("%s: %v", op, err))

		if errors.Is(err, usecase.ErrUserExist) {
			return echo.NewHTTPError(http.StatusConflict, "email or username is exist").SetInternal(err)
		}

		return echo.NewHTTPError(http.StatusInternalServerError, "internal error").SetInternal(err)
	}

	return c.JSON(http.StatusOK, map[string]any{
		"user_id": register,
	})
}

func (r *containerRoutes) UpdateUserInfo(c echo.Context) error {
	const op = "controller.UpdateUserInfo"

	ctx := c.Request().Context()

	token := jwtPkg.ExtractToken(c)
	if token == "" {
		err := errors.New("token is required")
		r.l.Error(ctx, fmt.Sprintf("%s: no token", op))
		return echo.NewHTTPError(http.StatusUnauthorized, "token is required").SetInternal(err)
	}

	userId, err := jwtPkg.ValidateTokenAndGetUserId(token)
	if err != nil {
		r.l.Error(ctx, fmt.Sprintf("%s: %v", op, err))
		return echo.NewHTTPError(http.StatusUnauthorized, "wrong token").SetInternal(err)
	}

	u := new(entity.UpdateUserRequest)
	if err = c.Bind(u); err != nil {
		r.l.Error(ctx, fmt.Sprintf("%s: %v", op, err))
		return echo.NewHTTPError(http.StatusBadRequest, "bad request").SetInternal(err)
	}

	if len(u.Password) < 10 && len(u.Password) > 0 {
		err := errors.New("bad request")
		r.l.Error(ctx, fmt.Sprintf("%s: invalid password", op))
		return echo.NewHTTPError(http.StatusBadRequest, "bad request").SetInternal(err)
	}

	user, err := r.t.UpdateUserInfo(ctx, userId, u.Email, u.Password, u.Name, u.Surname, u.Username, u.City)
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

	token := jwtPkg.ExtractToken(c)
	if token == "" {
		err := errors.New("token is required")
		r.l.Error(ctx, fmt.Sprintf("%s: token is required", op))
		return echo.NewHTTPError(http.StatusUnauthorized, "token is required").SetInternal(err)
	}

	userId, err := jwtPkg.ValidateTokenAndGetUserId(token)
	if err != nil {
		r.l.Error(ctx, fmt.Sprintf("%s: %v", op, err))
		return echo.NewHTTPError(http.StatusUnauthorized, "wrong token").SetInternal(err)
	}

	isSucceed, err := r.t.DeleteAccount(ctx, userId)
	if err != nil {
		r.l.Error(ctx, fmt.Sprintf("%s: %v", op, err))
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error").SetInternal(err)
	}

	return c.JSON(http.StatusOK, entity.DeleteUserResponse{
		IsSuccessfully: isSucceed,
	})
}

func (r *containerRoutes) RefreshToken(c echo.Context) error {
	const op = "controller.RefreshToken"

	ctx := c.Request().Context()
	refreshTokenOld := jwtPkg.ExtractToken(c)

	accessToken, refreshToken, err := r.t.RefreshToken(ctx, refreshTokenOld)
	if err != nil {
		r.l.Error(ctx, fmt.Sprintf("%s: %v", op, err))
		return echo.NewHTTPError(http.StatusUnauthorized, "token error").SetInternal(err)
	}

	// TODO устанавливать стразу в куки
	return c.JSON(http.StatusOK, entity.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}
