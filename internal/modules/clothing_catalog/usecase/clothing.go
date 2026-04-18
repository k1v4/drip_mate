package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/k1v4/drip_mate/internal/entity"
)

type IClothingRepository interface {
	GetItemByID(ctx context.Context, id uuid.UUID) (*entity.Catalog, error)
	CreateItem(ctx context.Context, item *entity.Catalog) (*entity.Catalog, error)
	UpdateItem(ctx context.Context, item *entity.Catalog) (*entity.Catalog, error)
	DeleteItem(ctx context.Context, id uuid.UUID) error
}

type ClothingCatalogUseCase struct {
	repoClothing IClothingRepository
}

func NewClothingCatalogUseCase(repoClothing IClothingRepository) *ClothingCatalogUseCase {
	return &ClothingCatalogUseCase{
		repoClothing: repoClothing,
	}
}

func (uc *ClothingCatalogUseCase) GetItemByID(ctx context.Context, id uuid.UUID) (*entity.Catalog, error) {
	const op = "ClothingCatalogUseCase.GetItemByID"

	item, err := uc.repoClothing.GetItemByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return item, nil
}

func (uc *ClothingCatalogUseCase) CreateItem(ctx context.Context, item *entity.Catalog) (uuid.UUID, error) {
	const op = "ClothingCatalogUseCase.CreateItem"

	createdItem, err := uc.repoClothing.CreateItem(ctx, item)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	return createdItem.ID, nil
}

func (uc *ClothingCatalogUseCase) UpdateItem(ctx context.Context, item *entity.Catalog) (uuid.UUID, error) {
	const op = "ClothingCatalogUseCase.UpdateItem"

	updatedItem, err := uc.repoClothing.UpdateItem(ctx, item)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	return updatedItem.ID, nil
}

func (uc *ClothingCatalogUseCase) DeleteItem(ctx context.Context, id uuid.UUID) error {
	const op = "ClothingCatalogUseCase.DeleteItem"

	err := uc.repoClothing.DeleteItem(ctx, id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
