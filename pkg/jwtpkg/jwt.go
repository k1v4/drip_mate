package jwtpkg

import (
	"errors"
	"fmt"
	"time"

	"github.com/k1v4/drip_mate/internal/modules/user_service/entity"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func ExtractToken(c echo.Context) string {
	// сначала пробуем куку
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

func ValidateTokenAndGetUserId(tokenString, secret, issuer string) (string, error) {
	// парсим токен
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	}, jwt.WithIssuer(issuer))
	if err != nil {
		return "", err
	}

	// проверяем claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// извлекаем userId
		userId, okUser := claims["id"].(string)
		if !okUser {
			return "", fmt.Errorf("userId not found in token")
		}

		// проверяем срок действия токена
		if exp, ok := claims["exp"].(float64); ok {
			if time.Now().Unix() > int64(exp) {
				return "", fmt.Errorf("token expired")
			}
		}

		return userId, nil
	}

	return "", errors.New("invalid token")
}
