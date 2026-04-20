package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/k1v4/drip_mate/internal/modules/user_service/entity"
	"github.com/k1v4/drip_mate/pkg/DataBase"
	"github.com/k1v4/drip_mate/pkg/DataBase/postgres"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
)

type AuthRepository struct {
	*postgres.Postgres
}

func NewAuthRepository(pg *postgres.Postgres) *AuthRepository {
	return &AuthRepository{
		Postgres: pg,
	}
}

// SaveUser adds new user to Database
func (a *AuthRepository) SaveUser(
	ctx context.Context,
	email string,
	password string,
) (string, int, error) {
	const op = "repository.SaveUser"

	var id string
	var accessID int
	err := postgres.WithTx(ctx, a.Pool, func(tx pgx.Tx) error {
		sqlReq, args, err := a.Builder.
			Insert("users").
			Columns("email", "password").
			Values(email, password).
			Suffix("RETURNING id, access_id").
			ToSql()
		if err != nil {
			return fmt.Errorf("%s: build sql: %w", op, err)
		}

		if err = tx.QueryRow(ctx, sqlReq, args...).Scan(&id, &accessID); err != nil {
			if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == "23505" {
				return DataBase.ErrUserExists
			}

			return fmt.Errorf("%s: exec: %w", op, err)
		}

		return nil
	})

	if err != nil {
		return "", 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, accessID, nil
}

// GetUser takes user from Database by Email
func (a *AuthRepository) GetUser(ctx context.Context, email string) (*entity.User, error) {
	const op = "repository.GetUserById"

	query := `
	SELECT
		u.id,
		u.email,
		u.password,
		u.username,
		u.name,
		u.surname,
		u.city,
		u.access_id,
		al.name AS access_level,

		COALESCE((
			SELECT json_agg(m.name)
			FROM music_user mu
			JOIN music m ON m.id = mu.music_id
			WHERE mu.user_id = u.id
		), '[]') AS music,

		COALESCE((
			SELECT json_agg(st.name)
			FROM style_user su
			JOIN style_types st ON st.id = su.style_id
			WHERE su.user_id = u.id
		), '[]') AS styles,

		COALESCE((
			SELECT json_agg(ct.name)
			FROM color_user cu
			JOIN color_types ct ON ct.id = cu.color_id
			WHERE cu.user_id = u.id
		), '[]') AS colors,

		COALESCE((
			SELECT json_agg(
				json_build_object(
					'id', son.id,
					'name', son.name,
					'items', COALESCE(items.items, '[]')
				)
			)
			FROM saved_outfits_name son
			LEFT JOIN (
				SELECT
					so.outfit_id,
					json_agg(json_build_object(
						'id', c.id,
						'name', c.name,
						'image', c.image_url,
						'material', c.material
					)) AS items
				FROM saved_outfits so
				LEFT JOIN catalog c ON c.id = so.catalog_item_id
				GROUP BY so.outfit_id
			) items ON items.outfit_id = son.id
			WHERE son.user_id = u.id
		), '[]') AS outfits

	FROM users u
	LEFT JOIN access_level al ON u.access_id = al.id
	WHERE u.email = $1;
	`

	var (
		result entity.User

		musicJSON   []byte
		stylesJSON  []byte
		colorsJSON  []byte
		outfitsJSON []byte
	)

	err := a.Pool.QueryRow(ctx, query, email).Scan(
		&result.ID,
		&result.Email,
		&result.Password,
		&result.Username,
		&result.Name,
		&result.Surname,
		&result.City,
		&result.AccessID,
		&result.AccessLevel,

		&musicJSON,
		&stylesJSON,
		&colorsJSON,
		&outfitsJSON,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, DataBase.ErrUserNotFound
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := json.Unmarshal(musicJSON, &result.Music); err != nil {
		return nil, fmt.Errorf("%s: music unmarshal: %w", op, err)
	}

	if err := json.Unmarshal(stylesJSON, &result.Styles); err != nil {
		return nil, fmt.Errorf("%s: styles unmarshal: %w", op, err)
	}

	if err := json.Unmarshal(colorsJSON, &result.Colors); err != nil {
		return nil, fmt.Errorf("%s: colors unmarshal: %w", op, err)
	}

	if err := json.Unmarshal(outfitsJSON, &result.Outfits); err != nil {
		return nil, fmt.Errorf("%s: outfits unmarshal: %w", op, err)
	}

	return new(result), nil
}

// GetUserById takes user from Database by Id
func (a *AuthRepository) GetUserById(ctx context.Context, id string) (*entity.User, error) {
	const op = "repository.GetUserById"

	query := `
	SELECT
		u.id,
		u.email,
		u.password,
		u.username,
		u.name,
		u.surname,
		u.city,
		u.access_id,
		al.name AS access_level,

		COALESCE((
			SELECT json_agg(m.name)
			FROM music_user mu
			JOIN music m ON m.id = mu.music_id
			WHERE mu.user_id = u.id
		), '[]') AS music,

		COALESCE((
			SELECT json_agg(st.name)
			FROM style_user su
			JOIN style_types st ON st.id = su.style_id
			WHERE su.user_id = u.id
		), '[]') AS styles,

		COALESCE((
			SELECT json_agg(ct.name)
			FROM color_user cu
			JOIN color_types ct ON ct.id = cu.color_id
			WHERE cu.user_id = u.id
		), '[]') AS colors,

		COALESCE((
			SELECT json_agg(
				json_build_object(
					'id', son.id,
					'name', son.name,
					'items', COALESCE(items.items, '[]')
				)
			)
			FROM saved_outfits_name son
			LEFT JOIN (
				SELECT
					so.outfit_id,
					json_agg(json_build_object(
						'id', c.id,
						'name', c.name,
						'image', c.image_url,
						'material', c.material
					)) AS items
				FROM saved_outfits so
				LEFT JOIN catalog c ON c.id = so.catalog_item_id
				GROUP BY so.outfit_id
			) items ON items.outfit_id = son.id
			WHERE son.user_id = u.id
		), '[]') AS outfits

	FROM users u
	LEFT JOIN access_level al ON u.access_id = al.id
	WHERE u.id = $1;
	`

	var (
		result entity.User

		musicJSON   []byte
		stylesJSON  []byte
		colorsJSON  []byte
		outfitsJSON []byte
	)

	err := a.Pool.QueryRow(ctx, query, id).Scan(
		&result.ID,
		&result.Email,
		&result.Password,
		&result.Username,
		&result.Name,
		&result.Surname,
		&result.City,
		&result.AccessID,
		&result.AccessLevel,

		&musicJSON,
		&stylesJSON,
		&colorsJSON,
		&outfitsJSON,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, DataBase.ErrUserNotFound
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := json.Unmarshal(musicJSON, &result.Music); err != nil {
		return nil, fmt.Errorf("%s: music unmarshal: %w", op, err)
	}

	if err := json.Unmarshal(stylesJSON, &result.Styles); err != nil {
		return nil, fmt.Errorf("%s: styles unmarshal: %w", op, err)
	}

	if err := json.Unmarshal(colorsJSON, &result.Colors); err != nil {
		return nil, fmt.Errorf("%s: colors unmarshal: %w", op, err)
	}

	if err := json.Unmarshal(outfitsJSON, &result.Outfits); err != nil {
		return nil, fmt.Errorf("%s: outfits unmarshal: %w", op, err)
	}

	return new(result), nil
}

func (a *AuthRepository) DeleteUser(ctx context.Context, id string) error {
	const op = "repository.DeleteUser"

	if err := postgres.WithTx(ctx, a.Pool, func(tx pgx.Tx) error {
		sqlReq, args, err := a.Builder.
			Delete("users").
			Where(sq.Eq{"id": id}).
			ToSql()
		if err != nil {
			return fmt.Errorf("%s: build sql: %w", op, err)
		}

		if _, err := tx.Exec(ctx, sqlReq, args...); err != nil {
			return fmt.Errorf("%s: exec: %w", op, err)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *AuthRepository) UpdateUserPersonal(ctx context.Context, newUser *entity.UpdatePersonal) (string, error) {
	const op = "repository.UpdateUser"

	err := postgres.WithTx(ctx, a.Pool, func(tx pgx.Tx) error {
		builder := a.Builder.Update("users")

		if newUser.Username != "" {
			builder = builder.Set("username", newUser.Username)
		}
		if newUser.Name != "" {
			builder = builder.Set("name", newUser.Name)
		}
		if newUser.Surname != "" {
			builder = builder.Set("surname", newUser.Surname)
		}

		sqlReq, args, err := builder.
			Where(sq.Eq{"id": newUser.ID}).
			ToSql()
		if err != nil {
			return fmt.Errorf("%s: build sql: %w", op, err)
		}

		if _, err = tx.Exec(ctx, sqlReq, args...); err != nil {
			if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == "23505" {
				return DataBase.ErrUserExists
			}

			return fmt.Errorf("%s: exec: %w", op, err)
		}

		return nil
	})
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return newUser.ID, nil
}

func (a *AuthRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, newPasswordHash string) error {
	const op = "repository.UpdatePassword"

	err := postgres.WithTx(ctx, a.Pool, func(tx pgx.Tx) error {
		builder := a.Builder.Update("users")

		if newPasswordHash != "" {
			builder = builder.Set("password", newPasswordHash)
		}

		sqlReq, args, err := builder.
			Where(sq.Eq{"id": userID}).
			ToSql()
		if err != nil {
			return fmt.Errorf("%s: build sql: %w", op, err)
		}

		if _, err = tx.Exec(ctx, sqlReq, args...); err != nil {
			return fmt.Errorf("%s: exec: %w", op, err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *AuthRepository) SaveOutfit(ctx context.Context, userID uuid.UUID, saveItems entity.SaveOutfitRequest) (uuid.UUID, error) {
	const op = "repository.SaveOutfit"

	var outfitID uuid.UUID

	if err := postgres.WithTx(ctx, a.Pool, func(tx pgx.Tx) error {
		// создаём запись с именем аутфита
		err := tx.QueryRow(ctx, `
            INSERT INTO saved_outfits_name (name, user_id)
            VALUES ($1, $2)
            RETURNING id
        `, saveItems.Name, userID).Scan(&outfitID)
		if err != nil {
			return fmt.Errorf("%s: insert outfit name: %w", op, err)
		}

		// вставляем предметы батчем
		insert := a.Builder.
			Insert("saved_outfits").
			Columns("outfit_id", "catalog_item_id", "created_at")
		for _, itemID := range saveItems.CatalogItemIDs {
			insert = insert.Values(outfitID, itemID, sq.Expr("NOW()"))
		}

		sqlReq, args, err := insert.ToSql()
		if err != nil {
			return fmt.Errorf("%s: build sql: %w", op, err)
		}

		if _, err = tx.Exec(ctx, sqlReq, args...); err != nil {
			return fmt.Errorf("%s: insert outfit items: %w", op, err)
		}

		return nil
	}); err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	return outfitID, nil
}

func (a *AuthRepository) GetUserOutfits(ctx context.Context, userID uuid.UUID) ([]entity.Outfit, error) {
	const op = "repository.GetUserOutfits"

	// Запрос собирает все аутфиты пользователя, группируя вещи в JSON-массив
	query := `
	SELECT 
		son.id, 
		son.name,
		COALESCE(
			json_agg(json_build_object(
				'id', c.id,
				'name', c.name,
				'image', c.image_url,
				'material', c.material
			)) FILTER (WHERE c.id IS NOT NULL), 
			'[]'
		) AS items
	FROM saved_outfits_name son
	LEFT JOIN saved_outfits so ON son.id = so.outfit_id
	LEFT JOIN catalog c ON so.catalog_item_id = c.id
	WHERE son.user_id = $1
	GROUP BY son.id, son.name;
	`

	rows, err := a.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: query: %w", op, err)
	}
	defer rows.Close()

	outfits, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (entity.Outfit, error) {
		var o entity.Outfit
		err = row.Scan(&o.ID, &o.Name, &o.Items)
		return o, err
	})

	if err != nil {
		return nil, fmt.Errorf("%s: collect rows: %w", op, err)
	}

	return outfits, nil
}

func (a *AuthRepository) DeleteOutfit(ctx context.Context, userID, outfitID uuid.UUID) error {
	const op = "repository.DeleteOutfit"

	if err := postgres.WithTx(ctx, a.Pool, func(tx pgx.Tx) error {
		// удаляем из название, а там каскадом и второй таблицы удалится
		result, err := tx.Exec(ctx, `
			DELETE FROM saved_outfits_name
			WHERE id = $1 AND user_id = $2
		`, outfitID, userID)
		if err != nil {
			return fmt.Errorf("%s: delete outfit: %w", op, err)
		}

		if result.RowsAffected() == 0 {
			return DataBase.ErrOutfitNotFound
		}

		return nil
	}); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *AuthRepository) UpdateUserContext(ctx context.Context, req *entity.UpdateContext) error {
	const op = "repository.UpdateUserContext"

	err := postgres.WithTx(ctx, a.Pool, func(tx pgx.Tx) error {
		if req.City != nil {
			builder := a.Builder.
				Update("users").
				Set("city", *req.City).
				Where(sq.Eq{"id": req.ID})

			sqlReq, args, err := builder.ToSql()
			if err != nil {
				return fmt.Errorf("%s: build city update: %w", op, err)
			}

			if _, err = tx.Exec(ctx, sqlReq, args...); err != nil {
				return fmt.Errorf("%s: exec city update: %w", op, err)
			}
		}

		// styles
		if req.Styles != nil && len(*req.Styles) > 0 {
			// delete old
			if _, err := tx.Exec(ctx,
				`DELETE FROM style_user WHERE user_id = $1`,
				req.ID,
			); err != nil {
				return fmt.Errorf("%s: delete styles: %w", op, err)
			}

			// insert new
			builder := a.Builder.
				Insert("style_user").
				Columns("user_id", "style_id")

			for _, styleID := range *req.Styles {
				builder = builder.Values(req.ID, styleID)
			}

			sqlReq, args, err := builder.ToSql()
			if err != nil {
				return fmt.Errorf("%s: build styles insert: %w", op, err)
			}

			if _, err = tx.Exec(ctx, sqlReq, args...); err != nil {
				return fmt.Errorf("%s: insert styles: %w", op, err)
			}
		}

		// colors
		if req.Colors != nil && len(*req.Colors) > 0 {
			if _, err := tx.Exec(ctx,
				`DELETE FROM color_user WHERE user_id = $1`,
				req.ID,
			); err != nil {
				return fmt.Errorf("%s: delete colors: %w", op, err)
			}

			builder := a.Builder.
				Insert("color_user").
				Columns("user_id", "color_id")

			for _, colorID := range *req.Colors {
				builder = builder.Values(req.ID, colorID)
			}

			sqlReq, args, err := builder.ToSql()
			if err != nil {
				return fmt.Errorf("%s: build colors insert: %w", op, err)
			}

			if _, err = tx.Exec(ctx, sqlReq, args...); err != nil {
				return fmt.Errorf("%s: insert colors: %w", op, err)
			}
		}

		// music
		if req.Music != nil && len(*req.Music) > 0 {
			if _, err := tx.Exec(ctx,
				`DELETE FROM music_user WHERE user_id = $1`,
				req.ID,
			); err != nil {
				return fmt.Errorf("%s: delete music: %w", op, err)
			}

			builder := a.Builder.
				Insert("music_user").
				Columns("user_id", "music_id")

			for _, musicID := range *req.Music {
				builder = builder.Values(req.ID, musicID)
			}

			sqlReq, args, err := builder.ToSql()
			if err != nil {
				return fmt.Errorf("%s: build music insert: %w", op, err)
			}

			if _, err = tx.Exec(ctx, sqlReq, args...); err != nil {
				return fmt.Errorf("%s: insert music: %w", op, err)
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
