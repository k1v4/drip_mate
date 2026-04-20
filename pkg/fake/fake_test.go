package fake

import (
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/k1v4/drip_mate/internal/entity"
	mocks "github.com/k1v4/drip_mate/mocks/pkg/auth"
	"github.com/stretchr/testify/assert"
)

func Test_CreateUser(t *testing.T) {
	email := gofakeit.Email()
	password := gofakeit.Password(true, true, true, true, true, 12)
	accessLevel := entity.Role(gofakeit.IntRange(1, 2))
	hasher := mocks.NewPasswordHasher(t)

	user := CreateUser(email, password, accessLevel, hasher)

	assert.NotNil(t, user)

	assert.Equal(t, user.Email, email)
	assert.NotEmpty(t, user.Password)
	assert.Equal(t, user.AccessID, accessLevel)
}
