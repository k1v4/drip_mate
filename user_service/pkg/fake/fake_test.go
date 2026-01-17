package fake

import (
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
)

func Test_CreateUser(t *testing.T) {
	email := gofakeit.Email()
	password := gofakeit.Password(true, true, true, true, true, 12)
	accessLevel := gofakeit.IntRange(1, 2)

	user := CreateUser(email, password, accessLevel)

	assert.NotNil(t, user)

	assert.Equal(t, user.Email, email)
	assert.NotEmpty(t, user.Password)
	assert.Equal(t, user.AccessLevelId, accessLevel)
}
