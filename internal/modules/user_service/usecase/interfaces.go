package usecase

import (
	"context"

	"github.com/google/uuid"
	totalEntity "github.com/k1v4/drip_mate/internal/entity"
	"github.com/k1v4/drip_mate/internal/modules/user_service/entity"
)

type ISsoRepository interface {
	SaveUser(ctx context.Context, email string, password string) (string, int, error)
	GetUser(ctx context.Context, email string) (*entity.User, error)
	GetUserById(ctx context.Context, id string) (*entity.User, error)
	DeleteUser(ctx context.Context, id string) error
	UpdateUserPersonal(ctx context.Context, newUser *entity.UpdatePersonal) (string, error)
	SaveOutfit(ctx context.Context, userID uuid.UUID, saveItems entity.SaveOutfitRequest) (uuid.UUID, error)
	GetUserOutfits(ctx context.Context, userID uuid.UUID) ([]entity.Outfit, error)
	DeleteOutfit(ctx context.Context, userID, outfitID uuid.UUID) error
	UpdatePassword(ctx context.Context, userID uuid.UUID, newPasswordHash string) error
	UpdateUserContext(ctx context.Context, req *entity.UpdateContext) error
}

type ISsoService interface {
	Login(ctx context.Context, email string, password string) (totalEntity.Role, string, error)
	Register(ctx context.Context, email, password string) (string, string, error)
	DeleteAccount(ctx context.Context, id string) (bool, error)
	UpdateUserInfo(
		ctx context.Context,
		id string,
		name, surname, username string,
	) (*entity.User, error)
	SaveOutfit(ctx context.Context, userID uuid.UUID, saveItems entity.SaveOutfitRequest) (uuid.UUID, error)
	GetOutfits(ctx context.Context, userID uuid.UUID) ([]entity.Outfit, error)
	DeleteOutfit(ctx context.Context, userID, outfitID uuid.UUID) error
	UpdatePassword(ctx context.Context, id uuid.UUID, pass *entity.UpdatePass) error
	UpdateContext(ctx context.Context, userID uuid.UUID, req *entity.UpdateContext) error
	GetUserByID(ctx context.Context, userID uuid.UUID) (*entity.User, error)
}
