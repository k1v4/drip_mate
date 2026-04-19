package usecase

import (
	"context"

	"github.com/k1v4/drip_mate/internal/entity"
)

type IReferenceRepository interface {
	GetStyles(ctx context.Context) ([]entity.StyleType, error)
	GetColors(ctx context.Context) ([]entity.ColorType, error)
	GetMusics(ctx context.Context) ([]entity.MusicType, error)
	GetCategories(ctx context.Context) ([]entity.Category, error)
	GetSeasons(ctx context.Context) ([]entity.Season, error)
}

type ReferenceUseCase struct {
	repo IReferenceRepository
}

func NewReferenceUseCase(repo IReferenceRepository) *ReferenceUseCase {
	return &ReferenceUseCase{repo: repo}
}

func (uc *ReferenceUseCase) GetStyles(ctx context.Context) ([]entity.StyleType, error) {
	return uc.repo.GetStyles(ctx)
}

func (uc *ReferenceUseCase) GetColors(ctx context.Context) ([]entity.ColorType, error) {
	return uc.repo.GetColors(ctx)
}

func (uc *ReferenceUseCase) GetMusics(ctx context.Context) ([]entity.MusicType, error) {
	return uc.repo.GetMusics(ctx)
}

func (uc *ReferenceUseCase) GetCategories(ctx context.Context) ([]entity.Category, error) {
	return uc.repo.GetCategories(ctx)
}

func (uc *ReferenceUseCase) GetSeasons(ctx context.Context) ([]entity.Season, error) {
	return uc.repo.GetSeasons(ctx)
}
