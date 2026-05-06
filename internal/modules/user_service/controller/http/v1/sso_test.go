package v1_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/k1v4/drip_mate/internal/config"
	internalEntity "github.com/k1v4/drip_mate/internal/entity"
	v1 "github.com/k1v4/drip_mate/internal/modules/user_service/controller/http/v1"
	"github.com/k1v4/drip_mate/internal/modules/user_service/entity"
	"github.com/k1v4/drip_mate/internal/modules/user_service/usecase"
	mockSvc "github.com/k1v4/drip_mate/mocks/internal_/modules/user_service/usecase"
	mockLogger "github.com/k1v4/drip_mate/mocks/pkg/logger"
	"github.com/k1v4/drip_mate/pkg/DataBase"
	"github.com/k1v4/drip_mate/pkg/jwtpkg"
	appValidator "github.com/k1v4/drip_mate/pkg/validator"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func defaultTokenCfg() *config.Token {
	return &config.Token{
		TTL:    time.Hour,
		Secret: "test-secret",
		Issuer: "test-issuer",
	}
}

func newEcho() *echo.Echo {
	e := echo.New()
	e.Validator = appValidator.New()
	return e
}

func setupRouter(e *echo.Echo, svc *mockSvc.ISsoService, log *mockLogger.Logger) {
	g := e.Group("")
	v1.NewSsoRoutes(g, svc, log, defaultTokenCfg())
}

// generateToken создаёт валидный JWT токен для тестов
func generateToken(t *testing.T, userID uuid.UUID) string {
	t.Helper()
	cfg := defaultTokenCfg()
	token, err := jwtpkg.NewAccessToken(
		&entity.User{ID: userID, AccessID: 1},
		cfg.TTL,
		cfg.Secret,
		cfg.Issuer,
	)
	require.NoError(t, err)
	return token
}

// makeReq создаёт http.Request с JSON-телом и опциональной JWT-кукой
func makeReq(method, path, body string, token *string) (*http.Request, *httptest.ResponseRecorder) {
	var reqBody *bytes.Reader
	if body != "" {
		reqBody = bytes.NewReader([]byte(body))
	} else {
		reqBody = bytes.NewReader(nil)
	}

	req := httptest.NewRequestWithContext(context.Background(), method, path, reqBody)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	if token != nil {
		req.AddCookie(&http.Cookie{
			Name:  "access_token",
			Value: *token,
		})
	}

	return req, httptest.NewRecorder()
}

