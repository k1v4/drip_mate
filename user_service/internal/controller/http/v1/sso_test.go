package v1

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"user_service/internal/entity"
	"user_service/internal/usecase"
	mocks "user_service/mocks/internal_/usecase"
	"user_service/pkg/jwtpkg"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupEcho() *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	return e
}

func TestAuthController_Login(t *testing.T) {
	e := setupEcho()
	mockSvc := mocks.NewISsoService(t)
	newSsoRoutes(e.Group("/api/v1"), mockSvc, nil)

	tests := []struct {
		name           string
		reqBody        string
		mockReturn     func()
		expectedStatus int
		expectedBody   string
	}{
		{
			name:    "success",
			reqBody: `{"email":"user@mail.com","password":"password123"}`,
			mockReturn: func() {
				mockSvc.
					EXPECT().
					Login(mock.Anything, "user@mail.com", "password123").
					Return(1, "access-token", "refresh-token", nil).
					Once()
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"access_token":"access-token","refresh_token":"refresh-token","access_id":1}`,
		},
		{
			name:           "invalid request",
			reqBody:        `not json`,
			mockReturn:     func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"bad request"}`,
		},
		{
			name:           "empty password",
			reqBody:        `{"email":"user@mail.com","password":""}`,
			mockReturn:     func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"bad request"}`,
		},
		{
			name:    "invalid credentials",
			reqBody: `{"email":"user@mail.com","password":"wrong"}`,
			mockReturn: func() {
				mockSvc.
					EXPECT().
					Login(mock.Anything, "user@mail.com", "wrong").
					Return(0, "", "", usecase.ErrInvalidCredentials).
					Once()
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"invalid credentials"}`,
		},
		{
			name:    "no user",
			reqBody: `{"email":"nouser@mail.com","password":"password123"}`,
			mockReturn: func() {
				mockSvc.
					EXPECT().
					Login(mock.Anything, "nouser@mail.com", "password123").
					Return(0, "", "", usecase.ErrNoUser).
					Once()
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"no user"}`,
		},
		{
			name:    "internal error",
			reqBody: `{"email":"user@mail.com","password":"password123"}`,
			mockReturn: func() {
				mockSvc.
					EXPECT().
					Login(mock.Anything, "user@mail.com", "password123").
					Return(0, "", "", errors.New("something bad")).
					Once()
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"internal error"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(tc.reqBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			ctx := e.NewContext(req, rec)

			tc.mockReturn()

			handler := &containerRoutes{t: mockSvc, l: nil}
			_ = handler.Auth(ctx)

			assert.Equal(t, tc.expectedStatus, rec.Code)
			if tc.expectedBody != "" {
				assert.Contains(t, rec.Body.String(), tc.expectedBody)
			}
		})
	}
}

func TestAuthController_Register(t *testing.T) {
	e := setupEcho()
	mockSvc := mocks.NewISsoService(t)
	handler := &containerRoutes{t: mockSvc, l: nil}

	tests := []struct {
		name           string
		reqBody        string
		mockReturn     func()
		expectedStatus int
		expectedBody   string
	}{
		{
			name:    "success",
			reqBody: `{"email":"new@mail.com","password":"password12345"}`,
			mockReturn: func() {
				mockSvc.EXPECT().
					Register(mock.Anything, "new@mail.com", "password12345").
					Return(10, nil).
					Once()
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"user_id":10}`,
		},
		{
			name:           "invalid request",
			reqBody:        `not json`,
			mockReturn:     func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"bad request"}`,
		},
		{
			name:           "short password",
			reqBody:        `{"email":"short@mail.com","password":"123"}`,
			mockReturn:     func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"password must be equal or longer than 10"}`,
		},
		{
			name:           "empty email",
			reqBody:        `{"email":"","password":"password12345"}`,
			mockReturn:     func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"email is required"}`,
		},
		{
			name:    "user exist",
			reqBody: `{"email":"exist@mail.com","password":"password12345"}`,
			mockReturn: func() {
				mockSvc.EXPECT().
					Register(mock.Anything, "exist@mail.com", "password12345").
					Return(0, usecase.ErrUserExist).
					Once()
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"email or username is exist"}`,
		},
		{
			name:    "user exist",
			reqBody: `{"email":"exist@mail.com","password":"password12345"}`,
			mockReturn: func() {
				mockSvc.EXPECT().
					Register(mock.Anything, "exist@mail.com", "password12345").
					Return(0, errors.New("something bad")).
					Once()
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"internal error"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(tc.reqBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			ctx := e.NewContext(req, rec)

			tc.mockReturn()
			_ = handler.Register(ctx)

			assert.Equal(t, tc.expectedStatus, rec.Code)
			if tc.expectedBody != "" {
				assert.Contains(t, rec.Body.String(), tc.expectedBody)
			}
		})
	}
}

