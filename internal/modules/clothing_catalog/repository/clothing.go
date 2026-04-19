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
	sqlReq := `
		SELECT
			c.id,
			c.name,
			ca.name AS category,
			c.category_id,
			c.gender,
			c.season_id,
			s.name  AS season,
			c.formality_level,
			c.material,
			c.image_url,
			c.created_at,
			c.updated_at,
			c.is_deleted,
			ARRAY_AGG(DISTINCT col.name) FILTER (WHERE col.name IS NOT NULL) AS colors,
			ARRAY_AGG(DISTINCT st.name)  FILTER (WHERE st.name  IS NOT NULL) AS styles
		FROM catalog c
				 LEFT JOIN season   s   ON s.id   = c.season_id
				 LEFT JOIN category ca  ON ca.id  = c.category_id
				 LEFT JOIN color_catalog cc  ON cc.catalog_id = c.id
				 LEFT JOIN color_types   col ON col.id         = cc.color_id
				 LEFT JOIN style_catalog sc  ON sc.catalog_id = c.id
				 LEFT JOIN style_types        st  ON st.id          = sc.style_id
		WHERE c.id = $1 AND c.is_deleted = false
		GROUP BY c.id, ca.name, s.name
    `

	var item entity.Catalog
	err := cr.Pool.QueryRow(ctx, sqlReq, id).Scan(
		&item.ID,
		&item.Name,
		&item.Category,
		&item.CategoryID,
		&item.Gender,
		&item.SeasonID,
		&item.Season,
		&item.FormalityLevel,
		&item.Material,
		&item.ImageURL,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.IsDeleted,
		&item.Colors,
		&item.Styles,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, DataBase.ErrCatalogItemNotFound
		}
		return nil, fmt.Errorf("failed to collect item: %w", err)
	}

	return new(item), nil
}

