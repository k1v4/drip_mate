package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/k1v4/drip_mate/internal/entity"
	"github.com/k1v4/drip_mate/pkg/DataBase/postgres"
)

type RecommendationsRepository struct {
	*postgres.Postgres
}

func NewRecommendationsRepository(pg *postgres.Postgres) *RecommendationsRepository {
	return &RecommendationsRepository{pg}
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

	var profile entity.UserProfile
	var city string

	err := r.Pool.QueryRow(ctx, query, userID).Scan(
		&profile.GenderPref,
		&profile.Styles,
		&profile.Colors,
		&profile.MusicGenres,
		&city,
	)
	if err != nil {
		return nil, "", fmt.Errorf("get user profile: %w", err)
	}

	return &profile, city, nil
}

func (r *RecommendationsRepository) SaveRecommendationLog(ctx context.Context, userID uuid.UUID, outfits []uuid.UUID, modelPhase string, reqContext *entity.RecommendationContext) (int, error) {
	outfitsJSON, err := json.Marshal(outfits)
	if err != nil {
		return 0, fmt.Errorf("marshal outfits: %w", err)
	}

	contextJSON, err := json.Marshal(reqContext)
	if err != nil {
		return 0, fmt.Errorf("marshal context: %w", err)
	}

	var logID int
	if err = postgres.WithTx(ctx, r.Pool, func(tx pgx.Tx) error {
		err = tx.QueryRow(ctx, `
			INSERT INTO recommendation_log (user_id, outfits_shown, model_phase, request_context)
			VALUES ($1, $2, $3, $4)
			RETURNING id
		`, userID, outfitsJSON, modelPhase, contextJSON).Scan(&logID)
		if err != nil {
			return fmt.Errorf("insert recommendation_log: %w", err)
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO user_interactions (user_id, recommendation_log_id, outfit_items, event_type, context_snapshot)
			VALUES ($1, $2, $3, 'skip', $4)
		`, userID, logID, outfitsJSON, contextJSON)
		if err != nil {
			return fmt.Errorf("insert user_interactions: %w", err)
		}

		return nil
	}); err != nil {
		return 0, fmt.Errorf("save recommendation log: %w", err)
	}

	return logID, nil
}
