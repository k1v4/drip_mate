package v1

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/k1v4/drip_mate/pkg/logger"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestNewRouter_HealthCheck(t *testing.T) {
	e := echo.New()
	l := logger.NewLogger() // или nil, если логгер не важен

	NewRouter(e, l)

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
