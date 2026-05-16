package v1

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	internalEntity "github.com/k1v4/drip_mate/internal/entity"
	mockSvc "github.com/k1v4/drip_mate/mocks/internal_/modules/reference_module/controller/http/v1"
	mockLogger "github.com/k1v4/drip_mate/mocks/pkg/logger"
	appValidator "github.com/k1v4/drip_mate/pkg/validator"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newEcho() *echo.Echo {
	e := echo.New()
	e.Validator = appValidator.New()
	return e
}

func setupRouter(e *echo.Echo, svc *mockSvc.IReferenceUseCase, log *mockLogger.Logger) {
	g := e.Group("")
	NewReferencesRoutes(g, svc, log)
}

func TestReferencesRoutes_GetStyles(t *testing.T) {
	tests := []struct {
		name       string
		setupSvc   func(s *mockSvc.IReferenceUseCase)
		wantStatus int
	}{
		{
			name: "success",
			setupSvc: func(s *mockSvc.IReferenceUseCase) {
				s.On("GetStyles", mock.Anything).
					Return([]internalEntity.StyleType{
						{
							ID:   1,
							Name: "Casual",
						},
					}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "internal error",
			setupSvc: func(s *mockSvc.IReferenceUseCase) {
				s.On("GetStyles", mock.Anything).
					Return(nil, errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := newEcho()

			svc := mockSvc.NewIReferenceUseCase(t)
			log := mockLogger.NewLogger(t)

			tc.setupSvc(svc)

			setupRouter(e, svc, log)

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/references/styles", nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.wantStatus, rec.Code)
		})
	}
}

func TestReferencesRoutes_GetColors(t *testing.T) {
	tests := []struct {
		name       string
		setupSvc   func(s *mockSvc.IReferenceUseCase)
		wantStatus int
	}{
		{
			name: "success",
			setupSvc: func(s *mockSvc.IReferenceUseCase) {
				s.On("GetColors", mock.Anything).
					Return([]internalEntity.ColorType{
						{
							ID:   1,
							Name: "Black",
							Hex:  "#000000",
						},
					}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "internal error",
			setupSvc: func(s *mockSvc.IReferenceUseCase) {
				s.On("GetColors", mock.Anything).
					Return(nil, errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := newEcho()

			svc := mockSvc.NewIReferenceUseCase(t)
			log := mockLogger.NewLogger(t)

			tc.setupSvc(svc)

			setupRouter(e, svc, log)

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/references/colors", nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.wantStatus, rec.Code)
		})
	}
}

func TestReferencesRoutes_GetMusics(t *testing.T) {
	tests := []struct {
		name       string
		setupSvc   func(s *mockSvc.IReferenceUseCase)
		wantStatus int
	}{
		{
			name: "success",
			setupSvc: func(s *mockSvc.IReferenceUseCase) {
				s.On("GetMusics", mock.Anything).
					Return([]internalEntity.MusicType{
						{
							ID:   1,
							Name: "Rock",
						},
					}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "internal error",
			setupSvc: func(s *mockSvc.IReferenceUseCase) {
				s.On("GetMusics", mock.Anything).
					Return(nil, errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := newEcho()

			svc := mockSvc.NewIReferenceUseCase(t)
			log := mockLogger.NewLogger(t)

			tc.setupSvc(svc)

			setupRouter(e, svc, log)

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/references/musics", nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.wantStatus, rec.Code)
		})
	}
}

func TestReferencesRoutes_GetCategories(t *testing.T) {
	tests := []struct {
		name       string
		setupSvc   func(s *mockSvc.IReferenceUseCase)
		wantStatus int
	}{
		{
			name: "success",
			setupSvc: func(s *mockSvc.IReferenceUseCase) {
				s.On("GetCategories", mock.Anything).
					Return([]internalEntity.Category{
						{
							ID:   1,
							Name: "T-Shirt",
						},
					}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "internal error",
			setupSvc: func(s *mockSvc.IReferenceUseCase) {
				s.On("GetCategories", mock.Anything).
					Return(nil, errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := newEcho()

			svc := mockSvc.NewIReferenceUseCase(t)
			log := mockLogger.NewLogger(t)

			tc.setupSvc(svc)

			setupRouter(e, svc, log)

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/references/categories", nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.wantStatus, rec.Code)
		})
	}
}

func TestReferencesRoutes_GetSeasons(t *testing.T) {
	tests := []struct {
		name       string
		setupSvc   func(s *mockSvc.IReferenceUseCase)
		wantStatus int
	}{
		{
			name: "success",
			setupSvc: func(s *mockSvc.IReferenceUseCase) {
				s.On("GetSeasons", mock.Anything).
					Return([]internalEntity.Season{
						{
							ID:   1,
							Name: "Winter",
						},
					}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "internal error",
			setupSvc: func(s *mockSvc.IReferenceUseCase) {
				s.On("GetSeasons", mock.Anything).
					Return(nil, errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := newEcho()

			svc := mockSvc.NewIReferenceUseCase(t)
			log := mockLogger.NewLogger(t)

			tc.setupSvc(svc)

			setupRouter(e, svc, log)

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/references/seasons", nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.wantStatus, rec.Code)
		})
	}
}

func TestNewReferencesRoutes(t *testing.T) {
	e := newEcho()

	svc := mockSvc.NewIReferenceUseCase(t)
	log := mockLogger.NewLogger(t)

	g := e.Group("")

	NewReferencesRoutes(g, svc, log)

	routes := e.Routes()

	assert.NotEmpty(t, routes)

	expected := map[string]bool{
		"/references/styles":     false,
		"/references/colors":     false,
		"/references/musics":     false,
		"/references/categories": false,
		"/references/seasons":    false,
	}

	for _, route := range routes {
		if _, ok := expected[route.Path]; ok {
			expected[route.Path] = true
		}
	}

	for route, exists := range expected {
		assert.True(t, exists, "route %s not registered", route)
	}
}

func TestReferencesRoutes_ContextPassed(t *testing.T) {
	e := newEcho()

	svc := mockSvc.NewIReferenceUseCase(t)
	log := mockLogger.NewLogger(t)

	svc.
		On("GetStyles", mock.AnythingOfType("*context.valueCtx")).
		Return([]internalEntity.StyleType{}, nil)

	setupRouter(e, svc, log)

	//nolint:staticcheck
	ctx := context.WithValue(context.Background(), "test-key", "test-value")

	req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/references/styles", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}
