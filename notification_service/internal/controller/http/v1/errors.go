package v1

import (
	"errors"
	"notification_service/internal/entity"

	"github.com/labstack/echo/v4"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

//nolint:unused
func errorResponse(c echo.Context, code int, msg string) error {
	return c.JSON(code, entity.ErrorResponse{Error: msg})
}
