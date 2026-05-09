package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/k1v4/drip_mate/internal/entity"
	mockCatalog "github.com/k1v4/drip_mate/mocks/internal_/modules/clothing_catalog/controller/http/v1"
	mockRepo "github.com/k1v4/drip_mate/mocks/internal_/modules/recommendation_core_module/usecase"
	mockAdapter "github.com/k1v4/drip_mate/mocks/pkg/adapter"
	mockLogger "github.com/k1v4/drip_mate/mocks/pkg/logger"
	redispkg "github.com/k1v4/drip_mate/pkg/DataBase/redis"
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

func buildUC(
	repo *mockRepo.IRecommendationRepository,
	weather *mockAdapter.WeatherProvider,
	ml *mockAdapter.MLProvider,
	catalog *mockCatalog.IClothingUseCase,
	log *mockLogger.Logger,
	cache *redis.Client,
) *RecommendationsUseCase {
	return NewRecommendationsUseCase(repo, weather, ml, catalog, log, cache)
}

func defaultProfile() *entity.UserProfile {
	return &entity.UserProfile{
		GenderPref:  "male",
		Styles:      []string{"casual"},
		Colors:      []string{"black"},
		MusicGenres: []string{"rock"},
	}
}

func defaultWeather() *entity.Weather {
	return &entity.Weather{
		City:        "Moscow",
		Temperature: 22.0, // summer
	}
}

func defaultOutfitItems(ids []uuid.UUID) []entity.OutfitItem {
	items := make([]entity.OutfitItem, len(ids))
	for i, id := range ids {
		items[i] = entity.OutfitItem{ItemID: id.String(), Score: 0.9}
	}
	return items
}

func defaultCatalog(id uuid.UUID) *entity.Catalog {
	return &entity.Catalog{
		ID:             id,
		Name:           "Test jacket",
		CategoryID:     1,
		Gender:         new("male"),
		SeasonID:       1,
		FormalityLevel: new(int16(2)),
		Material:       new("cotton"),
		ImageURL:       "https://example.com/img.jpg",
	}
}

