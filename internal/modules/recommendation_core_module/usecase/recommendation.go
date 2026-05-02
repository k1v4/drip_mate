package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/k1v4/drip_mate/internal/entity"
	v1 "github.com/k1v4/drip_mate/internal/modules/clothing_catalog/controller/http/v1"
	"github.com/k1v4/drip_mate/pkg/adapter"
	"github.com/k1v4/drip_mate/pkg/logger"
)

type IRecommendationRepository interface {
	GetUserProfile(ctx context.Context, userID uuid.UUID) (*entity.UserProfile, string, error)
	SaveRecommendationLog(
		ctx context.Context,
		userID uuid.UUID,
		outfits []uuid.UUID,
		modelPhase string,
		reqContext *entity.RecommendationContext,
	) (int, error)
}

type RecommendationsUseCase struct {
	recommendationsRepo IRecommendationRepository
	clothingUseCase     v1.IClothingUseCase
	weatherAdapter      *adapter.OpenWeatherAdapter
	ml                  *adapter.MLClient
	l                   logger.Logger
}

func NewRecommendationsUseCase(recommendationsRepo IRecommendationRepository, weatherAdapter *adapter.OpenWeatherAdapter, ml *adapter.MLClient, clothingUseCase v1.IClothingUseCase, l logger.Logger) *RecommendationsUseCase {
	return &RecommendationsUseCase{
		recommendationsRepo: recommendationsRepo,
		weatherAdapter:      weatherAdapter,
		ml:                  ml,
		clothingUseCase:     clothingUseCase,
		l:                   l,
	}
}

func (uc *RecommendationsUseCase) GetUserRecommendation(ctx context.Context, formality int, userID uuid.UUID) (*entity.RecommendationsCatalogRequest, error) {
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
		Context: entity.RecommendationContext{
			Season:    season,
			Formality: formality,
		},
		K: 20, // TODO в конфиг
	})
	if err != nil {
		return nil, fmt.Errorf("recommend: ml client: %w", err)
	}

	reqContext := &entity.RecommendationContext{
		Season:      season,
		Formality:   formality,
		Styles:      profile.Styles,
		Colors:      profile.Colors,
		MusicGenres: profile.MusicGenres,
		Gender:      profile.GenderPref,
	}

	resultItems := make([]entity.Catalog, 0, len(items))
	outfitsForLog := make([]uuid.UUID, 0, len(items))
	// TODO переделать под один запрос в бд
	for _, item := range items {
		uuidItem, err := uuid.Parse(item.ItemID)
		if err != nil {
			uc.l.Error(ctx, fmt.Sprintf("failed to parse item id: %s", item.ItemID))
			continue
		}

		clothingItem, err := uc.clothingUseCase.GetItemByID(ctx, uuidItem)
		if err != nil {
			uc.l.Error(ctx, fmt.Sprintf("failed to get clothing item: %s", uuidItem))
			continue
		}

		resultItems = append(resultItems, *clothingItem)
		outfitsForLog = append(outfitsForLog, uuidItem)
	}

	logID, err := uc.recommendationsRepo.SaveRecommendationLog(ctx, userID, outfitsForLog, "fm", reqContext)
	if err != nil {
		uc.l.Error(ctx, fmt.Sprintf("failed to save recommendation log: %s", err))
	}

	return &entity.RecommendationsCatalogRequest{
		Catalog: resultItems,
		LogID:   logID,
	}, nil
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
