package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"user_service/internal/entity"
	"user_service/pkg/DataBase"
	"user_service/pkg/DataBase/postgres"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
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
	password []byte,
) (int, error) {
	const op = "repository.SaveUser"

	var id int
	err := withTx(ctx, a.Pool, func(tx pgx.Tx) error {
		sqlReq, args, err := a.Builder.
			Insert("users").
			Columns("email", "password").
			Values(email, password).
			Suffix("RETURNING id").
			ToSql()
		if err != nil {
			return fmt.Errorf("%s: build sql: %w", op, err)
		}

		if err = tx.QueryRow(ctx, sqlReq, args...).Scan(&id); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return DataBase.ErrUserExists
			}

			return fmt.Errorf("%s: exec: %w", op, err)
		}

		return nil
	})

	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
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
func (a *AuthRepository) GetUserById(ctx context.Context, id int) (entity.User, error) {
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

func (a *AuthRepository) DeleteUser(ctx context.Context, id int) error {
	const op = "repository.DeleteUser"

	if err := withTx(ctx, a.Pool, func(tx pgx.Tx) error {
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

func (a *AuthRepository) UpdateUser(ctx context.Context, newUser entity.User) (entity.User, error) {
	const op = "repository.UpdateUser"

	// TODO обновить логику всех полей: если не приходит поле, то его и не обновляем
	err := withTx(ctx, a.Pool, func(tx pgx.Tx) error {
		sqlReq, args, err := a.Builder.Update("users").
			Set("email", newUser.Email).
			Set("password", newUser.Password).
			Set("username", newUser.Username).
			Set("name", newUser.Name).
			Set("surname", newUser.Surname).
			Set("city", newUser.City).
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
		return entity.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return newUser, nil
}

func withTx(
	ctx context.Context,
	pool *pgxpool.Pool,
	fn func(tx pgx.Tx) error,
) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		_ = tx.Rollback(ctx) // безопасно, если Commit уже был
	}()

	if err = fn(tx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
