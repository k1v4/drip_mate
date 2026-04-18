package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/k1v4/drip_mate/internal/entity"
	objectTransport "github.com/k1v4/drip_mate/internal/modules/object_gateway/transport/grpc"
)

type IClothingRepository interface {
	GetItemByID(ctx context.Context, id uuid.UUID) (*entity.Catalog, error)
	CreateItem(ctx context.Context, req *entity.CreateCatalogRequest) (*entity.Catalog, error)
	UpdateItem(ctx context.Context, req *entity.UpdateCatalogRequest) (*entity.Catalog, error)
	DeleteItem(ctx context.Context, id uuid.UUID) error
}

type ClothingCatalogUseCase struct {
	repoClothing  IClothingRepository
	objectService objectTransport.IUploadService
}

func NewClothingCatalogUseCase(
	repoClothing IClothingRepository,
	objectService objectTransport.IUploadService,
) *ClothingCatalogUseCase {
	return &ClothingCatalogUseCase{
		repoClothing:  repoClothing,
		objectService: objectService,
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

func (uc *ClothingCatalogUseCase) CreateItem(ctx context.Context, req *entity.CreateCatalogRequest, fileName string, imageData []byte) (*entity.Catalog, error) {
	imageURL, err := uc.objectService.UploadImage(ctx, fileName, imageData)
	if err != nil {
		return nil, fmt.Errorf("failed to upload image: %w", err)
	}
	req.ImageURL = imageURL

	item, err := uc.repoClothing.CreateItem(ctx, req)
	if err != nil {
		_, _ = uc.objectService.DeleteImage(ctx, fileName)
		return nil, fmt.Errorf("failed to create item: %w", err)
	}

	return item, nil
}

func (uc *ClothingCatalogUseCase) UpdateItem(ctx context.Context, req *entity.UpdateCatalogRequest, fileName string, imageData []byte) (*entity.Catalog, error) {
	// если файл пришёл — загружаем новый
	if len(imageData) > 0 {
		// достаём текущий айтем чтобы знать старый URL
		current, err := uc.repoClothing.GetItemByID(ctx, req.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get current item: %w", err)
		}

		newURL, err := uc.objectService.UploadImage(ctx, fileName, imageData)
		if err != nil {
			return nil, fmt.Errorf("failed to upload image: %w", err)
		}

		// удаляем старое только после успешной загрузки нового
		if current.ImageURL != "" {
			_, _ = uc.objectService.DeleteImage(ctx, current.ImageURL)
		}

		req.ImageURL = newURL
	}

	item, err := uc.repoClothing.UpdateItem(ctx, req)
	if err != nil {
		// если не удалось сохранить — откатываем новое изображение
		if req.ImageURL != "" {
			_, _ = uc.objectService.DeleteImage(ctx, req.ImageURL)
		}
		return nil, fmt.Errorf("failed to update item: %w", err)
	}

	return item, nil
}

func (uc *ClothingCatalogUseCase) DeleteItem(ctx context.Context, id uuid.UUID) error {
	const op = "ClothingCatalogUseCase.DeleteItem"

	err := uc.repoClothing.DeleteItem(ctx, id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