func (cr *ClothingRepository) CreateItem(ctx context.Context, req *entity.CreateCatalogRequest) (*entity.Catalog, error) {
	var created entity.Catalog

	if err := postgres.WithTx(ctx, cr.Pool, func(tx pgx.Tx) error {
		// 1. вставляем основную запись
		err := tx.QueryRow(ctx, `
            INSERT INTO catalog (name, category_id, gender, season_id, formality_level, material, image_url)
            VALUES ($1, $2, $3, $4, $5, $6, $7)
            RETURNING id, name, category_id, gender, season_id, formality_level, material, image_url, created_at, updated_at, is_deleted
        `, req.Name, req.CategoryID, req.Gender, req.SeasonID, req.FormalityLevel, req.Material, req.ImageURL).Scan(
			&created.ID,
			&created.Name,
			&created.CategoryID,
			&created.Gender,
			&created.SeasonID,
			&created.FormalityLevel,
			&created.Material,
			&created.ImageURL,
			&created.CreatedAt,
			&created.UpdatedAt,
			&created.IsDeleted,
		)
		if err != nil {
			return fmt.Errorf("failed to insert catalog item: %w", err)
		}

		if len(req.ColorIDs) > 0 {
			colorInsert := cr.Builder.
				Insert("color_catalog").
				Columns("catalog_id", "color_id")
			for _, colorID := range req.ColorIDs {
				colorInsert = colorInsert.Values(created.ID, colorID)
			}
			sqlReq, args, err := colorInsert.ToSql()
			if err != nil {
				return fmt.Errorf("failed to build colors query: %w", err)
			}
			if _, err = tx.Exec(ctx, sqlReq, args...); err != nil {
				return fmt.Errorf("failed to insert colors: %w", err)
			}
		} else {
			return DataBase.ErrNoColors
		}

		if len(req.StyleIDs) > 0 {
			styleInsert := cr.Builder.Insert("style_catalog").Columns("catalog_id", "style_id")
			for _, styleID := range req.StyleIDs {
				styleInsert = styleInsert.Values(created.ID, styleID)
			}
			sqlReq, args, err := styleInsert.ToSql()
			if err != nil {
				return fmt.Errorf("failed to build styles query: %w", err)
			}
			if _, err = tx.Exec(ctx, sqlReq, args...); err != nil {
				return fmt.Errorf("failed to insert styles: %w", err)
			}
		} else {
			return DataBase.ErrNoStyles
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed to create catalog item: %w", err)
	}

	return &created, nil
}

func (cr *ClothingRepository) UpdateItem(ctx context.Context, req *entity.UpdateCatalogRequest) (*entity.Catalog, error) {
	var updated entity.Catalog

	if err := postgres.WithTx(ctx, cr.Pool, func(tx pgx.Tx) error {
		// обновляем основную запись
		q := cr.Builder.
			Update("catalog").
			Set("updated_at", squirrel.Expr("NOW()")).
			Where("id = ? AND is_deleted = false", req.ID).
			Suffix("RETURNING id, name, category_id, gender, season_id, formality_level, material, image_url, created_at, updated_at, is_deleted")

		if req.Name != "" {
			q = q.Set("name", req.Name)
		}
		if req.Gender != nil {
			q = q.Set("gender", req.Gender)
		}
		if req.FormalityLevel != nil {
			q = q.Set("formality_level", req.FormalityLevel)
		}
		if req.Material != nil {
			q = q.Set("material", req.Material)
		}
		if req.ImageURL != "" {
			q = q.Set("image_url", req.ImageURL)
		}
		if req.CategoryID != 0 {
			q = q.Set("category_id", req.CategoryID)
		}
		if req.SeasonID != 0 {
			q = q.Set("season_id", req.SeasonID)
		}

		sqlReq, args, err := q.ToSql()
		if err != nil {
			return fmt.Errorf("failed to build query: %w", err)
		}

		err = tx.QueryRow(ctx, sqlReq, args...).Scan(
			&updated.ID,
			&updated.Name,
			&updated.CategoryID,
			&updated.Gender,
			&updated.SeasonID,
			&updated.FormalityLevel,
			&updated.Material,
			&updated.ImageURL,
			&updated.CreatedAt,
			&updated.UpdatedAt,
			&updated.IsDeleted,
		)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return DataBase.ErrCatalogItemNotFound
			}
			return fmt.Errorf("failed to collect item: %w", err)
		}

		// цвета только если переданы в запросе
		if len(req.ColorIDs) > 0 {
			if _, err = tx.Exec(ctx, `DELETE FROM color_catalog WHERE catalog_id = $1`, req.ID); err != nil {
				return fmt.Errorf("failed to delete colors: %w", err)
			}

			colorInsert := cr.Builder.
				Insert("color_catalog").
				Columns("catalog_id", "color_id")
			for _, colorID := range req.ColorIDs {
				colorInsert = colorInsert.Values(updated.ID, colorID)
			}
			sqlReq, args, err := colorInsert.ToSql()
			if err != nil {
				return fmt.Errorf("failed to build colors query: %w", err)
			}
			if _, err = tx.Exec(ctx, sqlReq, args...); err != nil {
				return fmt.Errorf("failed to insert colors: %w", err)
			}
		}

		// стили только если переданы в запросе
		if len(req.StyleIDs) > 0 {
			if _, err = tx.Exec(ctx, `DELETE FROM style_catalog WHERE catalog_id = $1`, req.ID); err != nil {
				return fmt.Errorf("failed to delete styles: %w", err)
			}

			styleInsert := cr.Builder.
				Insert("style_catalog").
				Columns("catalog_id", "style_id")
			for _, styleID := range req.StyleIDs {
				styleInsert = styleInsert.Values(updated.ID, styleID)
			}
			sqlReq, args, err := styleInsert.ToSql()
			if err != nil {
				return fmt.Errorf("failed to build styles query: %w", err)
			}
			if _, err = tx.Exec(ctx, sqlReq, args...); err != nil {
				return fmt.Errorf("failed to insert styles: %w", err)
			}
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed to update catalog item: %w", err)
	}

	return &updated, nil
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
