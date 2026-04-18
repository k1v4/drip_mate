package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/k1v4/drip_mate/internal/entity"
	"github.com/k1v4/drip_mate/pkg/DataBase"
	"github.com/k1v4/drip_mate/pkg/DataBase/postgres"
)

type ClothingRepository struct {
	*postgres.Postgres
}

func NewClothingRepository(pg *postgres.Postgres) *ClothingRepository {
	return &ClothingRepository{pg}
}

func (cr *ClothingRepository) GetItemByID(ctx context.Context, id uuid.UUID) (*entity.Catalog, error) {
	sqlReq, args, err := cr.Builder.
		Select("id", "name", "category_id", "gender", "season_id", "formality_level", "material", "image_url", "created_at", "updated_at", "is_deleted").
		From("catalog").
		Where("id = ? AND is_deleted = false", id).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	row, err := cr.Pool.Query(ctx, sqlReq, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("item not found")
		}
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	var item entity.Catalog
	item, err = pgx.CollectOneRow(row, pgx.RowToStructByName[entity.Catalog])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, DataBase.ErrCatalogItemNotFound
		}
		return nil, fmt.Errorf("failed to collect item: %w", err)
	}

	return new(item), nil
}

func (cr *ClothingRepository) CreateItem(ctx context.Context, item *entity.Catalog) (*entity.Catalog, error) {
	var created entity.Catalog

	if err := postgres.WithTx(ctx, cr.Pool, func(tx pgx.Tx) error {
		sqlReq, args, err := cr.Builder.
			Insert("catalog").
			Columns("name", "category_id", "gender", "season_id", "formality_level", "material", "image_url").
			Values(item.Name, item.CategoryID, item.Gender, item.SeasonID, item.FormalityLevel, item.Material, item.ImageURL).
			Suffix("RETURNING id, name, category_id, gender, season_id, formality_level, material, image_url, created_at, updated_at, is_deleted").
			ToSql()
		if err != nil {
			return fmt.Errorf("failed to build query: %w", err)
		}

		row, err := tx.Query(ctx, sqlReq, args...)
		if err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}

		created, err = pgx.CollectOneRow(row, pgx.RowToStructByName[entity.Catalog])
		if err != nil {
			return fmt.Errorf("failed to collect item: %w", err)
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed to create catalog item: %w", err)
	}

	return new(created), nil
}

func (cr *ClothingRepository) UpdateItem(ctx context.Context, item *entity.Catalog) (*entity.Catalog, error) {
	var updated entity.Catalog

	if err := postgres.WithTx(ctx, cr.Pool, func(tx pgx.Tx) error {
		sqlReq, args, err := cr.Builder.
			Update("catalog").
			Set("name", item.Name).
			Set("category_id", item.CategoryID).
			Set("gender", item.Gender).
			Set("season_id", item.SeasonID).
			Set("formality_level", item.FormalityLevel).
			Set("material", item.Material).
			Set("image_url", item.ImageURL).
			Set("updated_at", squirrel.Expr("NOW()")).
			Where("id = ? AND is_deleted = false", item.ID).
			Suffix("RETURNING id, name, category_id, gender, season_id, formality_level, material, image_url, created_at, updated_at, is_deleted").
			ToSql()
		if err != nil {
			return fmt.Errorf("failed to build query: %w", err)
		}

		row, err := tx.Query(ctx, sqlReq, args...)
		if err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}

		updated, err = pgx.CollectOneRow(row, pgx.RowToStructByName[entity.Catalog])
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return DataBase.ErrCatalogItemNotFound
			}
			return fmt.Errorf("failed to collect item: %w", err)
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed to update catalog item: %w", err)
	}

	return new(updated), nil
}

func (cr *ClothingRepository) DeleteItem(ctx context.Context, id uuid.UUID) error {
	if err := postgres.WithTx(ctx, cr.Pool, func(tx pgx.Tx) error {
		sqlReq, args, err := cr.Builder.
			Update("catalog").
			Set("is_deleted", true).
			Set("updated_at", "NOW()").
			Where("id = ? AND is_deleted = false", id).
			ToSql()
		if err != nil {
			return fmt.Errorf("failed to build query: %w", err)
		}

		result, err := tx.Exec(ctx, sqlReq, args...)
		if err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}

		if result.RowsAffected() == 0 {
			return DataBase.ErrCatalogItemNotFound
		}

		return nil
	}); err != nil {
		return fmt.Errorf("failed to delete catalog item: %w", err)
	}

	return nil
}