func TestRecommendationsUseCase_GetUserRecommendation(t *testing.T) {
	userID := uuid.New()
	item1ID := uuid.New()
	item2ID := uuid.New()
	formality := 2

	tests := []struct {
		name            string
		setupRepo       func(r *mockRepo.IRecommendationRepository)
		setupWeather    func(w *mockAdapter.WeatherProvider)
		setupML         func(m *mockAdapter.MLProvider)
		setupCatalog    func(c *mockCatalog.IClothingUseCase)
		setupLog        func(l *mockLogger.Logger)
		primeCache      func(cache *redis.Client) // предзаполнить кэш погоды
		wantErr         bool
		wantErrContains string
		wantLogID       int
		wantLen         int
	}{
		{
			name: "success — cache miss, fetches weather, gets recommendations",
			setupRepo: func(r *mockRepo.IRecommendationRepository) {
				r.On("GetUserProfile", mock.Anything, userID).
					Return(defaultProfile(), "Moscow", nil)
				r.On("SaveRecommendationLog", mock.Anything, userID, mock.Anything, "fm", mock.Anything).
					Return(42, nil)
			},
			setupWeather: func(w *mockAdapter.WeatherProvider) {
				w.On("GetCurrentWeather", mock.Anything, "Moscow").
					Return(defaultWeather(), nil)
			},
			setupML: func(m *mockAdapter.MLProvider) {
				m.On("GetRecommendation", mock.Anything, mock.Anything).
					Return(defaultOutfitItems([]uuid.UUID{item1ID, item2ID}), nil)
			},
			setupCatalog: func(c *mockCatalog.IClothingUseCase) {
				c.On("GetItemByID", mock.Anything, item1ID).Return(defaultCatalog(item1ID), nil)
				c.On("GetItemByID", mock.Anything, item2ID).Return(defaultCatalog(item2ID), nil)
			},
			setupLog: func(l *mockLogger.Logger) {
				l.On("Error", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()
			},
			wantErr:   false,
			wantLogID: 42,
			wantLen:   2,
		},
		{
			name: "success — weather from cache",
			setupRepo: func(r *mockRepo.IRecommendationRepository) {
				r.On("GetUserProfile", mock.Anything, userID).
					Return(defaultProfile(), "Moscow", nil)
				r.On("SaveRecommendationLog", mock.Anything, userID, mock.Anything, "fm", mock.Anything).
					Return(10, nil)
			},
			setupML: func(m *mockAdapter.MLProvider) {
				m.On("GetRecommendation", mock.Anything, mock.Anything).
					Return(defaultOutfitItems([]uuid.UUID{item1ID}), nil)
			},
			setupCatalog: func(c *mockCatalog.IClothingUseCase) {
				c.On("GetItemByID", mock.Anything, item1ID).Return(defaultCatalog(item1ID), nil)
			},
			setupLog: func(l *mockLogger.Logger) {
				l.On("Error", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()
			},
			primeCache: func(cache *redis.Client) {
				data, _ := json.Marshal(defaultWeather())
				_ = cache.Set(context.Background(), redispkg.GetWeatherCityKey("Moscow"), data, 30*time.Minute).Err()
			},
			wantErr:   false,
			wantLogID: 10,
			wantLen:   1,
		},
		{
			name: "GetUserProfile error",
			setupRepo: func(r *mockRepo.IRecommendationRepository) {
				r.On("GetUserProfile", mock.Anything, userID).
					Return(nil, "", errors.New("db error"))
			},
			wantErr:         true,
			wantErrContains: "GetUserProfile",
		},
		{
			name: "getWeather error — weather adapter fails",
			setupRepo: func(r *mockRepo.IRecommendationRepository) {
				r.On("GetUserProfile", mock.Anything, userID).
					Return(defaultProfile(), "UnknownCity", nil)
			},
			setupWeather: func(w *mockAdapter.WeatherProvider) {
				w.On("GetCurrentWeather", mock.Anything, "UnknownCity").
					Return(nil, errors.New("city not found"))
			},
			wantErr:         true,
			wantErrContains: "getWeather",
		},
		{
			name: "ML GetRecommendation error",
			setupRepo: func(r *mockRepo.IRecommendationRepository) {
				r.On("GetUserProfile", mock.Anything, userID).
					Return(defaultProfile(), "Moscow", nil)
			},
			setupWeather: func(w *mockAdapter.WeatherProvider) {
				w.On("GetCurrentWeather", mock.Anything, "Moscow").
					Return(defaultWeather(), nil)
			},
			setupML: func(m *mockAdapter.MLProvider) {
				m.On("GetRecommendation", mock.Anything, mock.Anything).
					Return(nil, errors.New("ml unavailable"))
			},
			setupLog: func(l *mockLogger.Logger) {
				l.On("Error", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()
			},
			wantErr:         true,
			wantErrContains: "ml client",
		},
		{
			name: "invalid item id from ML — skipped, rest returned",
			setupRepo: func(r *mockRepo.IRecommendationRepository) {
				r.On("GetUserProfile", mock.Anything, userID).
					Return(defaultProfile(), "Moscow", nil)
				r.On("SaveRecommendationLog", mock.Anything, userID, mock.Anything, "fm", mock.Anything).
					Return(1, nil)
			},
			setupWeather: func(w *mockAdapter.WeatherProvider) {
				w.On("GetCurrentWeather", mock.Anything, "Moscow").
					Return(defaultWeather(), nil)
			},
			setupML: func(m *mockAdapter.MLProvider) {
				m.On("GetRecommendation", mock.Anything, mock.Anything).
					Return([]entity.OutfitItem{
						{ItemID: "not-a-uuid", Score: 0.9},
						{ItemID: item1ID.String(), Score: 0.8},
					}, nil)
			},
			setupCatalog: func(c *mockCatalog.IClothingUseCase) {
				c.On("GetItemByID", mock.Anything, item1ID).Return(defaultCatalog(item1ID), nil)
			},
			setupLog: func(l *mockLogger.Logger) {
				l.On("Error", mock.Anything, mock.AnythingOfType("string")).Maybe()
				l.On("Error", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()
			},
			wantErr:   false,
			wantLen:   1,
			wantLogID: 1,
		},
		{
			name: "GetItemByID error — item skipped, rest returned",
			setupRepo: func(r *mockRepo.IRecommendationRepository) {
				r.On("GetUserProfile", mock.Anything, userID).
					Return(defaultProfile(), "Moscow", nil)
				r.On("SaveRecommendationLog", mock.Anything, userID, mock.Anything, "fm", mock.Anything).
					Return(1, nil)
			},
			setupWeather: func(w *mockAdapter.WeatherProvider) {
				w.On("GetCurrentWeather", mock.Anything, "Moscow").
					Return(defaultWeather(), nil)
			},
			setupML: func(m *mockAdapter.MLProvider) {
				m.On("GetRecommendation", mock.Anything, mock.Anything).
					Return(defaultOutfitItems([]uuid.UUID{item1ID, item2ID}), nil)
			},
			setupCatalog: func(c *mockCatalog.IClothingUseCase) {
				c.On("GetItemByID", mock.Anything, item1ID).Return(nil, errors.New("db error"))
				c.On("GetItemByID", mock.Anything, item2ID).Return(defaultCatalog(item2ID), nil)
			},
			setupLog: func(l *mockLogger.Logger) {
				l.On("Error", mock.Anything, mock.AnythingOfType("string")).Maybe()
				l.On("Error", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()
			},
			wantErr:   false,
			wantLen:   1,
			wantLogID: 1,
		},
		{
			name: "SaveRecommendationLog error — logged but result returned",
			setupRepo: func(r *mockRepo.IRecommendationRepository) {
				r.On("GetUserProfile", mock.Anything, userID).
					Return(defaultProfile(), "Moscow", nil)
				r.On("SaveRecommendationLog", mock.Anything, userID, mock.Anything, "fm", mock.Anything).
					Return(0, errors.New("db error"))
			},
			setupWeather: func(w *mockAdapter.WeatherProvider) {
				w.On("GetCurrentWeather", mock.Anything, "Moscow").
					Return(defaultWeather(), nil)
			},
			setupML: func(m *mockAdapter.MLProvider) {
				m.On("GetRecommendation", mock.Anything, mock.Anything).
					Return(defaultOutfitItems([]uuid.UUID{item1ID}), nil)
			},
			setupCatalog: func(c *mockCatalog.IClothingUseCase) {
				c.On("GetItemByID", mock.Anything, item1ID).Return(defaultCatalog(item1ID), nil)
			},
			setupLog: func(l *mockLogger.Logger) {
				l.On("Error", mock.Anything, mock.AnythingOfType("string")).Maybe()
				l.On("Error", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()
			},
			wantErr:   false,
			wantLogID: 0,
			wantLen:   1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := mockRepo.NewIRecommendationRepository(t)
			weather := mockAdapter.NewWeatherProvider(t)
			ml := mockAdapter.NewMLProvider(t)
			catalog := mockCatalog.NewIClothingUseCase(t)
			log := mockLogger.NewLogger(t)
			cache := newRedis(t)

			if tc.primeCache != nil {
				tc.primeCache(cache)
			}
			if tc.setupRepo != nil {
				tc.setupRepo(repo)
			}
			if tc.setupWeather != nil {
				tc.setupWeather(weather)
			}
			if tc.setupML != nil {
				tc.setupML(ml)
			}
			if tc.setupCatalog != nil {
				tc.setupCatalog(catalog)
			}
			if tc.setupLog != nil {
				tc.setupLog(log)
			}

			uc := buildUC(repo, weather, ml, catalog, log, cache)
			result, err := uc.GetUserRecommendation(context.Background(), formality, userID)

			time.Sleep(10 * time.Millisecond)

			if tc.wantErr {
				assert.Error(t, err)
				if tc.wantErrContains != "" {
					assert.ErrorContains(t, err, tc.wantErrContains)
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tc.wantLogID, result.LogID)
				assert.Len(t, result.Catalog, tc.wantLen)
			}
		})
	}
}

func TestSeasonFromContext(t *testing.T) {
	tests := []struct {
		name       string
		temp       float64
		month      int
		wantSeason string
	}{
		{name: "summer — temp >= 20", temp: 25.0, month: 7, wantSeason: "summer"},
		{name: "summer — temp exactly 20", temp: 20.0, month: 7, wantSeason: "summer"},
		{name: "winter — temp < 0", temp: -5.0, month: 1, wantSeason: "winter"},
		{name: "winter — temp exactly -0.1", temp: -0.1, month: 12, wantSeason: "winter"},
		{name: "spring — month 3", temp: 10.0, month: 3, wantSeason: "spring"},
		{name: "spring — month 5", temp: 15.0, month: 5, wantSeason: "spring"},
		{name: "autumn — month 6 temp 10", temp: 10.0, month: 6, wantSeason: "autumn"},
		{name: "autumn — month 10 temp 5", temp: 5.0, month: 10, wantSeason: "autumn"},
		{name: "autumn — month 2 temp 5", temp: 5.0, month: 2, wantSeason: "autumn"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.wantSeason, seasonFromContext(tc.temp, tc.month))
		})
	}
}
