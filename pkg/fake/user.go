package fake

import (
	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	totalEntity "github.com/k1v4/drip_mate/internal/entity"
	"github.com/k1v4/drip_mate/internal/modules/user_service/entity"
	"github.com/k1v4/drip_mate/pkg/auth"
)

func CreateUser(email, password string, accessLevel totalEntity.Role, hasher auth.PasswordHasher) entity.User {
	passHash, _ := hasher.Hash(password)
	return entity.User{
		ID:       uuid.MustParse(gofakeit.UUID()),
		Email:    email,
		Password: passHash,
		AccessID: int(accessLevel),
	}
}
