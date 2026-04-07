package fake

import (
	"github.com/k1v4/drip_mate/internal/modules/user_service/entity"

	"golang.org/x/crypto/bcrypt"
)

func CreateUser(email, password string, accessLevel int) entity.User {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	return entity.User{
		ID:            1,
		Email:         email,
		Password:      hashedPassword,
		AccessLevelId: accessLevel,
	}
}
