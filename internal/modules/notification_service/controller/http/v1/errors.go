package v1

import (
	"errors"

	"github.com/k1v4/drip_mate/internal/modules/notification_service/entity"

	"github.com/labstack/echo/v4"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

//nolint:unused
func errorResponse(c echo.Context, code int, msg string) error {
	return c.JSON(code, entity.ErrorResponse{Error: msg})
}
