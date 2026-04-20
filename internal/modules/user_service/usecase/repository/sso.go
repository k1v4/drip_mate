package repository

import (
	"context"
	"database/sql"
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
func (a *AuthRepository) GetUser(ctx context.Context, email string) (entity.User, error) {
	const op = "repository.GetUser"

	var (
		tmpName     sql.NullString
		tmpSurname  sql.NullString
		tmpUsername sql.NullString
		tmpCity     sql.NullString
	)

	s, args, err := a.Builder.Select(
		"u.id",
		"u.email",
		"u.password",
		"u.username",
		"u.name",
		"u.surname",
		"u.city",
		"u.access_id",
		"al.name AS access_level",
	).
		From("users u").
		LeftJoin("access_level al ON u.access_id = al.id").
		Where(sq.Eq{"email": email}).
		ToSql()
	if err != nil {
		return entity.User{}, fmt.Errorf("%s: %w", op, err)
	}

	var result entity.User
	err = a.Pool.QueryRow(ctx, s, args...).Scan(
		&result.ID,
		&result.Email,
		&result.Password,
		&tmpUsername,
		&tmpName,
		&tmpSurname,
		&tmpCity,
		&result.AccessLevelId,
		&result.AccessLevelName,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.User{}, DataBase.ErrUserNotFound
		}

		return entity.User{}, fmt.Errorf("%s: %w", op, err)
	}

	result.Name = ""
	if tmpName.Valid {
		result.Name = tmpName.String
	}

	result.Surname = ""
	if tmpSurname.Valid {
		result.Surname = tmpSurname.String
	}

	result.Username = ""
	if tmpUsername.Valid {
		result.Username = tmpUsername.String
	}

	result.City = ""
	if tmpCity.Valid {
		result.City = tmpCity.String
	}

	return result, nil
}

// GetUserById takes user from Database by Id
func (a *AuthRepository) GetUserById(ctx context.Context, id string) (entity.User, error) {
	const op = "repository.GetUser"

	var (
		tmpName     sql.NullString
		tmpSurname  sql.NullString
		tmpUsername sql.NullString
		tmpCity     sql.NullString
	)

	s, args, err := a.Builder.Select(
		"u.id",
		"u.email",
		"u.password",
		"u.username",
		"u.name",
		"u.surname",
		"u.city",
		"u.access_id",
		"al.name AS access_level",
	).
		From("users u").
		LeftJoin("access_level al ON u.access_id = al.id").
		Where(sq.Eq{"u.id": id}).
		ToSql()
	if err != nil {
		return entity.User{}, fmt.Errorf("%s: %w", op, err)
	}

	var result entity.User
	err = a.Pool.QueryRow(ctx, s, args...).Scan(
		&result.ID,
		&result.Email,
		&result.Password,
		&tmpUsername,
		&tmpName,
		&tmpSurname,
		&tmpCity,
		&result.AccessLevelId,
		&result.AccessLevelName,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.User{}, DataBase.ErrUserNotFound
		}

		return entity.User{}, fmt.Errorf("%s: %w", op, err)
	}

	result.Name = ""
	if tmpName.Valid {
		result.Name = tmpName.String
	}

	result.Surname = ""
	if tmpSurname.Valid {
		result.Surname = tmpSurname.String
	}

	result.Username = ""
	if tmpUsername.Valid {
		result.Username = tmpUsername.String
	}

	result.City = ""
	if tmpCity.Valid {
		result.City = tmpCity.String
	}

	return result, nil
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

func (a *AuthRepository) UpdateUser(ctx context.Context, newUser *entity.User) (string, error) {
	const op = "repository.UpdateUser"

	err := postgres.WithTx(ctx, a.Pool, func(tx pgx.Tx) error {
		builder := a.Builder.Update("users")

		if newUser.Email != "" {
			builder = builder.Set("email", newUser.Email)
		}
		if len(newUser.Password) != 0 {
			builder = builder.Set("password", newUser.Password)
		}
		if newUser.Username != "" {
			builder = builder.Set("username", newUser.Username)
		}
		if newUser.Name != "" {
			builder = builder.Set("name", newUser.Name)
		}
		if newUser.Surname != "" {
			builder = builder.Set("surname", newUser.Surname)
		}
		if newUser.City != "" {
			builder = builder.Set("city", newUser.City)
		}

		sqlReq, args, err := builder.
			Where(sq.Eq{"id": newUser.ID}).
			ToSql()
		if err != nil {
			return fmt.Errorf("%s: build sql: %w", op, err)
		}

		if _, err = tx.Exec(ctx, sqlReq, args...); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
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
