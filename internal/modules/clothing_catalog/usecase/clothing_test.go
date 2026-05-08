package usecase_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/k1v4/drip_mate/internal/entity"
	"github.com/k1v4/drip_mate/internal/modules/clothing_catalog/usecase"
	mockRepo "github.com/k1v4/drip_mate/mocks/internal_/modules/clothing_catalog/usecase"
	mockObject "github.com/k1v4/drip_mate/mocks/internal_/modules/object_gateway/transport/grpc"
	mockKafka "github.com/k1v4/drip_mate/mocks/pkg/kafkaPkg"
	mockLogger "github.com/k1v4/drip_mate/mocks/pkg/logger"
	redispkg "github.com/k1v4/drip_mate/pkg/DataBase/redis"
	"github.com/k1v4/drip_mate/pkg/kafkaPkg"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newRedis(t *testing.T) *redis.Client {
	t.Helper()
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	return client
}

func buildCatalogUC(
	repo *mockRepo.IClothingRepository,
	obj *mockObject.IUploadService,
	writer *mockKafka.KafkaWriter,
	log *mockLogger.Logger,
	cache *redis.Client,
) *usecase.ClothingCatalogUseCase {
	producer := kafkaPkg.NewProducer[entity.CatalogEvent](writer)
	return usecase.NewClothingCatalogUseCase(repo, obj, producer, log, cache)
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
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

func TestClothingCatalogUseCase_GetItemByID_CacheHit(t *testing.T) {
	itemID := uuid.New()
	catalog := newCatalog(itemID)

	cache := newRedis(t)
	repo := mockRepo.NewIClothingRepository(t)
	obj := mockObject.NewIUploadService(t)
	writer := mockKafka.NewKafkaWriter(t)
	log := mockLogger.NewLogger(t)

	data, err := json.Marshal(catalog)
	require.NoError(t, err)
	err = cache.Set(context.Background(), redispkg.GetCatalogItemKey(itemID), data, time.Hour).Err()
	require.NoError(t, err)

	uc := buildCatalogUC(repo, obj, writer, log, cache)
	result, err := uc.GetItemByID(context.Background(), itemID)

	assert.NoError(t, err)
	assert.Equal(t, itemID, result.ID)
}

func TestClothingCatalogUseCase_GetItemByID(t *testing.T) {
	itemID := uuid.New()
	catalog := newCatalog(itemID)

	tests := []struct {
		name      string
		id        uuid.UUID
		setupRepo func(r *mockRepo.IClothingRepository)
		wantErr   bool
	}{
		{
			name: "success — cache miss, fetches from repo and caches",
			id:   itemID,
			setupRepo: func(r *mockRepo.IClothingRepository) {
				r.On("GetItemByID", mock.Anything, itemID).Return(catalog, nil)
			},
			wantErr: false,
		},
		{
			name: "repo error",
			id:   itemID,
			setupRepo: func(r *mockRepo.IClothingRepository) {
				r.On("GetItemByID", mock.Anything, itemID).Return(nil, errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := mockRepo.NewIClothingRepository(t)
			obj := mockObject.NewIUploadService(t)
			writer := mockKafka.NewKafkaWriter(t)
			log := mockLogger.NewLogger(t)

			cache := newRedis(t)

			if tc.setupRepo != nil {
				tc.setupRepo(repo)
			}

			log.On("Error", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()

			uc := buildCatalogUC(repo, obj, writer, log, cache)
			result, err := uc.GetItemByID(context.Background(), tc.id)

			// даём горутине set-кэша завершиться
			time.Sleep(10 * time.Millisecond)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, catalog, result)
			}
		})
	}
}

func TestClothingCatalogUseCase_CreateItem(t *testing.T) {
	itemID := uuid.New()
	catalog := newCatalog(itemID)
	imageData := []byte("fake-image-bytes")
	fileName := "test.jpg"
	uploadedURL := "https://s3.example.com/test.jpg"

	tests := []struct {
		name            string
		req             *entity.CreateCatalogRequest
		setupRepo       func(r *mockRepo.IClothingRepository)
		setupObj        func(o *mockObject.IUploadService)
		setupWriter     func(w *mockKafka.KafkaWriter)
		setupLog        func(l *mockLogger.Logger)
		goroutine       bool
		wantErr         bool
		wantErrContains string
	}{
		{
			name: "success",
			req:  &entity.CreateCatalogRequest{Name: "Test"},
			setupObj: func(o *mockObject.IUploadService) {
				o.On("UploadImage", mock.Anything, fileName, imageData).Return(uploadedURL, nil)
			},
			setupRepo: func(r *mockRepo.IClothingRepository) {
				r.On("CreateItem", mock.Anything, mock.Anything).Return(catalog, nil)
			},
			setupWriter: func(w *mockKafka.KafkaWriter) {
				w.On("WriteMessages", mock.Anything, mock.Anything).Return(nil)
			},
			goroutine: true,
			wantErr:   false,
		},
		{
			name: "upload image error",
			req:  &entity.CreateCatalogRequest{Name: "Test"},
			setupObj: func(o *mockObject.IUploadService) {
				o.On("UploadImage", mock.Anything, fileName, imageData).Return("", errors.New("s3 error"))
			},
			goroutine:       false,
			wantErr:         true,
			wantErrContains: "failed to upload image",
		},
		{
			name: "repo create error — deletes uploaded image",
			req:  &entity.CreateCatalogRequest{Name: "Test"},
			setupObj: func(o *mockObject.IUploadService) {
				o.On("UploadImage", mock.Anything, fileName, imageData).Return(uploadedURL, nil)
				o.On("DeleteImage", mock.Anything, fileName).Return(true, nil)
			},
			setupRepo: func(r *mockRepo.IClothingRepository) {
				r.On("CreateItem", mock.Anything, mock.Anything).Return(nil, errors.New("db error"))
			},
			goroutine:       false,
			wantErr:         true,
			wantErrContains: "failed to create item",
		},
		{
			name: "kafka send error — logged but item returned",
			req:  &entity.CreateCatalogRequest{Name: "Test"},
			setupObj: func(o *mockObject.IUploadService) {
				o.On("UploadImage", mock.Anything, fileName, imageData).Return(uploadedURL, nil)
			},
			setupRepo: func(r *mockRepo.IClothingRepository) {
				r.On("CreateItem", mock.Anything, mock.Anything).Return(catalog, nil)
			},
			setupWriter: func(w *mockKafka.KafkaWriter) {
				w.On("WriteMessages", mock.Anything, mock.Anything).Return(errors.New("kafka error"))
			},
			setupLog: func(l *mockLogger.Logger) {
				l.On("Error", mock.Anything, mock.AnythingOfType("string")).Maybe()
			},
			goroutine: true,
			wantErr:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := mockRepo.NewIClothingRepository(t)
			obj := mockObject.NewIUploadService(t)
			writer := mockKafka.NewKafkaWriter(t)
			log := mockLogger.NewLogger(t)

			var cache *redis.Client
			if tc.goroutine {
				cache = newRedis(t)
				log.On("Error", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()
			}

			if tc.setupRepo != nil {
				tc.setupRepo(repo)
			}
			if tc.setupObj != nil {
				tc.setupObj(obj)
			}
			if tc.setupWriter != nil {
				tc.setupWriter(writer)
			}
			if tc.setupLog != nil {
				tc.setupLog(log)
			}

			uc := buildCatalogUC(repo, obj, writer, log, cache)
			result, err := uc.CreateItem(context.Background(), tc.req, fileName, imageData)

			if tc.goroutine {
				time.Sleep(10 * time.Millisecond)
			}

			if tc.wantErr {
				assert.Error(t, err)
				if tc.wantErrContains != "" {
					assert.ErrorContains(t, err, tc.wantErrContains)
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, catalog, result)
			}
		})
	}
}

func TestClothingCatalogUseCase_UpdateItem(t *testing.T) {
	itemID := uuid.New()
	catalog := newCatalog(itemID)
	imageData := []byte("new-image-bytes")
	fileName := "new.jpg"
	newURL := "https://s3.example.com/new.jpg"
	oldURL := "https://s3.example.com/old.jpg"

	tests := []struct {
		name            string
		req             *entity.UpdateCatalogRequest
		imageData       []byte
		setupRepo       func(r *mockRepo.IClothingRepository)
		setupObj        func(o *mockObject.IUploadService)
		setupWriter     func(w *mockKafka.KafkaWriter)
		setupLog        func(l *mockLogger.Logger)
		goroutine       bool
		wantErr         bool
		wantErrContains string
	}{
		{
			name:      "success — without image update",
			req:       &entity.UpdateCatalogRequest{ID: itemID, Name: "Updated"},
			imageData: nil,
			setupRepo: func(r *mockRepo.IClothingRepository) {
				r.On("UpdateItem", mock.Anything, mock.Anything).Return(catalog, nil)
			},
			setupWriter: func(w *mockKafka.KafkaWriter) {
				w.On("WriteMessages", mock.Anything, mock.Anything).Return(nil)
			},
			goroutine: true,
			wantErr:   false,
		},
		{
			name:      "success — with image update",
			req:       &entity.UpdateCatalogRequest{ID: itemID, Name: "Updated"},
			imageData: imageData,
			setupRepo: func(r *mockRepo.IClothingRepository) {
				existing := newCatalog(itemID)
				existing.ImageURL = oldURL
				r.On("GetItemByID", mock.Anything, itemID).Return(existing, nil)
				r.On("UpdateItem", mock.Anything, mock.Anything).Return(catalog, nil)
			},
			setupObj: func(o *mockObject.IUploadService) {
				o.On("UploadImage", mock.Anything, fileName, imageData).Return(newURL, nil)
				o.On("DeleteImage", mock.Anything, oldURL).Return(true, nil)
			},
			setupWriter: func(w *mockKafka.KafkaWriter) {
				w.On("WriteMessages", mock.Anything, mock.Anything).Return(nil)
			},
			goroutine: true,
			wantErr:   false,
		},
		{
			name:      "get current item error when image provided",
			req:       &entity.UpdateCatalogRequest{ID: itemID},
			imageData: imageData,
			setupRepo: func(r *mockRepo.IClothingRepository) {
				r.On("GetItemByID", mock.Anything, itemID).Return(nil, errors.New("db error"))
			},
			goroutine:       false,
			wantErr:         true,
			wantErrContains: "failed to get current item",
		},
		{
			name:      "upload new image error",
			req:       &entity.UpdateCatalogRequest{ID: itemID},
			imageData: imageData,
			setupRepo: func(r *mockRepo.IClothingRepository) {
				r.On("GetItemByID", mock.Anything, itemID).Return(newCatalog(itemID), nil)
			},
			setupObj: func(o *mockObject.IUploadService) {
				o.On("UploadImage", mock.Anything, fileName, imageData).Return("", errors.New("s3 error"))
			},
			goroutine:       false,
			wantErr:         true,
			wantErrContains: "failed to upload image",
		},
		{
			name:      "repo update error — rolls back new image",
			req:       &entity.UpdateCatalogRequest{ID: itemID},
			imageData: imageData,
			setupRepo: func(r *mockRepo.IClothingRepository) {
				existing := newCatalog(itemID)
				existing.ImageURL = oldURL
				r.On("GetItemByID", mock.Anything, itemID).Return(existing, nil)
				r.On("UpdateItem", mock.Anything, mock.Anything).Return(nil, errors.New("db error"))
			},
			setupObj: func(o *mockObject.IUploadService) {
				o.On("UploadImage", mock.Anything, fileName, imageData).Return(newURL, nil)
				o.On("DeleteImage", mock.Anything, oldURL).Return(true, nil)
				o.On("DeleteImage", mock.Anything, newURL).Return(true, nil)
			},
			goroutine:       false,
			wantErr:         true,
			wantErrContains: "failed to update item",
		},
		{
			name:      "kafka send error — logged but item returned",
			req:       &entity.UpdateCatalogRequest{ID: itemID, Name: "Updated"},
			imageData: nil,
			setupRepo: func(r *mockRepo.IClothingRepository) {
				r.On("UpdateItem", mock.Anything, mock.Anything).Return(catalog, nil)
			},
			setupWriter: func(w *mockKafka.KafkaWriter) {
				w.On("WriteMessages", mock.Anything, mock.Anything).Return(errors.New("kafka error"))
			},
			setupLog: func(l *mockLogger.Logger) {
				l.On("Error", mock.Anything, mock.AnythingOfType("string")).Maybe()
			},
			goroutine: true,
			wantErr:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := mockRepo.NewIClothingRepository(t)
			obj := mockObject.NewIUploadService(t)
			writer := mockKafka.NewKafkaWriter(t)
			log := mockLogger.NewLogger(t)

			var cache *redis.Client
			if tc.goroutine {
				cache = newRedis(t)
				log.On("Error", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()
			}

			if tc.setupRepo != nil {
				tc.setupRepo(repo)
			}
			if tc.setupObj != nil {
				tc.setupObj(obj)
			}
			if tc.setupWriter != nil {
				tc.setupWriter(writer)
			}
			if tc.setupLog != nil {
				tc.setupLog(log)
			}

			uc := buildCatalogUC(repo, obj, writer, log, cache)
			result, err := uc.UpdateItem(context.Background(), tc.req, fileName, tc.imageData)

			if tc.goroutine {
				time.Sleep(10 * time.Millisecond)
			}

			if tc.wantErr {
				assert.Error(t, err)
				if tc.wantErrContains != "" {
					assert.ErrorContains(t, err, tc.wantErrContains)
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, catalog, result)
			}
		})
	}
}

func TestClothingCatalogUseCase_DeleteItem(t *testing.T) {
	itemID := uuid.New()

	tests := []struct {
		name        string
		setupRepo   func(r *mockRepo.IClothingRepository)
		setupWriter func(w *mockKafka.KafkaWriter)
		setupLog    func(l *mockLogger.Logger)
		goroutine   bool
		wantErr     bool
	}{
		{
			name: "success",
			setupRepo: func(r *mockRepo.IClothingRepository) {
				r.On("DeleteItem", mock.Anything, itemID).Return(nil)
			},
			setupWriter: func(w *mockKafka.KafkaWriter) {
				w.On("WriteMessages", mock.Anything, mock.Anything).Return(nil)
			},
			goroutine: true,
			wantErr:   false,
		},
		{
			name: "repo error",
			setupRepo: func(r *mockRepo.IClothingRepository) {
				r.On("DeleteItem", mock.Anything, itemID).Return(errors.New("db error"))
			},
			goroutine: false,
			wantErr:   true,
		},
		{
			name: "kafka send error — logged but nil returned",
			setupRepo: func(r *mockRepo.IClothingRepository) {
				r.On("DeleteItem", mock.Anything, itemID).Return(nil)
			},
			setupWriter: func(w *mockKafka.KafkaWriter) {
				w.On("WriteMessages", mock.Anything, mock.Anything).Return(errors.New("kafka error"))
			},
			setupLog: func(l *mockLogger.Logger) {
				l.On("Error", mock.Anything, mock.AnythingOfType("string")).Maybe()
			},
			goroutine: true,
			wantErr:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := mockRepo.NewIClothingRepository(t)
			obj := mockObject.NewIUploadService(t)
			writer := mockKafka.NewKafkaWriter(t)
			log := mockLogger.NewLogger(t)

			var cache *redis.Client
			if tc.goroutine {
				cache = newRedis(t)
				log.On("Error", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()
			}

			if tc.setupRepo != nil {
				tc.setupRepo(repo)
			}
			if tc.setupWriter != nil {
				tc.setupWriter(writer)
			}
			if tc.setupLog != nil {
				tc.setupLog(log)
			}

			uc := buildCatalogUC(repo, obj, writer, log, cache)
			err := uc.DeleteItem(context.Background(), itemID)

			if tc.goroutine {
				time.Sleep(10 * time.Millisecond)
			}

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
