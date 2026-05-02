package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/k1v4/drip_mate/internal/entity"
	objectTransport "github.com/k1v4/drip_mate/internal/modules/object_gateway/transport/grpc"
	redispkg "github.com/k1v4/drip_mate/pkg/DataBase/redis"
	"github.com/k1v4/drip_mate/pkg/kafkaPkg"
	"github.com/k1v4/drip_mate/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type IClothingRepository interface {
	GetItemByID(ctx context.Context, id uuid.UUID) (*entity.Catalog, error)
	CreateItem(ctx context.Context, req *entity.CreateCatalogRequest) (*entity.Catalog, error)
	UpdateItem(ctx context.Context, req *entity.UpdateCatalogRequest) (*entity.Catalog, error)
	DeleteItem(ctx context.Context, id uuid.UUID) error
	GetAllItems(
		ctx context.Context,
		limit, offset int,
	) ([]entity.Catalog, int, error)
}

type ClothingCatalogUseCase struct {
	repoClothing  IClothingRepository
	objectService objectTransport.IUploadService
	kafkaProducer *kafkaPkg.Producer[entity.CatalogEvent]
	l             logger.Logger
	cache         *redis.Client
}

func NewClothingCatalogUseCase(
	repoClothing IClothingRepository,
	objectService objectTransport.IUploadService,
	kafkaProducer *kafkaPkg.Producer[entity.CatalogEvent],
	l logger.Logger,
	cache *redis.Client,
) *ClothingCatalogUseCase {
	return &ClothingCatalogUseCase{
		repoClothing:  repoClothing,
		objectService: objectService,
		kafkaProducer: kafkaProducer,
		l:             l,
		cache:         cache,
	}
}

func (uc *ClothingCatalogUseCase) GetItemByID(ctx context.Context, id uuid.UUID) (*entity.Catalog, error) {
	const op = "ClothingCatalogUseCase.GetItemByID"

	val, err := uc.cache.Get(ctx, redispkg.GetCatalogItemKey(id)).Bytes()
	if err == nil {
		var cached entity.Catalog
		if err = json.Unmarshal(val, &cached); err == nil {
			return &cached, nil
		}
	}

	item, err := uc.repoClothing.GetItemByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		if data, err := json.Marshal(item); err == nil {
			if err = uc.cache.Set(bgCtx, redispkg.GetCatalogItemKey(id), data, time.Hour).Err(); err != nil {
				uc.l.Error(bgCtx, "failed to set cache", zap.Error(err))
			}
		}
	}()

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

	err = uc.kafkaProducer.Send(ctx, entity.CatalogEvent{
		Type:    entity.CatalogCreated,
		Payload: item.ID,
	})
	if err != nil {
		uc.l.Error(ctx, fmt.Sprintf("failed to send create catalog event to ml: %v", err))
	}

	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		if data, err := json.Marshal(item); err == nil && item.ID != uuid.Nil {
			if err = uc.cache.Set(bgCtx, redispkg.GetCatalogItemKey(item.ID), data, time.Hour).Err(); err != nil {
				uc.l.Error(bgCtx, "failed to set cache", zap.Error(err))
			}
		}
	}()

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

	err = uc.kafkaProducer.Send(ctx, entity.CatalogEvent{
		Type:    entity.CatalogUpdated,
		Payload: item.ID,
	})
	if err != nil {
		uc.l.Error(ctx, fmt.Sprintf("failed to send update catalog event to ml: %v", err))
	}

	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		if err := uc.cache.Del(bgCtx, redispkg.GetCatalogItemKey(req.ID)).Err(); err != nil {
			uc.l.Error(bgCtx, "failed to invalidate cache", zap.Error(err))
		}
	}()

	return item, nil
}

func (uc *ClothingCatalogUseCase) DeleteItem(ctx context.Context, id uuid.UUID) error {
	const op = "ClothingCatalogUseCase.DeleteItem"

	err := uc.repoClothing.DeleteItem(ctx, id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	err = uc.kafkaProducer.Send(ctx, entity.CatalogEvent{
		Type:    entity.CatalogDeleted,
		Payload: id,
	})
	if err != nil {
		uc.l.Error(ctx, fmt.Sprintf("failed to send delete catalog event to ml: %v", err))
	}

	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		if err := uc.cache.Del(bgCtx, redispkg.GetCatalogItemKey(id)).Err(); err != nil {
			uc.l.Error(bgCtx, "failed to invalidate cache", zap.Error(err))
		}
	}()

	return nil
}

func (uc *ClothingCatalogUseCase) GetAllItems(
	ctx context.Context,
	limit, offset int,
) ([]entity.Catalog, int, error) {
	const op = "ClothingCatalogUseCase.GetAllItems"

	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	items, total, err := uc.repoClothing.GetAllItems(ctx, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("%s: %w", op, err)
	}

	return items, total, nil
}