func TestAuthController_UpdateUserInfo(t *testing.T) {
	e := setupEcho()
	mockSvc := mocks.NewISsoService(t)
	handler := &containerRoutes{t: mockSvc, l: nil}

	tests := []struct {
		name           string
		reqBody        string
		needToken      bool
		token          string
		mockReturn     func()
		expectedStatus int
		expectedBody   string
	}{
		{
			name:      "success",
			reqBody:   `{"email":"upd@mail.com","password":"password12345","name":"John","surname":"Doe","username":"jdoe","city":"NY"}`,
			needToken: true,
			token:     "valid-token",
			mockReturn: func() {
				mockSvc.EXPECT().
					UpdateUserInfo(mock.Anything, 1, "upd@mail.com", "password12345", "John", "Doe", "jdoe", "NY").
					Return(entity.User{ID: 1, Email: "upd@mail.com"}, nil).Once()
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"email":"upd@mail.com"`,
		},
		{
			name:      "usecase error",
			reqBody:   `{"email":"upd@mail.com","password":"password12345","name":"John","surname":"Doe","username":"jdoe","city":"NY"}`,
			needToken: true,
			token:     "valid-token",
			mockReturn: func() {
				mockSvc.EXPECT().
					UpdateUserInfo(mock.Anything, 1, "upd@mail.com", "password12345", "John", "Doe", "jdoe", "NY").
					Return(entity.User{}, errors.New("usecase error")).Once()
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `"error":"internal error"`,
		},
		{
			name:           "invalid password",
			reqBody:        `{"email":"upd@mail.com","password":"short","name":"John","surname":"Doe","username":"jdoe","city":"NY"}`,
			needToken:      true,
			token:          "valid-token",
			mockReturn:     func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"error":"bad request"`,
		},
		{
			name:           "no email",
			reqBody:        `{"password":"password12345","name":"John","surname":"Doe","username":"jdoe","city":"NY"}`,
			needToken:      true,
			token:          "valid-token",
			mockReturn:     func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"error":"bad request"`,
		},
		{
			name:           "wrong token",
			needToken:      false,
			token:          "valid-token",
			mockReturn:     func() {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `"error":"wrong token"`,
		},
		{
			name:           "wrong request",
			reqBody:        `not a json`,
			needToken:      true,
			token:          "valid-token",
			mockReturn:     func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"error":"bad request"`,
		},
		{
			name:           "no token",
			needToken:      false,
			reqBody:        `{}`,
			token:          "",
			mockReturn:     func() {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "token is required",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPut, "/users", strings.NewReader(tc.reqBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			if tc.token != "" {
				req.Header.Set(echo.HeaderAuthorization, "Bearer "+tc.token)
			}
			ctx := e.NewContext(req, rec)

			if tc.needToken {
				user := entity.User{ID: 1, Email: "upd@mail.com", AccessLevelId: 1}
				token, _ := jwtpkg.NewAccessToken(user, 15*time.Minute)
				req.Header.Set("Authorization", "Bearer "+token)
			}

			tc.mockReturn()
			_ = handler.UpdateUserInfo(ctx)

			assert.Equal(t, tc.expectedStatus, rec.Code)
			if tc.expectedBody != "" {
				assert.Contains(t, rec.Body.String(), tc.expectedBody)
			}
		})
	}
}

func TestAuthController_DeleteAccount(t *testing.T) {
	e := setupEcho()
	mockSvc := mocks.NewISsoService(t)
	handler := &containerRoutes{t: mockSvc, l: nil}

	tests := []struct {
		name           string
		token          string
		needToken      bool
		mockReturn     func()
		expectedStatus int
		expectedBody   string
	}{
		{
			name:      "success",
			needToken: true,
			token:     "valid-token",
			mockReturn: func() {
				mockSvc.EXPECT().
					DeleteAccount(mock.Anything, 1).
					Return(true, nil).
					Once()
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"is_successfully":true`,
		},
		{
			name:      "usecase error",
			needToken: true,
			token:     "valid-token",
			mockReturn: func() {
				mockSvc.EXPECT().
					DeleteAccount(mock.Anything, 1).
					Return(false, errors.New("usecase_error")).
					Once()
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"internal error"}`,
		},
		{
			name:           "invalid token",
			needToken:      false,
			token:          "invalid-token",
			mockReturn:     func() {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"bad request"}`,
		},
		{
			name:           "no token",
			needToken:      false,
			token:          "",
			mockReturn:     func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "bad request",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/users", nil)
			if tc.token != "" {
				req.Header.Set(echo.HeaderAuthorization, "Bearer "+tc.token)
			}
			ctx := e.NewContext(req, rec)

			if tc.needToken {
				user := entity.User{ID: 1, Email: "upd@mail.com", AccessLevelId: 1}
				token, _ := jwtpkg.NewAccessToken(user, 15*time.Minute)
				req.Header.Set("Authorization", "Bearer "+token)
			}

			tc.mockReturn()
			_ = handler.DeleteAccount(ctx)

			assert.Equal(t, tc.expectedStatus, rec.Code)
			if tc.expectedBody != "" {
				assert.Contains(t, rec.Body.String(), tc.expectedBody)
			}
		})
	}
}

func TestAuthController_RefreshToken(t *testing.T) {
	e := setupEcho()
	mockSvc := mocks.NewISsoService(t)
	handler := &containerRoutes{t: mockSvc, l: nil}

	tests := []struct {
		name           string
		token          string
		mockReturn     func()
		expectedStatus int
		expectedBody   string
	}{
		{
			name:  "success",
			token: "old-refresh",
			mockReturn: func() {
				mockSvc.EXPECT().
					RefreshToken(mock.Anything, "old-refresh").
					Return("new-access", "new-refresh", nil).
					Once()
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"access_token":"new-access"`,
		},
		{
			name:  "invalid token",
			token: "bad-refresh",
			mockReturn: func() {
				mockSvc.EXPECT().
					RefreshToken(mock.Anything, "bad-refresh").
					Return("", "", errors.New("invalid")).
					Once()
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "token error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/refresh", nil)
			if tc.token != "" {
				req.Header.Set(echo.HeaderAuthorization, "Bearer "+tc.token)
			}
			ctx := e.NewContext(req, rec)

			tc.mockReturn()
			_ = handler.RefreshToken(ctx)

			assert.Equal(t, tc.expectedStatus, rec.Code)
			if tc.expectedBody != "" {
				assert.Contains(t, rec.Body.String(), tc.expectedBody)
			}
		})
	}
}
