package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"
	"user_service/internal/entity"
	"user_service/pkg/DataBase"
	"user_service/pkg/jwtpkg"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrNoUser             = errors.New("user not exist")
	ErrUserExist          = errors.New("user exist")
)

type AuthUseCase struct {
	repo            ISsoRepository
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

func NewAuthUseCase(repo ISsoRepository, accessTokenTTL, refreshTokenTTL time.Duration) *AuthUseCase {
	return &AuthUseCase{
		repo:            repo,
		AccessTokenTTL:  accessTokenTTL,
		RefreshTokenTTL: refreshTokenTTL,
	}
}

// Login checks is user already register and sent access-token
// if user is not exist, Login will return error
func (s *AuthUseCase) Login(ctx context.Context, email string, password string) (int, string, string, error) {
	const op = "service.Login"

	user, err := s.repo.GetUser(ctx, email)
	if err != nil {
		if errors.Is(err, DataBase.ErrUserNotFound) {
			return 0, "", "", ErrNoUser
		}

		return 0, "", "", fmt.Errorf("%s: %w", op, err)
	}

	if err = bcrypt.CompareHashAndPassword(user.Password, []byte(password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			fmt.Println(err)
			return 0, "", "", ErrInvalidCredentials
		}

		return 0, "", "", fmt.Errorf("%s: %w", op, err)
	}

	tokenAccess, err := jwtpkg.NewAccessToken(user, s.AccessTokenTTL)
	if err != nil {
		return 0, "", "", fmt.Errorf("%s: %w", op, err)
	}

	tokenRefresh, err := jwtpkg.NewAccessToken(user, s.RefreshTokenTTL)
	if err != nil {
		return 0, "", "", fmt.Errorf("%s: %w", op, err)
	}

	return user.AccessLevelId, tokenAccess, tokenRefresh, nil
}

// Register adds new user to app
// If user with given email already exists, returns error.
func (s *AuthUseCase) Register(ctx context.Context, email, password, username string) (int, error) {
	const op = "service.Register"

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := s.repo.SaveUser(ctx, email, passHash, username)
	if err != nil {
		if errors.Is(err, DataBase.ErrUserExists) {
			return 0, ErrUserExist
		}

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *AuthUseCase) DeleteAccount(ctx context.Context, id int) (bool, error) {
	const op = "service.DeleteAccount"

	err := s.repo.DeleteUser(ctx, id)
	if err != nil {
		if errors.Is(err, DataBase.ErrUserNotFound) {
			return false, fmt.Errorf("%s: %w", op, ErrNoUser)
		}

		return false, fmt.Errorf("%s: %w", op, err)
	}

	return true, nil
}

func (s *AuthUseCase) UpdateUserInfo(ctx context.Context, id int, email, password, name, surname, username string) (entity.User, error) {
	const op = "service.UpdateUserInfo"

	passhash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return entity.User{}, fmt.Errorf("%s: %w", op, err)
	}

	user, err := s.repo.UpdateUser(ctx, entity.User{
		ID:       id,
		Email:    email,
		Password: passhash,
		Name:     name,
		Surname:  surname,
		Username: username,
	})
	if err != nil {
		return entity.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (s *AuthUseCase) RefreshToken(ctx context.Context, refreshToken string) (string, string, error) {
	const op = "service.RefreshToken"

	newAccessToken, err := jwtpkg.RefreshAccessToken(refreshToken, s.RefreshTokenTTL)
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	return newAccessToken, refreshToken, nil
}
