package usecase

import (
	"context"

	"github.com/google/uuid"
	totalEntity "github.com/k1v4/drip_mate/internal/entity"
	"github.com/k1v4/drip_mate/internal/modules/user_service/entity"
)

type ISsoRepository interface {
	SaveUser(ctx context.Context, email string, password string) (string, int, error)
	GetUser(ctx context.Context, email string) (entity.User, error)
	GetUserById(ctx context.Context, id string) (entity.User, error)
	DeleteUser(ctx context.Context, id string) error
	UpdateUser(ctx context.Context, newUser *entity.User) (string, error)
	SaveOutfit(ctx context.Context, userID uuid.UUID, saveItems entity.SaveOutfitRequest) (uuid.UUID, error)
	GetUserOutfits(ctx context.Context, userID uuid.UUID) ([]entity.Outfit, error)
	DeleteOutfit(ctx context.Context, userID, outfitID uuid.UUID) error
}

type ISsoService interface {
	Login(ctx context.Context, email string, password string) (totalEntity.Role, string, error)
	Register(ctx context.Context, email, password string) (string, string, error)
	DeleteAccount(ctx context.Context, id string) (bool, error)
	UpdateUserInfo(ctx context.Context, id string, email, password, name, surname, username, city string) (entity.User, error)
	SaveOutfit(ctx context.Context, userID uuid.UUID, saveItems entity.SaveOutfitRequest) (uuid.UUID, error)
	GetOutfits(ctx context.Context, userID uuid.UUID) ([]entity.Outfit, error)
	DeleteOutfit(ctx context.Context, userID, outfitID uuid.UUID) error
}
