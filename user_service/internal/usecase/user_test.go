package usecase

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
	"user_service/internal/entity"
	mocks "user_service/mocks/internal_/usecase"
	"user_service/pkg/DataBase"
	"user_service/pkg/fake"
	"user_service/pkg/jwtpkg"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthUseCase_Login(t *testing.T) {
	ctx := context.Background()

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
			mockUser:      fake.CreateUser("test@example.com", "correctPassword123", 2),
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
			mockUser:      fake.CreateUser("test@example.com", "correctPassword123", 1),
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
			mockRepo := new(mocks.ISsoRepository)
			authUC := NewAuthUseCase(mockRepo, 15*time.Minute, 7*24*time.Hour)

			mockRepo.On("GetUser", ctx, tc.email).Return(tc.mockUser, tc.mockError)

			accessLevelId, accessToken, refreshToken, err := authUC.Login(ctx, tc.email, tc.password)

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
		})
	}
}

func TestAuthUseCase_Register(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		name          string
		id            int
		email         string
		password      string
		mockError     error
		expectedError error
	}{
		{
			name:          "success",
			id:            1,
			email:         gofakeit.Email(),
			password:      gofakeit.Password(true, true, true, true, true, 12),
			mockError:     nil,
			expectedError: nil,
		},
		{
			name:          "success #2",
			id:            111,
			email:         gofakeit.Email(),
			password:      gofakeit.Password(true, true, true, true, true, 12),
			mockError:     nil,
			expectedError: nil,
		},
		{
			name:          "user exist",
			id:            0,
			email:         gofakeit.Email(),
			password:      gofakeit.Password(true, true, true, true, true, 12),
			mockError:     DataBase.ErrUserExists,
			expectedError: ErrUserExist,
		},
		{
			name:          "repo error",
			id:            0,
			email:         gofakeit.Email(),
			password:      gofakeit.Password(true, true, true, true, true, 12),
			mockError:     errors.New("something bad with repo"),
			expectedError: errors.New("service.Register: something bad with repo"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := mocks.NewISsoRepository(t)
			useCase := NewAuthUseCase(mockRepo, 15*time.Minute, 7*24*time.Hour)

			mockRepo.EXPECT().SaveUser(ctx, tc.email, mock.Anything).Return(tc.id, tc.mockError).Once()

			registerId, err := useCase.Register(ctx, tc.email, tc.password)

			if tc.expectedError != nil {
				assert.Error(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, registerId, tc.id)
		})
	}
}

func TestAuthUseCase_DeleteAccount(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		name          string
		id            int
		mockError     error
		expectedOk    bool
		expectedError error
	}{
		{
			name:          "success",
			id:            1,
			mockError:     nil,
			expectedOk:    true,
			expectedError: nil,
		},
		{
			name:          "user not found",
			id:            10,
			mockError:     DataBase.ErrUserNotFound,
			expectedOk:    false,
			expectedError: fmt.Errorf("service.DeleteAccount: %w", ErrNoUser),
		},
		{
			name:          "repo error",
			id:            5,
			mockError:     errors.New("repo failed"),
			expectedOk:    false,
			expectedError: fmt.Errorf("service.DeleteAccount: repo failed"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := mocks.NewISsoRepository(t)
			useCase := NewAuthUseCase(mockRepo, 15*time.Minute, 7*24*time.Hour)

			mockRepo.
				EXPECT().
				DeleteUser(ctx, tc.id).
				Return(tc.mockError).
				Once()

			ok, err := useCase.DeleteAccount(ctx, tc.id)

			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tc.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tc.expectedOk, ok)
		})
	}
}

func TestAuthUseCase_UpdateUserInfo(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		name          string
		inputUser     entity.User
		password      string
		updateID      int
		updateErr     error
		getUser       entity.User
		getErr        error
		expectedUser  entity.User
		expectedError error
	}{
		{
			name: "success",
			inputUser: entity.User{
				ID:       1,
				Email:    gofakeit.Email(),
				Name:     "John",
				Surname:  "Doe",
				Username: "jdoe",
				City:     "Berlin",
			},
			password: "strong-password",
			updateID: 1,
			getUser: entity.User{
				ID:       1,
				Email:    "john@mail.com",
				Name:     "John",
				Surname:  "Doe",
				Username: "jdoe",
				City:     "Berlin",
			},
			expectedUser: entity.User{
				ID:       1,
				Email:    "john@mail.com",
				Name:     "John",
				Surname:  "Doe",
				Username: "jdoe",
				City:     "Berlin",
			},
		},
		{
			name: "update repo error",
			inputUser: entity.User{
				ID:    2,
				Email: gofakeit.Email(),
			},
			password:      "pass",
			updateErr:     errors.New("update failed"),
			expectedError: fmt.Errorf("service.UpdateUserInfo: update failed"),
		},
		{
			name: "get user error",
			inputUser: entity.User{
				ID:    3,
				Email: gofakeit.Email(),
			},
			password:      "pass",
			updateID:      3,
			getErr:        errors.New("get failed"),
			expectedError: fmt.Errorf("service.UpdateUserInfo: get failed"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := mocks.NewISsoRepository(t)
			useCase := NewAuthUseCase(mockRepo, 15*time.Minute, 7*24*time.Hour)

			mockRepo.
				EXPECT().
				UpdateUser(ctx, mock.MatchedBy(func(u entity.User) bool {
					return u.ID == tc.inputUser.ID &&
						u.Email == tc.inputUser.Email &&
						bcrypt.CompareHashAndPassword(u.Password, []byte(tc.password)) == nil
				})).
				Return(tc.updateID, tc.updateErr).
				Once()

			if tc.updateErr == nil {
				mockRepo.
					EXPECT().
					GetUserById(ctx, tc.updateID).
					Return(tc.getUser, tc.getErr).
					Once()
			}

			user, err := useCase.UpdateUserInfo(
				ctx,
				tc.inputUser.ID,
				tc.inputUser.Email,
				tc.password,
				tc.inputUser.Name,
				tc.inputUser.Surname,
				tc.inputUser.Username,
				tc.inputUser.City,
			)

			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tc.expectedError.Error())
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tc.expectedUser, user)
		})
	}
}

func TestAuthUseCase_RefreshToken(t *testing.T) {
	ctx := context.Background()

	user := fake.CreateUser(
		gofakeit.Email(),
		gofakeit.Password(true, true, true, true, true, 12),
		1,
	)

	cases := []struct {
		name          string
		refreshToken  string
		expectedError bool
	}{
		{
			name: "success",
		},
		{
			name:          "invalid token",
			refreshToken:  "invalid",
			expectedError: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			useCase := NewAuthUseCase(nil, 15*time.Minute, 7*24*time.Hour)

			if tc.refreshToken == "" {
				tc.refreshToken, _ = jwtpkg.NewAccessToken(user, 1*time.Hour)
			}

			access, refresh, err := useCase.RefreshToken(ctx, tc.refreshToken)

			if tc.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "service.RefreshToken")
				return
			}

			assert.NoError(t, err)
			assert.NotEmpty(t, access)
			assert.Equal(t, tc.refreshToken, refresh)
		})
	}
}
