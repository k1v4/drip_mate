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
	_ "github.com/k1v4/drip_mate/internal/swagger"
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

	handler.POST("/auth/login", r.Auth)
	handler.POST("/auth/change-password", r.PassChange, middlewareJWT.JWTAuth(cfg))

	handler.POST("/users/register", r.Register)

	handler.DELETE("/users", r.DeleteAccount, middlewareJWT.JWTAuth(cfg))
	handler.POST("/users/outfits", r.SaveOutfit, middlewareJWT.JWTAuth(cfg))
	handler.GET("/users/outfits", r.GetOutfits, middlewareJWT.JWTAuth(cfg))
	handler.GET("/users", r.GetUserByID, middlewareJWT.JWTAuth(cfg))
	handler.DELETE("/users/outfits/:id", r.DeleteOutfit, middlewareJWT.JWTAuth(cfg))
	handler.PATCH("/users/profile", r.UpdateUserInfo, middlewareJWT.JWTAuth(cfg))
	handler.PATCH("/users/context", r.UpdateUserContext, middlewareJWT.JWTAuth(cfg))
}

// Auth godoc
// @Summary      Login user
// @Description  Authenticates user by email and password, sets access_token cookie
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      entity.LoginRequest   true  "Login credentials"
// @Success      200   {object}  entity.LoginResponse
// @Failure      400   {object}  swagger.ErrorResponse
// @Failure      401   {object}  swagger.ErrorResponse
// @Failure      500   {object}  swagger.ErrorResponse
// @Router       /auth/login [post]
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
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   int(r.cfg.TTL.Seconds()),
	})

	return c.JSON(http.StatusOK, entity.LoginResponse{
		AccessId: accessID,
	})
}

// Register godoc
// @Summary      Register user
// @Description  Creates a new user account and sets access_token cookie
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      entity.RegisterRequest  true  "Registration data"
// @Success      200   {object}  entity.LoginResponse
// @Failure      400   {object}  swagger.ErrorResponse
// @Failure      409   {object}  swagger.ErrorResponse
// @Failure      500   {object}  swagger.ErrorResponse
// @Router       /users/register [post]
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
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   int(r.cfg.TTL.Seconds()),
	})

	return c.JSON(http.StatusOK, entity.LoginResponse{
		AccessId: accessID,
	})
}

// UpdateUserInfo godoc
// @Summary      Update personal info
// @Description  Updates user's name, surname, username and gender
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        body  body      entity.UpdatePersonal  true  "Personal info"
// @Success      200   {object}  entity.UpdateUserResponse
// @Failure      400   {object}  swagger.ErrorResponse
// @Failure      401   {object}  swagger.ErrorResponse
// @Failure      500   {object}  swagger.ErrorResponse
// @Security     CookieAuth
// @Router       /users/profile [patch]
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

// GetUserByID godoc
// @Summary      Get current user
// @Description  Returns profile of the authenticated user
// @Tags         users
// @Produce      json
// @Success      200  {object}  entity.User
// @Failure      400  {object}  swagger.ErrorResponse
// @Failure      401  {object}  swagger.ErrorResponse
// @Failure      500  {object}  swagger.ErrorResponse
// @Security     CookieAuth
// @Router       /users [get]
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

// DeleteAccount godoc
// @Summary      Delete account
// @Description  Permanently deletes the authenticated user's account and clears cookie
// @Tags         users
// @Produce      json
// @Success      200  {object}  entity.DeleteUserResponse
// @Failure      401  {object}  swagger.ErrorResponse
// @Failure      500  {object}  swagger.ErrorResponse
// @Security     CookieAuth
// @Router       /users [delete]
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
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   -1,
	})

	return c.JSON(http.StatusOK, entity.DeleteUserResponse{
		IsSuccessfully: isSucceed,
	})
}

// SaveOutfit godoc
// @Summary      Save outfit
// @Description  Saves a new outfit for the authenticated user
// @Tags         outfits
// @Accept       json
// @Produce      json
// @Param        body  body      entity.SaveOutfitRequest  true  "Outfit data"
// @Success      200   {object}  map[string]string
// @Failure      400   {object}  swagger.ErrorResponse
// @Failure      401   {object}  swagger.ErrorResponse
// @Failure      500   {object}  swagger.ErrorResponse
// @Security     CookieAuth
// @Router       /users/outfits [post]
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

// GetOutfits godoc
// @Summary      Get outfits
// @Description  Returns all saved outfits of the authenticated user
// @Tags         outfits
// @Produce      json
// @Success      200  {array}   entity.Outfit
// @Failure      400  {object}  swagger.ErrorResponse
// @Failure      401  {object}  swagger.ErrorResponse
// @Failure      500  {object}  swagger.ErrorResponse
// @Security     CookieAuth
// @Router       /users/outfits [get]
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

// DeleteOutfit godoc
// @Summary      Delete outfit
// @Description  Deletes a specific outfit by ID for the authenticated user
// @Tags         outfits
// @Produce      json
// @Param        id   path  string  true  "Outfit UUID"
// @Success      204
// @Failure      400  {object}  swagger.ErrorResponse
// @Failure      401  {object}  swagger.ErrorResponse
// @Failure      500  {object}  swagger.ErrorResponse
// @Security     CookieAuth
// @Router       /users/outfits/{id} [delete]
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

// PassChange godoc
// @Summary      Change password
// @Description  Updates the authenticated user's password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body  entity.UpdatePass  true  "Password update data"
// @Success      200
// @Failure      400  {object}  swagger.ErrorResponse
// @Failure      401  {object}  swagger.ErrorResponse
// @Security     CookieAuth
// @Router       /auth/change-password [post]
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

// UpdateUserContext godoc
// @Summary      Update user context
// @Description  Updates additional context/preferences for the authenticated user
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        body  body  entity.UpdateContext  true  "User context data"
// @Success      200
// @Failure      400  {object}  swagger.ErrorResponse
// @Failure      401  {object}  swagger.ErrorResponse
// @Failure      500  {object}  swagger.ErrorResponse
// @Security     CookieAuth
// @Router       /users/context [patch]
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
