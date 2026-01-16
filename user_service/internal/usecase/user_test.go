package usecase

import (
	"context"
	"errors"
	"testing"
	"time"
	"user_service/internal/entity"
	mocks "user_service/mocks/internal_/usecase"
	"user_service/pkg/DataBase"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthUseCase_Login(t *testing.T) {
	ctx := context.Background()

	// Вспомогательная функция для создания тестового пользователя
	createTestUser := func(email, password string, accessLevel int) entity.User {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		return entity.User{
			ID:            1,
			Email:         email,
			Password:      hashedPassword,
			AccessLevelId: accessLevel,
		}
	}

	testCases := []struct {
		name          string
		email         string
		password      string
		mockUser      entity.User
		mockError     error
		expectedError error
		expectTokens  bool
		expectedLevel int
	}{
		{
			name:          "success",
			email:         "test@example.com",
			password:      "correctPassword123",
			mockUser:      createTestUser("test@example.com", "correctPassword123", 2),
			expectedError: nil,
			expectTokens:  true,
			expectedLevel: 2,
		},
		{
			name:          "user not found",
			email:         "notfound@example.com",
			password:      "anyPassword",
			mockUser:      entity.User{},
			mockError:     DataBase.ErrUserNotFound,
			expectedError: ErrNoUser,
			expectTokens:  false,
			expectedLevel: 0,
		},
		{
			name:          "wrong pass",
			email:         "test@example.com",
			password:      "wrongPassword",
			mockUser:      createTestUser("test@example.com", "correctPassword123", 1),
			expectedError: ErrInvalidCredentials,
			expectTokens:  false,
			expectedLevel: 0,
		},
		{
			name:          "repository error",
			email:         "test@example.com",
			password:      "password123",
			mockUser:      entity.User{},
			mockError:     assert.AnError,
			expectedError: assert.AnError,
			expectTokens:  false,
			expectedLevel: 0,
		},
		{
			name:          "empty email",
			email:         "",
			password:      "password123",
			mockUser:      entity.User{},
			mockError:     DataBase.ErrUserNotFound,
			expectedError: ErrNoUser,
			expectTokens:  false,
			expectedLevel: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mockRepo := new(mocks.ISsoRepository)
			authUC := NewAuthUseCase(mockRepo, 15*time.Minute, 7*24*time.Hour)

			mockRepo.On("GetUser", ctx, tc.email).Return(tc.mockUser, tc.mockError)

			// Act
			accessLevelId, accessToken, refreshToken, err := authUC.Login(ctx, tc.email, tc.password)

			// Assert
			if tc.expectedError != nil {
				require.Error(t, err)
				if errors.Is(tc.expectedError, ErrNoUser) || errors.Is(tc.expectedError, ErrInvalidCredentials) {
					assert.ErrorIs(t, err, tc.expectedError)
				}
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tc.expectedLevel, accessLevelId)

			if tc.expectTokens {
				assert.NotEmpty(t, accessToken)
				assert.NotEmpty(t, refreshToken)
				assert.NotEqual(t, accessToken, refreshToken)
			} else {
				assert.Empty(t, accessToken)
				assert.Empty(t, refreshToken)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
