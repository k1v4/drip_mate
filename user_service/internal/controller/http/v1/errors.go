package v1

import (
	"errors"
	"user_service/internal/entity"

	"github.com/labstack/echo/v4"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

func errorResponse(c echo.Context, code int, msg string) error {
	return c.JSON(code, entity.ErrorResponse{Error: msg})
}
