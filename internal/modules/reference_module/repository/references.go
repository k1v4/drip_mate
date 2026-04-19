package repository

import (
	"context"
	"fmt"

	"github.com/k1v4/drip_mate/internal/entity"
	"github.com/k1v4/drip_mate/pkg/DataBase/postgres"
)

type ReferenceRepository struct {
	*postgres.Postgres
}

func NewReferenceRepository(pg *postgres.Postgres) *ReferenceRepository {
	return &ReferenceRepository{pg}
}

func (r *ReferenceRepository) GetStyles(ctx context.Context) ([]entity.StyleType, error) {
	rows, err := r.Pool.Query(ctx, `SELECT id, name FROM style_types ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("failed to get styles: %w", err)
	}
	defer rows.Close()

	var styles []entity.StyleType
	for rows.Next() {
		var s entity.StyleType
		if err := rows.Scan(&s.ID, &s.Name); err != nil {
			return nil, fmt.Errorf("failed to scan style: %w", err)
		}
		styles = append(styles, s)
	}
	return styles, nil
}

func (r *ReferenceRepository) GetColors(ctx context.Context) ([]entity.ColorType, error) {
	rows, err := r.Pool.Query(ctx, `SELECT id, name FROM color_types ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("failed to get colors: %w", err)
	}
	defer rows.Close()

	var colors []entity.ColorType
	for rows.Next() {
		var c entity.ColorType
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			return nil, fmt.Errorf("failed to scan color: %w", err)
		}
		colors = append(colors, c)
	}
	return colors, nil
}

func (r *ReferenceRepository) GetMusics(ctx context.Context) ([]entity.MusicType, error) {
	rows, err := r.Pool.Query(ctx, `SELECT id, name FROM music ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("failed to get musics: %w", err)
	}
	defer rows.Close()

	var musics []entity.MusicType
	for rows.Next() {
		var m entity.MusicType
		if err := rows.Scan(&m.ID, &m.Name); err != nil {
			return nil, fmt.Errorf("failed to scan music: %w", err)
		}
		musics = append(musics, m)
	}
	return musics, nil
}

func (r *ReferenceRepository) GetCategories(ctx context.Context) ([]entity.Category, error) {
	rows, err := r.Pool.Query(ctx, `SELECT id, name FROM category ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}
	defer rows.Close()

	var categories []entity.Category
	for rows.Next() {
		var c entity.Category
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, c)
	}
	return categories, nil
}

func (r *ReferenceRepository) GetSeasons(ctx context.Context) ([]entity.Season, error) {
	rows, err := r.Pool.Query(ctx, `SELECT id, name FROM season ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("failed to get seasons: %w", err)
	}
	defer rows.Close()

	var seasons []entity.Season
	for rows.Next() {
		var s entity.Season
		if err := rows.Scan(&s.ID, &s.Name); err != nil {
			return nil, fmt.Errorf("failed to scan season: %w", err)
		}
		seasons = append(seasons, s)
	}
	return seasons, nil
}
