package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/k1v4/drip_mate/internal/entity"
	"github.com/k1v4/drip_mate/pkg/adapter"
)

type IRecommendationRepository interface {
	GetUserProfile(ctx context.Context, userID uuid.UUID) (*entity.UserProfile, string, error)
}

type RecommendationsUseCase struct {
	recommendationsRepo IRecommendationRepository
	weatherAdapter      *adapter.OpenWeatherAdapter
	ml                  *adapter.MLClient
}

func NewRecommendationsUseCase(recommendationsRepo IRecommendationRepository, weatherAdapter *adapter.OpenWeatherAdapter, ml *adapter.MLClient) *RecommendationsUseCase {
	return &RecommendationsUseCase{
		recommendationsRepo: recommendationsRepo,
		weatherAdapter:      weatherAdapter,
		ml:                  ml,
	}
}

func (uc *RecommendationsUseCase) GetUserRecommendation(ctx context.Context, formality int, userID uuid.UUID) ([]entity.OutfitItem, error) {
	profile, city, err := uc.recommendationsRepo.GetUserProfile(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("GetUserProfile: %w", err)
	}

	weather, err := uc.weatherAdapter.GetCurrentWeather(ctx, city)
	if err != nil {
		return nil, fmt.Errorf("failed to get weather: %w", err)
	}

	season := seasonFromContext(weather.Temperature, int(time.Now().Month()))

	items, err := uc.ml.GetRecommendation(ctx, &entity.RequestData{
		UserProfile: *profile,
		Context: entity.Context{
			Season:    season,
			Formality: formality,
		},
		K: 20, // TODO в конфиг
	})
	if err != nil {
		return nil, fmt.Errorf("recommend: ml client: %w", err)
	}

	return items, nil
}

func seasonFromContext(tempC float64, month int) string {
	switch {
	case tempC >= 20:
		return "summer"
	case tempC < 0:
		return "winter"
	case month >= 3 && month <= 5:
		return "spring"
	default:
		return "autumn"
	}
}
