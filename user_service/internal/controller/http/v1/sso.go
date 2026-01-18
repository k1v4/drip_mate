package v1

import (
	"errors"
	"fmt"
	"net/http"
	"user_service/internal/entity"
	"user_service/internal/usecase"
	jwtPkg "user_service/pkg/jwtpkg"
	"user_service/pkg/logger"

	"github.com/labstack/echo/v4"
)

type containerRoutes struct {
	t usecase.ISsoService
	l logger.Logger
}

func newSsoRoutes(handler *echo.Group, t usecase.ISsoService, l logger.Logger) {
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
		errorResponse(c, http.StatusBadRequest, "bad request")
		r.l.Error(ctx, fmt.Sprintf("%s: %v", op, err))
		return fmt.Errorf("%s: %w", op, err)
	}

	if len(u.Password) == 0 || len([]rune(u.Email)) == 0 {
		errorResponse(c, http.StatusBadRequest, "bad request")

		r.l.Error(ctx, fmt.Sprintf("%s: invalid params", op))
		return fmt.Errorf("%s: %w", op, errors.New("bad request"))
	}

	accessID, accessToken, refreshToken, err := r.t.Login(ctx, u.Email, u.Password)
	if err != nil {
		r.l.Error(ctx, fmt.Sprintf("%s: %v", op, err))

		if errors.Is(err, usecase.ErrInvalidCredentials) {
			errorResponse(c, http.StatusUnauthorized, "invalid credentials")

			return fmt.Errorf("%s: %w", op, err)
		}

		if errors.Is(err, usecase.ErrNoUser) {
			errorResponse(c, http.StatusUnauthorized, "no user")

			return fmt.Errorf("%s: %w", op, err)
		}

		errorResponse(c, http.StatusInternalServerError, "internal error")

		return fmt.Errorf("%s: %w", op, err)
	}

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
		errorResponse(c, http.StatusBadRequest, "bad request")

		r.l.Error(ctx, fmt.Sprintf("%s: %v", op, err))
		return fmt.Errorf("%s: %w", op, err)
	}

	if len([]rune(u.Password)) < 10 {
		errorResponse(c, http.StatusBadRequest, "password must be equal or longer than 10")

		r.l.Error(ctx, fmt.Sprintf("%s: invalid password", op))
		return fmt.Errorf("%s: %w", op, errors.New("password must be equal or longer than 10"))
	}

	if len([]rune(u.Email)) == 0 {
		errorResponse(c, http.StatusBadRequest, "email is required")

		r.l.Error(ctx, fmt.Sprintf("%s: invalid email", op))
		return fmt.Errorf("%s: %w", op, errors.New("email is required"))
	}

	register, err := r.t.Register(ctx, u.Email, u.Password)
	if err != nil {
		r.l.Error(ctx, fmt.Sprintf("%s: %v", op, err))

		if errors.Is(err, usecase.ErrUserExist) {
			errorResponse(c, http.StatusUnauthorized, "email or username is exist")

			return fmt.Errorf("%s: %w", op, err)
		}

		errorResponse(c, http.StatusInternalServerError, "internal error")

		return fmt.Errorf("%s: %w", op, err)
	}

	// TODO подумать про добавление автоматической авторизации
	return c.JSON(http.StatusOK, map[string]interface{}{
		"user_id": register,
	})
}

func (r *containerRoutes) UpdateUserInfo(c echo.Context) error {
	const op = "controller.UpdateUserInfo"

	ctx := c.Request().Context()

	// достаём access token
	token := jwtPkg.ExtractToken(c)
	if token == "" {
		errorResponse(c, http.StatusUnauthorized, "token is required")
		r.l.Error(ctx, fmt.Sprintf("%s: no token", op))
		return fmt.Errorf("%s: %s", op, "token is required")
	}

	// валидируем токен и достаём id пользователя
	userId, err := jwtPkg.ValidateTokenAndGetUserId(token)
	if err != nil {
		errorResponse(c, http.StatusUnauthorized, "wrong token")
		r.l.Error(ctx, fmt.Sprintf("%s: %v", op, err))
		return fmt.Errorf("%s: %s", op, err)
	}

	u := new(entity.UpdateUserRequest)
	if err = c.Bind(u); err != nil {
		errorResponse(c, http.StatusBadRequest, "bad request")
		r.l.Error(ctx, fmt.Sprintf("%s: %v", op, err))
		return fmt.Errorf("%s: %w", op, err)
	}

	if len(u.Password) < 10 && len(u.Password) > 0 {
		errorResponse(c, http.StatusBadRequest, "bad request")
		r.l.Error(ctx, fmt.Sprintf("%s: invalid password", op))
		return fmt.Errorf("%s: %w", op, errors.New("bad request"))
	}

	user, err := r.t.UpdateUserInfo(ctx, userId, u.Email, u.Password, u.Name, u.Surname, u.Username, u.City)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "internal error")
		r.l.Error(ctx, fmt.Sprintf("%s: %v", op, err))
		return fmt.Errorf("%s: %w", op, err)
	}

	return c.JSON(http.StatusOK, entity.UpdateUserResponse{
		User: user,
	})
}

func (r *containerRoutes) DeleteAccount(c echo.Context) error {
	const op = "controller.DeleteAccount"

	// достаём access token
	token := jwtPkg.ExtractToken(c)
	if token == "" {
		errorResponse(c, http.StatusBadRequest, "bad request")

		return fmt.Errorf("%s: %s", op, "token is required")
	}

	// валидируем токен и достаём id пользователя
	userId, err := jwtPkg.ValidateTokenAndGetUserId(token)
	if err != nil {
		errorResponse(c, http.StatusUnauthorized, "bad request")

		return fmt.Errorf("%s: %s", op, err)
	}

	ctx := c.Request().Context()

	isSucceed, err := r.t.DeleteAccount(ctx, userId)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "internal error")
		r.l.Error(ctx, fmt.Sprintf("%s: %v", op, err))
		return fmt.Errorf("%s: %w", op, err)
	}

	return c.JSON(http.StatusOK, entity.DeleteUserResponse{
		IsSuccessfully: isSucceed,
	})
}

func (r *containerRoutes) RefreshToken(c echo.Context) error {
	const op = "controller.RefreshToken"

	refreshTokenOld := jwtPkg.ExtractToken(c)
	ctx := c.Request().Context()

	accessToken, refreshToken, err := r.t.RefreshToken(ctx, refreshTokenOld)
	if err != nil {
		errorResponse(c, http.StatusUnauthorized, "token error")

		return fmt.Errorf("%s: %w", op, err)
	}

	return c.JSON(http.StatusOK, entity.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}
