package fake

import (
	"github.com/k1v4/drip_mate/internal/modules/user_service/entity"
	"github.com/k1v4/drip_mate/pkg/auth"
)

func CreateUser(email, password string, accessLevel int, hasher auth.PasswordHasher) entity.User {
	passHash, _ := hasher.Hash(password)
	return entity.User{
		ID:            "1",
		Email:         email,
		Password:      passHash,
		AccessLevelId: accessLevel,
	}
}
