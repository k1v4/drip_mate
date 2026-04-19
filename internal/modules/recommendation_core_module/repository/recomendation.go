package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/k1v4/drip_mate/internal/entity"
	"github.com/k1v4/drip_mate/pkg/DataBase/postgres"
	"github.com/lib/pq"
)

type RecommendationsRepository struct {
	*postgres.Postgres
}

func NewRecommendationsRepository(pg *postgres.Postgres) *RecommendationsRepository {
	return &RecommendationsRepository{pg}
}

type userProfileRow struct {
	Gender      string         `db:"gender"`
	Styles      pq.StringArray `db:"styles"`
	Colors      pq.StringArray `db:"colors"`
	MusicGenres pq.StringArray `db:"music_genres"`
	City        string         `db:"city"`
}

func (r *RecommendationsRepository) GetUserProfile(ctx context.Context, userID uuid.UUID) (*entity.UserProfile, string, error) {
	const query = `
        SELECT
            u.gender,
            COALESCE(array_agg(DISTINCT st.name) FILTER (WHERE st.name IS NOT NULL), '{}') AS styles,
            COALESCE(array_agg(DISTINCT ct.name) FILTER (WHERE ct.name IS NOT NULL), '{}') AS colors,
            COALESCE(array_agg(DISTINCT m.name)  FILTER (WHERE m.name  IS NOT NULL), '{}') AS music_genres,
            u.city
        FROM users u
        LEFT JOIN style_user  su ON su.user_id  = u.id
        LEFT JOIN style_types st ON st.id       = su.style_id
        LEFT JOIN color_user  cu ON cu.user_id  = u.id
        LEFT JOIN color_types ct ON ct.id       = cu.color_id
        LEFT JOIN music_user  mu ON mu.user_id  = u.id
        LEFT JOIN music       m  ON m.id        = mu.music_id
        WHERE u.id = $1
        GROUP BY u.id, u.gender, u.city`

	var row userProfileRow
	err := r.Pool.QueryRow(ctx, query, userID).Scan(
		&row.Gender,
		&row.Styles,
		&row.Colors,
		&row.MusicGenres,
		&row.City,
	)
	if err != nil {
		return nil, "", fmt.Errorf("get user profile: %w", err)
	}

	profile := entity.UserProfile{
		GenderPref:  row.Gender,
		Styles:      []string(row.Styles),
		Colors:      []string(row.Colors),
		MusicGenres: []string(row.MusicGenres),
	}

	return new(profile), row.City, nil
}
