package v1_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/k1v4/drip_mate/internal/config"
	"github.com/k1v4/drip_mate/internal/entity"
	v1 "github.com/k1v4/drip_mate/internal/modules/recommendation_core_module/controller/http/v1"
	userEntity "github.com/k1v4/drip_mate/internal/modules/user_service/entity"
	mockUC "github.com/k1v4/drip_mate/mocks/internal_/modules/recommendation_core_module/controller/http/v1"
	mockLogger "github.com/k1v4/drip_mate/mocks/pkg/logger"
	"github.com/k1v4/drip_mate/pkg/jwtpkg"
	appValidator "github.com/k1v4/drip_mate/pkg/validator"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func defaultCfg() *config.Token {
	return &config.Token{TTL: time.Hour, Secret: "test-secret", Issuer: "test-issuer"}
}

func newEcho() *echo.Echo {
	e := echo.New()
	e.Validator = appValidator.New()
	return e
}

func setupRouter(e *echo.Echo, uc *mockUC.IRecommendationUseCase, log *mockLogger.Logger) {
	g := e.Group("")
	v1.NewRecommendationRoutes(g, uc, log, defaultCfg())
}

func generateToken(t *testing.T, userID uuid.UUID) string {
	t.Helper()
	cfg := defaultCfg()
	token, err := jwtpkg.NewAccessToken(
		&userEntity.User{ID: userID, AccessID: 1},
		cfg.TTL, cfg.Secret, cfg.Issuer,
	)
	require.NoError(t, err)
	return token
}

func makeReq(method, path, body string, token *string) (*http.Request, *httptest.ResponseRecorder) {
	req := httptest.NewRequestWithContext(context.Background(), method, path, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	if token != nil {
		req.AddCookie(&http.Cookie{Name: "access_token", Value: *token})
	}
	return req, httptest.NewRecorder()
}

func TestRecommendationRoutes_GetRecommendation(t *testing.T) {
	userID := uuid.New()
	token := generateToken(t, userID)

	successResult := &entity.RecommendationsCatalogRequest{
		Catalog: []entity.Catalog{
			{ID: uuid.New(), Name: "Test jacket"},
		},
		LogID: 42,
	}

	tests := []struct {
		name           string
		body           string
		token          *string
		setupUC        func(u *mockUC.IRecommendationUseCase)
		wantStatus     int
		wantLogID      *int
		wantCatalogLen *int
	}{
		{
			name:  "success",
			body:  `{"formality":3}`,
			token: &token,
			setupUC: func(u *mockUC.IRecommendationUseCase) {
				u.On("GetUserRecommendation", mock.Anything, 3, userID).
					Return(successResult, nil)
			},
			wantStatus:     http.StatusOK,
			wantLogID:      func() *int { v := 42; return &v }(),
			wantCatalogLen: func() *int { v := 1; return &v }(),
		},
		{
			name:       "no token",
			body:       `{"formality":3}`,
			token:      nil,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "missing formality — validation error",
			body:       `{}`,
			token:      &token,
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "formality below min=1",
			body:       `{"formality":0}`,
			token:      &token,
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "formality above max=5",
			body:       `{"formality":6}`,
			token:      &token,
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:  "formality exactly 1 — valid",
			body:  `{"formality":1}`,
			token: &token,
			setupUC: func(u *mockUC.IRecommendationUseCase) {
				u.On("GetUserRecommendation", mock.Anything, 1, userID).
					Return(successResult, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "formality exactly 5 — valid",
			body:  `{"formality":5}`,
			token: &token,
			setupUC: func(u *mockUC.IRecommendationUseCase) {
				u.On("GetUserRecommendation", mock.Anything, 5, userID).
					Return(successResult, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid json body",
			body:       `{invalid`,
			token:      &token,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:  "usecase internal error",
			body:  `{"formality":3}`,
			token: &token,
			setupUC: func(u *mockUC.IRecommendationUseCase) {
				u.On("GetUserRecommendation", mock.Anything, 3, userID).
					Return(nil, errors.New("ml unavailable"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := newEcho()
			uc := mockUC.NewIRecommendationUseCase(t)
			log := mockLogger.NewLogger(t)

			if tc.setupUC != nil {
				tc.setupUC(uc)
			}

			setupRouter(e, uc, log)

			req, rec := makeReq(http.MethodPut, "/recommendation", tc.body, tc.token)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.wantStatus, rec.Code)

			if tc.wantStatus == http.StatusOK {
				var resp entity.RecommendationsCatalogRequest
				require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
				if tc.wantLogID != nil {
					assert.Equal(t, *tc.wantLogID, resp.LogID)
				}
				if tc.wantCatalogLen != nil {
					assert.Len(t, resp.Catalog, *tc.wantCatalogLen)
				}
			}
		})
	}
}
