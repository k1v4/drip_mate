package jwtpkg

import (
	"errors"
	"fmt"
	"time"

	totalEntity "github.com/k1v4/drip_mate/internal/entity"
	"github.com/k1v4/drip_mate/internal/modules/user_service/entity"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func ExtractToken(c echo.Context) string {
	cookie, err := c.Cookie("access_token")
	if err != nil {
		return ""
	}

	return cookie.Value
}

func NewAccessToken(user *entity.User, duration time.Duration, secret, issuer string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)

	claims["iss"] = issuer
	claims["id"] = user.ID
	claims["access_level_id"] = user.AccessLevelId
	claims["exp"] = time.Now().Add(duration).Unix()

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ValidateTokenAndGetUserId(tokenString, secret, issuer string) (string, totalEntity.Role, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	}, jwt.WithIssuer(issuer))
	if err != nil {
		return "", 0, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID, ok := claims["id"].(string)
		if !ok {
			return "", 0, fmt.Errorf("user_id not found in token")
		}

		if exp, ok := claims["exp"].(float64); ok {
			if time.Now().Unix() > int64(exp) {
				return "", 0, fmt.Errorf("token expired")
			}
		}

		// access_level_id в JWT хранится как float64
		accessLevel, ok := claims["access_level_id"].(float64)
		if !ok {
			return "", 0, fmt.Errorf("access_level_id not found in token")
		}

		return userID, totalEntity.Role(accessLevel), nil
	}

	return "", 0, errors.New("invalid token")
}