func TestContainerRoutes_Auth(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		setupSvc   func(s *mockSvc.ISsoService)
		setupLog   func(l *mockLogger.Logger)
		wantStatus int
		wantCookie bool
	}{
		{
			name: "success",
			body: `{"email":"user@example.com","password":"secret"}`,
			setupSvc: func(s *mockSvc.ISsoService) {
				s.On("Login", mock.Anything, "user@example.com", "secret").
					Return(internalEntity.Role(1), "token_abc", nil)
			},
			wantStatus: http.StatusOK,
			wantCookie: true,
		},
		{
			name: "empty password",
			body: `{"email":"user@example.com","password":""}`,
			setupLog: func(l *mockLogger.Logger) {
				l.On("Error", mock.Anything, mock.AnythingOfType("string")).Maybe()
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "empty email",
			body: `{"email":"","password":"secret"}`,
			setupLog: func(l *mockLogger.Logger) {
				l.On("Error", mock.Anything, mock.AnythingOfType("string")).Maybe()
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "invalid credentials",
			body: `{"email":"user@example.com","password":"wrong"}`,
			setupSvc: func(s *mockSvc.ISsoService) {
				s.On("Login", mock.Anything, "user@example.com", "wrong").
					Return(internalEntity.Role(0), "", usecase.ErrInvalidCredentials)
			},
			setupLog: func(l *mockLogger.Logger) {
				l.On("Error", mock.Anything, mock.AnythingOfType("string")).Maybe()
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "user not found",
			body: `{"email":"ghost@example.com","password":"secret"}`,
			setupSvc: func(s *mockSvc.ISsoService) {
				s.On("Login", mock.Anything, "ghost@example.com", "secret").
					Return(internalEntity.Role(0), "", usecase.ErrNoUser)
			},
			setupLog: func(l *mockLogger.Logger) {
				l.On("Error", mock.Anything, mock.AnythingOfType("string")).Maybe()
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "internal error",
			body: `{"email":"user@example.com","password":"secret"}`,
			setupSvc: func(s *mockSvc.ISsoService) {
				s.On("Login", mock.Anything, "user@example.com", "secret").
					Return(internalEntity.Role(0), "", errors.New("db down"))
			},
			setupLog: func(l *mockLogger.Logger) {
				l.On("Error", mock.Anything, mock.AnythingOfType("string")).Maybe()
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := newEcho()
			svc := mockSvc.NewISsoService(t)
			log := mockLogger.NewLogger(t)

			if tc.setupSvc != nil {
				tc.setupSvc(svc)
			}
			if tc.setupLog != nil {
				tc.setupLog(log)
			}

			setupRouter(e, svc, log)

			req, rec := makeReq(http.MethodPost, "/login", tc.body, nil)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.wantStatus, rec.Code)
			if tc.wantCookie {
				assert.NotEmpty(t, rec.Header().Get("Set-Cookie"))
			}
		})
	}
}

func TestContainerRoutes_Register(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		setupSvc   func(s *mockSvc.ISsoService)
		setupLog   func(l *mockLogger.Logger)
		wantStatus int
		wantCookie bool
	}{
		{
			name: "success",
			body: `{"email":"new@example.com","password":"strongpassword"}`,
			setupSvc: func(s *mockSvc.ISsoService) {
				s.On("Register", mock.Anything, "new@example.com", "strongpassword").
					Return(internalEntity.Role(1), "token_xyz", nil)
			},
			wantStatus: http.StatusOK,
			wantCookie: true,
		},
		{
			name:       "password too short",
			body:       `{"email":"new@example.com","password":"short"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid email",
			body:       `{"email":"not-an-email","password":"strongpassword"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "user already exists",
			body: `{"email":"exist@example.com","password":"strongpassword"}`,
			setupSvc: func(s *mockSvc.ISsoService) {
				s.On("Register", mock.Anything, "exist@example.com", "strongpassword").
					Return(internalEntity.Role(0), "", usecase.ErrUserExist)
			},
			setupLog: func(l *mockLogger.Logger) {
				l.On("Error", mock.Anything, mock.AnythingOfType("string")).Maybe()
			},
			wantStatus: http.StatusConflict,
		},
		{
			name: "internal error",
			body: `{"email":"new@example.com","password":"strongpassword"}`,
			setupSvc: func(s *mockSvc.ISsoService) {
				s.On("Register", mock.Anything, "new@example.com", "strongpassword").
					Return(internalEntity.Role(0), "", errors.New("db down"))
			},
			setupLog: func(l *mockLogger.Logger) {
				l.On("Error", mock.Anything, mock.AnythingOfType("string")).Maybe()
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := newEcho()
			svc := mockSvc.NewISsoService(t)
			log := mockLogger.NewLogger(t)

			if tc.setupSvc != nil {
				tc.setupSvc(svc)
			}
			if tc.setupLog != nil {
				tc.setupLog(log)
			}

			setupRouter(e, svc, log)

			req, rec := makeReq(http.MethodPost, "/register", tc.body, nil)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.wantStatus, rec.Code)
			if tc.wantCookie {
				assert.NotEmpty(t, rec.Header().Get("Set-Cookie"))
			}
		})
	}
}

func TestContainerRoutes_DeleteAccount(t *testing.T) {
	userID := uuid.New()
	token := generateToken(t, userID)

	tests := []struct {
		name       string
		token      *string
		setupSvc   func(s *mockSvc.ISsoService)
		setupLog   func(l *mockLogger.Logger)
		wantStatus int
	}{
		{
			name:  "success",
			token: &token,
			setupSvc: func(s *mockSvc.ISsoService) {
				s.On("DeleteAccount", mock.Anything, userID.String()).Return(true, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "no token",
			token:      nil,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:  "internal error",
			token: &token,
			setupSvc: func(s *mockSvc.ISsoService) {
				s.On("DeleteAccount", mock.Anything, userID.String()).Return(false, errors.New("db error"))
			},
			setupLog: func(l *mockLogger.Logger) {
				l.On("Error", mock.Anything, mock.AnythingOfType("string")).Maybe()
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := newEcho()
			svc := mockSvc.NewISsoService(t)
			log := mockLogger.NewLogger(t)

			if tc.setupSvc != nil {
				tc.setupSvc(svc)
			}
			if tc.setupLog != nil {
				tc.setupLog(log)
			}

			setupRouter(e, svc, log)

			req, rec := makeReq(http.MethodDelete, "/users", "", tc.token)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.wantStatus, rec.Code)
		})
	}
}

func TestContainerRoutes_GetUserByID(t *testing.T) {
	userID := uuid.New()
	token := generateToken(t, userID)
	user := &entity.User{ID: userID, Email: "user@example.com"}

	tests := []struct {
		name       string
		token      *string
		setupSvc   func(s *mockSvc.ISsoService)
		wantStatus int
	}{
		{
			name:  "success",
			token: &token,
			setupSvc: func(s *mockSvc.ISsoService) {
				s.On("GetUserByID", mock.Anything, userID).Return(user, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "no token",
			token:      nil,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:  "internal error",
			token: &token,
			setupSvc: func(s *mockSvc.ISsoService) {
				s.On("GetUserByID", mock.Anything, userID).Return(nil, errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := newEcho()
			svc := mockSvc.NewISsoService(t)
			log := mockLogger.NewLogger(t)

			if tc.setupSvc != nil {
				tc.setupSvc(svc)
			}

			setupRouter(e, svc, log)

			req, rec := makeReq(http.MethodGet, "/users", "", tc.token)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.wantStatus, rec.Code)
			if tc.wantStatus == http.StatusOK {
				var resp entity.User
				require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
				assert.Equal(t, userID, resp.ID)
			}
		})
	}
}

func TestContainerRoutes_SaveOutfit(t *testing.T) {
	userID := uuid.New()
	token := generateToken(t, userID)
	outfitID := uuid.New()
	catalogID := uuid.New()

	validBody := func() string {
		b, _ := json.Marshal(entity.SaveOutfitRequest{
			Name:           "Test outfit",
			CatalogItemIDs: []uuid.UUID{catalogID},
		})
		return string(b)
	}

	tests := []struct {
		name       string
		body       string
		token      *string
		setupSvc   func(s *mockSvc.ISsoService)
		setupLog   func(l *mockLogger.Logger)
		wantStatus int
	}{
		{
			name:  "success",
			body:  validBody(),
			token: &token,
			setupSvc: func(s *mockSvc.ISsoService) {
				s.On("SaveOutfit", mock.Anything, userID, mock.Anything).Return(outfitID, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "no token",
			body:       validBody(),
			token:      nil,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "missing name",
			body:       `{"catalog_item_ids":["` + catalogID.String() + `"]}`,
			token:      &token,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "empty catalog_item_ids",
			body:       `{"name":"outfit","catalog_item_ids":[]}`,
			token:      &token,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:  "service error",
			body:  validBody(),
			token: &token,
			setupSvc: func(s *mockSvc.ISsoService) {
				s.On("SaveOutfit", mock.Anything, userID, mock.Anything).Return(uuid.Nil, errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := newEcho()
			svc := mockSvc.NewISsoService(t)
			log := mockLogger.NewLogger(t)

			if tc.setupSvc != nil {
				tc.setupSvc(svc)
			}
			if tc.setupLog != nil {
				tc.setupLog(log)
			}

			setupRouter(e, svc, log)

			req, rec := makeReq(http.MethodPost, "/users/outfit", tc.body, tc.token)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.wantStatus, rec.Code)
		})
	}
}

func TestContainerRoutes_GetOutfits(t *testing.T) {
	userID := uuid.New()
	token := generateToken(t, userID)
	outfits := []entity.Outfit{
		{ID: uuid.New(), Name: "Summer"},
		{ID: uuid.New(), Name: "Winter"},
	}

	tests := []struct {
		name       string
		token      *string
		setupSvc   func(s *mockSvc.ISsoService)
		setupLog   func(l *mockLogger.Logger)
		wantStatus int
		wantLen    int
	}{
		{
			name:  "success",
			token: &token,
			setupSvc: func(s *mockSvc.ISsoService) {
				s.On("GetOutfits", mock.Anything, userID).Return(outfits, nil)
			},
			wantStatus: http.StatusOK,
			wantLen:    2,
		},
		{
			name:       "no token",
			token:      nil,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:  "internal error",
			token: &token,
			setupSvc: func(s *mockSvc.ISsoService) {
				s.On("GetOutfits", mock.Anything, userID).Return(nil, errors.New("db error"))
			},
			setupLog: func(l *mockLogger.Logger) {
				l.On("Error", mock.Anything, mock.AnythingOfType("string")).Maybe()
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := newEcho()
			svc := mockSvc.NewISsoService(t)
			log := mockLogger.NewLogger(t)

			if tc.setupSvc != nil {
				tc.setupSvc(svc)
			}
			if tc.setupLog != nil {
				tc.setupLog(log)
			}

			setupRouter(e, svc, log)

			req, rec := makeReq(http.MethodGet, "/users/outfit", "", tc.token)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.wantStatus, rec.Code)
			if tc.wantStatus == http.StatusOK {
				var resp []entity.Outfit
				require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
				assert.Len(t, resp, tc.wantLen)
			}
		})
	}
}

func TestContainerRoutes_DeleteOutfit(t *testing.T) {
	userID := uuid.New()
	token := generateToken(t, userID)
	outfitID := uuid.New()

	tests := []struct {
		name       string
		outfitID   string
		token      *string
		setupSvc   func(s *mockSvc.ISsoService)
		wantStatus int
	}{
		{
			name:     "success",
			outfitID: outfitID.String(),
			token:    &token,
			setupSvc: func(s *mockSvc.ISsoService) {
				s.On("DeleteOutfit", mock.Anything, userID, outfitID).Return(nil)
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "no token",
			outfitID:   outfitID.String(),
			token:      nil,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "invalid outfit uuid",
			outfitID:   "bad-uuid",
			token:      &token,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "outfit not found",
			outfitID: outfitID.String(),
			token:    &token,
			setupSvc: func(s *mockSvc.ISsoService) {
				s.On("DeleteOutfit", mock.Anything, userID, outfitID).Return(DataBase.ErrOutfitNotFound)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "internal error",
			outfitID: outfitID.String(),
			token:    &token,
			setupSvc: func(s *mockSvc.ISsoService) {
				s.On("DeleteOutfit", mock.Anything, userID, outfitID).Return(errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := newEcho()
			svc := mockSvc.NewISsoService(t)
			log := mockLogger.NewLogger(t)

			if tc.setupSvc != nil {
				tc.setupSvc(svc)
			}

			setupRouter(e, svc, log)

			req, rec := makeReq(http.MethodDelete, "/users/outfit/"+tc.outfitID, "", tc.token)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.wantStatus, rec.Code)
		})
	}
}

func TestContainerRoutes_PassChange(t *testing.T) {
	userID := uuid.New()
	token := generateToken(t, userID)

	tests := []struct {
		name       string
		body       string
		token      *string
		setupSvc   func(s *mockSvc.ISsoService)
		wantStatus int
	}{
		{
			name:  "success",
			body:  `{"curr_password":"oldpassword1","new_password":"newpassword1"}`,
			token: &token,
			setupSvc: func(s *mockSvc.ISsoService) {
				s.On("UpdatePassword", mock.Anything, userID, mock.Anything).Return(nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "no token",
			body:       `{"curr_password":"oldpassword1","new_password":"newpassword1"}`,
			token:      nil,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "passwords are equal",
			body:       `{"curr_password":"samepassword1","new_password":"samepassword1"}`,
			token:      &token,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "new password too short",
			body:       `{"curr_password":"oldpassword1","new_password":"short"}`,
			token:      &token,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:  "invalid current password",
			body:  `{"curr_password":"wrongpassword","new_password":"newpassword1"}`,
			token: &token,
			setupSvc: func(s *mockSvc.ISsoService) {
				s.On("UpdatePassword", mock.Anything, userID, mock.Anything).
					Return(usecase.ErrInvalidCredentials)
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:  "service error",
			body:  `{"curr_password":"oldpassword1","new_password":"newpassword1"}`,
			token: &token,
			setupSvc: func(s *mockSvc.ISsoService) {
				s.On("UpdatePassword", mock.Anything, userID, mock.Anything).
					Return(errors.New("db error"))
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := newEcho()
			svc := mockSvc.NewISsoService(t)
			log := mockLogger.NewLogger(t)

			if tc.setupSvc != nil {
				tc.setupSvc(svc)
			}

			setupRouter(e, svc, log)

			req, rec := makeReq(http.MethodPost, "/auth/change-password", tc.body, tc.token)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.wantStatus, rec.Code)
		})
	}
}

func TestContainerRoutes_UpdateUserInfo(t *testing.T) {
	userID := uuid.New()
	token := generateToken(t, userID)
	updatedUser := &entity.User{ID: userID, Name: "Ivan", Surname: "Petrov", Username: "ivan_p"}

	tests := []struct {
		name       string
		body       string
		token      *string
		setupSvc   func(s *mockSvc.ISsoService)
		setupLog   func(l *mockLogger.Logger)
		wantStatus int
	}{
		{
			name:  "success",
			body:  `{"name":"Ivan","surname":"Petrov","username":"ivan_p","gender":"male"}`,
			token: &token,
			setupSvc: func(s *mockSvc.ISsoService) {
				s.On("UpdateUserInfo", mock.Anything, userID.String(), "Ivan", "Petrov", "ivan_p", "male").
					Return(updatedUser, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "no token",
			body:       `{"name":"Ivan","surname":"Petrov","username":"ivan_p","gender":"male"}`,
			token:      nil,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:  "service error",
			body:  `{"name":"Ivan","surname":"Petrov","username":"ivan_p","gender":"male"}`,
			token: &token,
			setupSvc: func(s *mockSvc.ISsoService) {
				s.On("UpdateUserInfo", mock.Anything, userID.String(), "Ivan", "Petrov", "ivan_p", "male").
					Return(nil, errors.New("db error"))
			},
			setupLog: func(l *mockLogger.Logger) {
				l.On("Error", mock.Anything, mock.AnythingOfType("string")).Maybe()
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := newEcho()
			svc := mockSvc.NewISsoService(t)
			log := mockLogger.NewLogger(t)

			if tc.setupSvc != nil {
				tc.setupSvc(svc)
			}
			if tc.setupLog != nil {
				tc.setupLog(log)
			}

			setupRouter(e, svc, log)

			req, rec := makeReq(http.MethodPatch, "/me/profile", tc.body, tc.token)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.wantStatus, rec.Code)
		})
	}
}

func TestContainerRoutes_UpdateUserContext(t *testing.T) {
	userID := uuid.New()
	token := generateToken(t, userID)

	tests := []struct {
		name       string
		body       string
		token      *string
		setupSvc   func(s *mockSvc.ISsoService)
		wantStatus int
	}{
		{
			name:  "success with all fields",
			body:  `{"city":"Moscow","styles":[1,2],"colors":[3],"music":[5]}`,
			token: &token,
			setupSvc: func(s *mockSvc.ISsoService) {
				s.On("UpdateContext", mock.Anything, userID, mock.Anything).Return(nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "success with only city",
			body:  `{"city":"Perm"}`,
			token: &token,
			setupSvc: func(s *mockSvc.ISsoService) {
				s.On("UpdateContext", mock.Anything, userID, mock.Anything).Return(nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "no token",
			body:       `{"city":"Moscow"}`,
			token:      nil,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:  "service error",
			body:  `{"city":"Moscow"}`,
			token: &token,
			setupSvc: func(s *mockSvc.ISsoService) {
				s.On("UpdateContext", mock.Anything, userID, mock.Anything).Return(errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := newEcho()
			svc := mockSvc.NewISsoService(t)
			log := mockLogger.NewLogger(t)

			if tc.setupSvc != nil {
				tc.setupSvc(svc)
			}

			setupRouter(e, svc, log)

			req, rec := makeReq(http.MethodPatch, "/me/context", tc.body, tc.token)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.wantStatus, rec.Code)
		})
	}
}
