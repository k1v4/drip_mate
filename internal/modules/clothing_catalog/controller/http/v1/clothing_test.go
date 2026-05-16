package v1_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/k1v4/drip_mate/internal/config"
	"github.com/k1v4/drip_mate/internal/entity"
	v1 "github.com/k1v4/drip_mate/internal/modules/clothing_catalog/controller/http/v1"
	userEntity "github.com/k1v4/drip_mate/internal/modules/user_service/entity"
	mockUC "github.com/k1v4/drip_mate/mocks/internal_/modules/clothing_catalog/controller/http/v1"
	mockLogger "github.com/k1v4/drip_mate/mocks/pkg/logger"
	"github.com/k1v4/drip_mate/pkg/DataBase"
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

func setupRouter(e *echo.Echo, uc *mockUC.IClothingUseCase, log *mockLogger.Logger) {
	g := e.Group("")
	v1.NewCatalogRoutes(g, uc, log, defaultCfg())
}

func generateToken(t *testing.T, userID uuid.UUID, accessID int) string {
	t.Helper()
	cfg := defaultCfg()
	token, err := jwtpkg.NewAccessToken(
		&userEntity.User{ID: userID, AccessID: accessID},
		cfg.TTL, cfg.Secret, cfg.Issuer,
	)
	require.NoError(t, err)
	return token
}

func makeJSONReq(method, path, body string, token *string) (*http.Request, *httptest.ResponseRecorder) {
	var buf *bytes.Reader
	if body != "" {
		buf = bytes.NewReader([]byte(body))
	} else {
		buf = bytes.NewReader(nil)
	}
	req := httptest.NewRequestWithContext(context.Background(), method, path, buf)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	if token != nil {
		req.AddCookie(&http.Cookie{Name: "access_token", Value: *token})
	}
	return req, httptest.NewRecorder()
}

func makeMultipartReq(
	t *testing.T,
	method, path string,
	fields map[string]string,
	fileField, fileName string,
	fileData []byte,
	token *string,
) (*http.Request, *httptest.ResponseRecorder) {
	t.Helper()

	var body bytes.Buffer
	w := multipart.NewWriter(&body)

	for key, val := range fields {
		require.NoError(t, w.WriteField(key, val))
	}

	if fileData != nil {
		fw, err := w.CreateFormFile(fileField, fileName)
		require.NoError(t, err)
		_, err = fw.Write(fileData)
		require.NoError(t, err)
	}

	require.NoError(t, w.Close())

	req := httptest.NewRequestWithContext(context.Background(), method, path, &body)
	req.Header.Set(echo.HeaderContentType, w.FormDataContentType())
	if token != nil {
		req.AddCookie(&http.Cookie{Name: "access_token", Value: *token})
	}
	return req, httptest.NewRecorder()
}

func newCatalog(id uuid.UUID) *entity.Catalog {
	return &entity.Catalog{
		ID:             id,
		Name:           "Test item",
		CategoryID:     1,
		Gender:         new("male"),
		SeasonID:       1,
		FormalityLevel: new(int16(2)),
		Material:       new("cotton"),
		ImageURL:       "https://example.com/image.jpg",
	}
}

