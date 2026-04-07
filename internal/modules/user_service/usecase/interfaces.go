package usecase

import (
	"context"

	"github.com/k1v4/drip_mate/internal/modules/user_service/entity"
)

type ISsoRepository interface {
	SaveUser(ctx context.Context, email string, password []byte) (int, error)
	GetUser(ctx context.Context, email string) (entity.User, error)
	GetUserById(ctx context.Context, id int) (entity.User, error)
	DeleteUser(ctx context.Context, id int) error
	UpdateUser(ctx context.Context, newUser entity.User) (int, error)
}

type ISsoService interface {
	Login(ctx context.Context, email string, password string) (int, string, string, error)
	Register(ctx context.Context, email, password string) (int, error)
	DeleteAccount(ctx context.Context, id int) (bool, error)
	UpdateUserInfo(ctx context.Context, id int, email, password, name, surname, username, city string) (entity.User, error)
	RefreshToken(ctx context.Context, token string) (string, string, error)
}
