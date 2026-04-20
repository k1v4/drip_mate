package entity_test

import (
	"encoding/json"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/k1v4/drip_mate/internal/modules/user_service/entity"

	"github.com/stretchr/testify/assert"
)

func TestUser_MarshalUnmarshalBinary(t *testing.T) {
	cases := []struct {
		name string
		user entity.User
	}{
		{
			name: "all fields filled",
			user: entity.User{
				ID:          uuid.MustParse(gofakeit.UUID()),
				Email:       "test@mail.com",
				Password:    "secret",
				Name:        "John",
				Surname:     "Doe",
				Username:    "johndoe",
				City:        "NYC",
				AccessLevel: "admin",
				AccessID:    0, // код не сериализуется
			},
		},
		{
			name: "empty optional fields",
			user: entity.User{
				ID:    uuid.MustParse(gofakeit.UUID()),
				Email: "empty@mail.com",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Сериализация
			data, err := tc.user.MarshalBinary()
			assert.NoError(t, err)
			assert.NotEmpty(t, data)

			// Проверим, что результат можно распарсить через json напрямую
			var tmp map[string]any
			err = json.Unmarshal(data, &tmp)
			assert.NoError(t, err)

			// Десериализация
			var newUser entity.User
			err = newUser.UnmarshalBinary(data)
			assert.NoError(t, err)

			assert.Equal(t, tc.user.ID, newUser.ID)
			assert.Equal(t, tc.user.Email, newUser.Email)
			assert.Equal(t, tc.user.Name, newUser.Name)
			assert.Equal(t, tc.user.Surname, newUser.Surname)
			assert.Equal(t, tc.user.Username, newUser.Username)
			assert.Equal(t, tc.user.City, newUser.City)
			assert.Equal(t, tc.user.AccessLevel, newUser.AccessLevel)
			assert.Equal(t, tc.user.AccessID, newUser.AccessID)
		})
	}
}
