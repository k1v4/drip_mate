package v1

import (
	"net/http"
	"net/http/httptest"
	"testing"
	mocks "user_service/mocks/internal_/usecase"
	"user_service/pkg/logger"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewRouter_HealthCheck(t *testing.T) {
	e := echo.New()
	mockSvc := mocks.NewISsoService(t)
	l := logger.NewLogger() // или nil, если логгер не важен

	NewRouter(e, l, mockSvc)

	// создаём HTTP запрос к маршруту
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()

	// обрабатываем запрос через echo
	e.ServeHTTP(rec, req)

	// проверяем статус и тело ответа
	if assert.Equal(t, http.StatusOK, rec.Code) {
		assert.Contains(t, rec.Body.String(), `"status":"ok"`)
	}
}

func TestNewRouter_RoutesExist(t *testing.T) {
	e := echo.New()
	mockSvc := mocks.NewISsoService(t)
	l := logger.NewLogger()

	NewRouter(e, l, mockSvc)

	routes := []struct {
		method string
		path   string
		mock   func()
	}{
		{
			method: "POST",
			path:   "/api/v1/register",
		},
		{
			method: "POST",
			path:   "/api/v1/login",
		},
		{
			method: "PUT",
			path:   "/api/v1/users",
		},
		{
			method: "DELETE",
			path:   "/api/v1/users",
		},
		{
			method: "POST",
			path:   "/api/v1/refresh",
			mock: func() {
				mockSvc.
					EXPECT().
					RefreshToken(mock.Anything, "token").
					Return("access-token", "refresh-token", nil).
					Once()
			},
		},
	}

	for _, r := range routes {
		t.Run(r.method+" "+r.path, func(t *testing.T) {
			req := httptest.NewRequest(r.method, r.path, nil)
			rec := httptest.NewRecorder()

			if r.path == "/api/v1/refresh" {
				req.Header.Set(echo.HeaderAuthorization, "Bearer token")
			}

			if r.mock != nil {
				r.mock()
			}

			e.ServeHTTP(rec, req)

			// Проверяем, что маршрут существует (не вернул 404)
			assert.NotEqual(t, http.StatusNotFound, rec.Code, "route %s %s should exist", r.method, r.path)
		})
	}
}
