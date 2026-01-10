package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/k1v4/drip_mate/user_service/internal/entity"
	"github.com/k1v4/drip_mate/user_service/pkg/DataBase"
	"github.com/k1v4/drip_mate/user_service/pkg/DataBase/postgres"
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
func (a *AuthRepository) SaveUser(ctx context.Context, email string, password []byte, username string) (int, error) {
	const op = "repository.SaveUser"

	s, args, err := a.Builder.Insert("users").
		Columns("email", "password", "username").
		Values(email, password, username).
		Suffix("RETURNING id").
		ToSql()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	var id int
	err = a.Pool.QueryRow(ctx, s, args...).Scan(&id)
	if err != nil {
		var pqErr *pgconn.PgError
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return 0, DataBase.ErrUserExists
		}

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

// GetUser takes user from Database by Email
func (a *AuthRepository) GetUser(ctx context.Context, email string) (entity.User, error) {
	const op = "repository.GetUser"

	var (
		tmpName    sql.NullString
		tmpSurname sql.NullString
	)

	s, args, err := a.Builder.Select("*").
		From("users").
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
		&result.Username,
		&tmpName,
		&tmpSurname,
		&result.AccessLevelId,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.User{}, DataBase.ErrUserNotFound
		}

		return entity.User{}, fmt.Errorf("%s: %w", op, err)
	}

	if tmpName.Valid {
		result.Name = tmpName.String
	} else {
		result.Name = ""
	}

	if tmpSurname.Valid {
		result.Surname = tmpSurname.String
	} else {
		result.Surname = ""
	}

	return result, nil
}

// GetUserById takes user from Database by Id
func (a *AuthRepository) GetUserById(ctx context.Context, id int) (entity.User, error) {
	const op = "repository.GetUser"

	var result entity.User
	s, args, err := a.Builder.Select("*").
		From("users").
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return entity.User{}, fmt.Errorf("%s: %w", op, err)
	}

	err = a.Pool.QueryRow(ctx, s, args...).Scan(&result.ID, &result.Email, &result.Password, &result.Username,
		&result.Name, &result.Surname, &result.AccessLevelId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.User{}, fmt.Errorf("%s: %w", op, DataBase.ErrUserNotFound)
		}

		return entity.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return result, nil
}

func (a *AuthRepository) DeleteUser(ctx context.Context, id int) error {
	const op = "repository.DeleteUser"

	s, args, err := a.Builder.Delete("users").Where(sq.Eq{"id": id}).ToSql()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = a.Pool.Exec(ctx, s, args...)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *AuthRepository) UpdateUser(ctx context.Context, newUser entity.User) (entity.User, error) {
	const op = "repository.UpdateUser"

	s, args, err := a.Builder.Update("users").
		Set("email", newUser.Email).
		Set("password", newUser.Password).
		Set("username", newUser.Username).
		Set("name", newUser.Name).
		Set("surname", newUser.Surname).
		Where(sq.Eq{"id": newUser.ID}).
		ToSql()
	if err != nil {
		return entity.User{}, fmt.Errorf("%s: %w", op, err)
	}

	_, err = a.Pool.Exec(ctx, s, args...)
	if err != nil {
		return entity.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return newUser, nil
}