func TestCatalogRoutes_GetItem(t *testing.T) {
	itemID := uuid.New()
	userID := uuid.New()

	token := generateToken(t, userID, 1)
	catalog := newCatalog(itemID)

	tests := []struct {
		name       string
		itemID     string
		token      *string
		setupUC    func(u *mockUC.IClothingUseCase)
		wantStatus int
	}{
		{
			name:   "success",
			itemID: itemID.String(),
			token:  &token,
			setupUC: func(u *mockUC.IClothingUseCase) {
				u.On("GetItemByID", mock.Anything, itemID).Return(catalog, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "no token",
			itemID:     itemID.String(),
			token:      nil,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "invalid uuid",
			itemID:     "not-a-uuid",
			token:      &token,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "item not found",
			itemID: itemID.String(),
			token:  &token,
			setupUC: func(u *mockUC.IClothingUseCase) {
				u.On("GetItemByID", mock.Anything, itemID).Return(nil, DataBase.ErrCatalogItemNotFound)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "internal error",
			itemID: itemID.String(),
			token:  &token,
			setupUC: func(u *mockUC.IClothingUseCase) {
				u.On("GetItemByID", mock.Anything, itemID).Return(nil, errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := newEcho()
			uc := mockUC.NewIClothingUseCase(t)
			log := mockLogger.NewLogger(t)

			if tc.setupUC != nil {
				tc.setupUC(uc)
			}

			setupRouter(e, uc, log)

			req, rec := makeJSONReq(http.MethodGet, "/catalogs/"+tc.itemID, "", tc.token)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.wantStatus, rec.Code)
			if tc.wantStatus == http.StatusOK {
				var resp entity.Catalog
				require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
				assert.Equal(t, itemID, resp.ID)
			}
		})
	}
}

func TestCatalogRoutes_DeleteItem(t *testing.T) {
	itemID := uuid.New()
	userID := uuid.New()

	adminToken := generateToken(t, userID, int(entity.RoleAdmin))
	userToken := generateToken(t, userID, 1)

	tests := []struct {
		name       string
		itemID     string
		token      *string
		setupUC    func(u *mockUC.IClothingUseCase)
		wantStatus int
	}{
		{
			name:   "success",
			itemID: itemID.String(),
			token:  &adminToken,
			setupUC: func(u *mockUC.IClothingUseCase) {
				u.On("DeleteItem", mock.Anything, itemID).Return(nil)
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "no token",
			itemID:     itemID.String(),
			token:      nil,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "non-admin token — forbidden",
			itemID:     itemID.String(),
			token:      &userToken,
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "invalid uuid",
			itemID:     "bad-uuid",
			token:      &adminToken,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "item not found",
			itemID: itemID.String(),
			token:  &adminToken,
			setupUC: func(u *mockUC.IClothingUseCase) {
				u.On("DeleteItem", mock.Anything, itemID).Return(DataBase.ErrCatalogItemNotFound)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "internal error",
			itemID: itemID.String(),
			token:  &adminToken,
			setupUC: func(u *mockUC.IClothingUseCase) {
				u.On("DeleteItem", mock.Anything, itemID).Return(errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := newEcho()
			uc := mockUC.NewIClothingUseCase(t)
			log := mockLogger.NewLogger(t)

			if tc.setupUC != nil {
				tc.setupUC(uc)
			}

			setupRouter(e, uc, log)

			req, rec := makeJSONReq(http.MethodDelete, "/catalogs/"+tc.itemID, "", tc.token)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.wantStatus, rec.Code)
		})
	}
}

func TestCatalogRoutes_CreateItem(t *testing.T) {
	userID := uuid.New()
	adminToken := generateToken(t, userID, int(entity.RoleAdmin))
	createdCatalog := newCatalog(uuid.New())

	validFields := map[string]string{
		"name":            "Test jacket",
		"category_id":     "1",
		"gender":          "male",
		"season_id":       "1",
		"formality_level": "2",
		"material":        "cotton",
	}
	imageData := []byte("fake-image-bytes")

	tests := []struct {
		name       string
		fields     map[string]string
		fileData   []byte
		token      *string
		setupUC    func(u *mockUC.IClothingUseCase)
		wantStatus int
	}{
		{
			name:     "success",
			fields:   validFields,
			fileData: imageData,
			token:    &adminToken,
			setupUC: func(u *mockUC.IClothingUseCase) {
				u.On("CreateItem", mock.Anything, mock.Anything, "image.jpg", imageData).
					Return(createdCatalog, nil)
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "no token",
			fields:     validFields,
			fileData:   imageData,
			token:      nil,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "non-admin — forbidden",
			fields:     validFields,
			fileData:   imageData,
			token:      new(generateToken(t, userID, 1)),
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "missing image",
			fields:     validFields,
			fileData:   nil, // нет файла
			token:      &adminToken,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "service error",
			fields:   validFields,
			fileData: imageData,
			token:    &adminToken,
			setupUC: func(u *mockUC.IClothingUseCase) {
				u.On("CreateItem", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("s3 error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := newEcho()
			uc := mockUC.NewIClothingUseCase(t)
			log := mockLogger.NewLogger(t)

			if tc.setupUC != nil {
				tc.setupUC(uc)
			}

			setupRouter(e, uc, log)

			req, rec := makeMultipartReq(
				t,
				http.MethodPost, "/catalogs",
				tc.fields,
				"image", "image.jpg",
				tc.fileData,
				tc.token,
			)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.wantStatus, rec.Code)
		})
	}
}

func TestCatalogRoutes_UpdateItem(t *testing.T) {
	itemID := uuid.New()
	userID := uuid.New()
	adminToken := generateToken(t, userID, int(entity.RoleAdmin))
	updatedCatalog := newCatalog(itemID)

	tests := []struct {
		name       string
		itemID     string
		fields     map[string]string
		fileData   []byte
		token      *string
		setupUC    func(u *mockUC.IClothingUseCase)
		wantStatus int
	}{
		{
			name:     "success — without image",
			itemID:   itemID.String(),
			fields:   map[string]string{"name": "Updated name"},
			fileData: nil,
			token:    &adminToken,
			setupUC: func(u *mockUC.IClothingUseCase) {
				u.On("UpdateItem", mock.Anything, mock.Anything, "", []byte(nil)).
					Return(updatedCatalog, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:     "success — with image",
			itemID:   itemID.String(),
			fields:   map[string]string{"name": "Updated name"},
			fileData: []byte("new-image"),
			token:    &adminToken,
			setupUC: func(u *mockUC.IClothingUseCase) {
				u.On("UpdateItem", mock.Anything, mock.Anything, "image.jpg", []byte("new-image")).
					Return(updatedCatalog, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "no token",
			itemID:     itemID.String(),
			fields:     map[string]string{"name": "Updated"},
			token:      nil,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "non-admin — forbidden",
			itemID:     itemID.String(),
			fields:     map[string]string{"name": "Updated"},
			token:      new(generateToken(t, userID, 1)),
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "invalid uuid",
			itemID:     "bad-uuid",
			fields:     map[string]string{"name": "Updated"},
			token:      &adminToken,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "item not found",
			itemID: itemID.String(),
			fields: map[string]string{"name": "Updated"},
			token:  &adminToken,
			setupUC: func(u *mockUC.IClothingUseCase) {
				u.On("UpdateItem", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil, DataBase.ErrCatalogItemNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:   "internal error",
			itemID: itemID.String(),
			fields: map[string]string{"name": "Updated"},
			token:  &adminToken,
			setupUC: func(u *mockUC.IClothingUseCase) {
				u.On("UpdateItem", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := newEcho()
			uc := mockUC.NewIClothingUseCase(t)
			log := mockLogger.NewLogger(t)

			if tc.setupUC != nil {
				tc.setupUC(uc)
			}

			setupRouter(e, uc, log)

			req, rec := makeMultipartReq(
				t,
				http.MethodPut, "/catalogs/"+tc.itemID,
				tc.fields,
				"image", "image.jpg",
				tc.fileData,
				tc.token,
			)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.wantStatus, rec.Code)
		})
	}
}

func TestCatalogRoutes_GetAllItems(t *testing.T) {
	userID := uuid.New()
	adminToken := generateToken(t, userID, int(entity.RoleAdmin))
	items := []entity.Catalog{*newCatalog(uuid.New()), *newCatalog(uuid.New())}

	tests := []struct {
		name       string
		query      string
		token      *string
		setupUC    func(u *mockUC.IClothingUseCase)
		wantStatus int
		wantTotal  int
	}{
		{
			name:  "success — default pagination",
			query: "",
			token: &adminToken,
			setupUC: func(u *mockUC.IClothingUseCase) {
				// page=1, limit=10 → offset=0
				u.On("GetAllItems", mock.Anything, 10, 0).Return(items, 2, nil)
			},
			wantStatus: http.StatusOK,
			wantTotal:  2,
		},
		{
			name:  "success — custom pagination",
			query: "?page=2&limit=5",
			token: &adminToken,
			setupUC: func(u *mockUC.IClothingUseCase) {
				// page=2, limit=5 → offset=5
				u.On("GetAllItems", mock.Anything, 5, 5).Return(items, 2, nil)
			},
			wantStatus: http.StatusOK,
			wantTotal:  2,
		},
		{
			name:  "invalid page param — falls back to default",
			query: "?page=abc&limit=10",
			token: &adminToken,
			setupUC: func(u *mockUC.IClothingUseCase) {
				u.On("GetAllItems", mock.Anything, 10, 0).Return(items, 2, nil)
			},
			wantStatus: http.StatusOK,
			wantTotal:  2,
		},
		{
			name:       "no token",
			query:      "",
			token:      nil,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "non-admin — forbidden",
			query:      "",
			token:      new(generateToken(t, userID, 1)),
			wantStatus: http.StatusForbidden,
		},
		{
			name:  "internal error",
			query: "",
			token: &adminToken,
			setupUC: func(u *mockUC.IClothingUseCase) {
				u.On("GetAllItems", mock.Anything, 10, 0).Return(nil, 0, errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := newEcho()
			uc := mockUC.NewIClothingUseCase(t)
			log := mockLogger.NewLogger(t)

			if tc.setupUC != nil {
				tc.setupUC(uc)
			}

			setupRouter(e, uc, log)

			req, rec := makeJSONReq(http.MethodGet, "/catalogs"+tc.query, "", tc.token)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.wantStatus, rec.Code)
			if tc.wantStatus == http.StatusOK {
				var resp map[string]any
				require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
				total, ok := resp["total"].(float64)
				require.True(t, ok)
				assert.Equal(t, float64(tc.wantTotal), total)
			}
		})
	}
}
