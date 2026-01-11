package usecase

import (
	"context"
	"user_service/internal/entity"
)

type ISsoRepository interface {
	SaveUser(ctx context.Context, email string, password []byte, username string) (int, error)
	GetUser(ctx context.Context, email string) (entity.User, error)
	GetUserById(ctx context.Context, id int) (entity.User, error)
	DeleteUser(ctx context.Context, id int) error
	UpdateUser(ctx context.Context, newUser entity.User) (entity.User, error)
}

type ISsoService interface {
	Login(ctx context.Context, email string, password string) (int, string, string, error)
	Register(ctx context.Context, email, password, username string) (int, error)
	DeleteAccount(ctx context.Context, id int) (bool, error)
	UpdateUserInfo(ctx context.Context, id int, email, password, name, surname, username string) (entity.User, error)
	RefreshToken(ctx context.Context, token string) (string, string, error)
}
