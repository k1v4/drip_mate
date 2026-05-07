package jwtpkg

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	totalEntity "github.com/k1v4/drip_mate/internal/entity"
	userEntity "github.com/k1v4/drip_mate/internal/modules/user_service/entity"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

const (
	testSecret = "super-secret"
	testIssuer = "drip-mate"
)

func TestExtractToken_NoCookie(t *testing.T) {
	e := echo.New()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	token := ExtractToken(c)

	if token != "" {
		t.Errorf("expected empty token, got '%s'", token)
	}
}

func TestNewAccessToken_AndValidate(t *testing.T) {
	user := &userEntity.User{
		ID:       uuid.New(),
		AccessID: int(totalEntity.RoleAdmin),
	}

	token, err := NewAccessToken(
		user,
		time.Hour,
		testSecret,
		testIssuer,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	userID, role, err := ValidateTokenAndGetUserId(
		token,
		testSecret,
		testIssuer,
	)
	if err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}

	if !strings.EqualFold(userID, user.ID.String()) {
		t.Errorf("expected userID %s, got %s", user.ID, userID)
	}

	if role != totalEntity.RoleAdmin {
		t.Errorf("expected role %v, got %v", totalEntity.RoleAdmin, role)
	}
}

func TestValidateTokenAndGetUserId_InvalidSignature(t *testing.T) {
	user := &userEntity.User{
		ID:       uuid.New(),
		AccessID: 2,
	}

	token, err := NewAccessToken(
		user,
		time.Hour,
		testSecret,
		testIssuer,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, _, err = ValidateTokenAndGetUserId(
		token,
		"wrong-secret",
		testIssuer,
	)

	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestValidateTokenAndGetUserId_WrongIssuer(t *testing.T) {
	user := &userEntity.User{
		ID:       uuid.New(),
		AccessID: 2,
	}

	token, err := NewAccessToken(
		user,
		time.Hour,
		testSecret,
		testIssuer,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, _, err = ValidateTokenAndGetUserId(
		token,
		testSecret,
		"wrong-issuer",
	)

	if err == nil {
		t.Fatal("expected issuer validation error")
	}
}

func TestValidateTokenAndGetUserId_ExpiredToken(t *testing.T) {
	user := &userEntity.User{
		ID:       uuid.New(),
		AccessID: 2,
	}

	token, err := NewAccessToken(
		user,
		-time.Hour,
		testSecret,
		testIssuer,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, _, err = ValidateTokenAndGetUserId(
		token,
		testSecret,
		testIssuer,
	)

	if err == nil {
		t.Fatal("expected expired token error")
	}
}

func TestValidateTokenAndGetUserId_InvalidToken(t *testing.T) {
	_, _, err := ValidateTokenAndGetUserId(
		"invalid-token",
		testSecret,
		testIssuer,
	)

	if err == nil {
		t.Fatal("expected invalid token error")
	}
}

func TestValidateTokenAndGetUserId_NoUserID(t *testing.T) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)

	claims["iss"] = testIssuer
	claims["access_level_id"] = 2
	claims["exp"] = time.Now().Add(time.Hour).Unix()

	tokenString, err := token.SignedString([]byte(testSecret))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	_, _, err = ValidateTokenAndGetUserId(
		tokenString,
		testSecret,
		testIssuer,
	)

	if err == nil {
		t.Fatal("expected missing user id error")
	}
}

func TestValidateTokenAndGetUserId_NoAccessLevel(t *testing.T) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)

	claims["iss"] = testIssuer
	claims["id"] = uuid.New().String()
	claims["exp"] = time.Now().Add(time.Hour).Unix()

	tokenString, err := token.SignedString([]byte(testSecret))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	_, _, err = ValidateTokenAndGetUserId(
		tokenString,
		testSecret,
		testIssuer,
	)

	if err == nil {
		t.Fatal("expected missing access level error")
	}
}

func TestValidateTokenAndGetUserId_InvalidSigningMethod(t *testing.T) {
	token := jwt.New(jwt.SigningMethodNone)

	claims := token.Claims.(jwt.MapClaims)

	claims["iss"] = testIssuer
	claims["id"] = uuid.New().String()
	claims["access_level_id"] = 1
	claims["exp"] = time.Now().Add(time.Hour).Unix()

	tokenString, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	_, _, err = ValidateTokenAndGetUserId(
		tokenString,
		testSecret,
		testIssuer,
	)

	if err == nil {
		t.Fatal("expected invalid signing method error")
	}
}
