package middleware

import (
	"net/http"

	"github.com/k1v4/drip_mate/internal/config"
	"github.com/k1v4/drip_mate/internal/entity"
	jwtPkg "github.com/k1v4/drip_mate/pkg/jwtpkg"
	"github.com/labstack/echo/v4"
)

const (
	UserIDKey      = "user_id"
	AccessLevelKey = "access_level"
)

func JWTAuth(cfg *config.Token) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token := jwtPkg.ExtractToken(c)
			if token == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "token is required")
			}

			userID, err := jwtPkg.ValidateTokenAndGetUserId(token, cfg.Secret, cfg.Issuer)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token").SetInternal(err)
			}

			c.Set(UserIDKey, userID)

			return next(c)
		}
	}
}

func AdminOnly() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// достаем уровень доступа
			level, ok := c.Get(AccessLevelKey).(entity.Role)

			if !ok || level != entity.RoleAdmin {
				return echo.NewHTTPError(http.StatusForbidden, "access denied: admin rights required")
			}

			return next(c)
		}
	}
}
