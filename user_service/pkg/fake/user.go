package fake

import (
	"user_service/internal/entity"

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
