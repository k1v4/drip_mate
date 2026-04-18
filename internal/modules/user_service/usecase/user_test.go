package usecase

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/k1v4/drip_mate/internal/modules/user_service/entity"
	mocks "github.com/k1v4/drip_mate/mocks/internal_/modules/user_service/usecase"
	mocks2 "github.com/k1v4/drip_mate/mocks/pkg/auth"
	"github.com/k1v4/drip_mate/pkg/DataBase"
	"github.com/k1v4/drip_mate/pkg/fake"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAuthUseCase_Login(t *testing.T) {
	ctx := context.Background()
	hasher := mocks2.NewPasswordHasher(t)

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
			mockUser:      fake.CreateUser("test@example.com", "correctPassword123", 2, hasher),
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
			mockUser:      fake.CreateUser("test@example.com", "correctPassword123", 1, hasher),
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
			authUC := NewAuthUseCase(mockRepo, nil, nil, nil, hasher)

			mockRepo.On("GetUser", ctx, tc.email).Return(tc.mockUser, tc.mockError)

			accessLevelId, accessToken, err := authUC.Login(ctx, tc.email, tc.password)

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
			} else {
				assert.Empty(t, accessToken)
			}
		})
	}
}

func TestAuthUseCase_Register(t *testing.T) {
	ctx := context.Background()
	hasher := mocks2.NewPasswordHasher(t)

	cases := []struct {
		name          string
		id            string
		email         string
		password      string
		mockError     error
		expectedError error
	}{
		{
			name:          "success",
			id:            gofakeit.UUID(),
			email:         gofakeit.Email(),
			password:      gofakeit.Password(true, true, true, true, true, 12),
			mockError:     nil,
			expectedError: nil,
		},
		{
			name:          "success #2",
			id:            gofakeit.UUID(),
			email:         gofakeit.Email(),
			password:      gofakeit.Password(true, true, true, true, true, 12),
			mockError:     nil,
			expectedError: nil,
		},
		{
			name:          "user exist",
			id:            gofakeit.UUID(),
			email:         gofakeit.Email(),
			password:      gofakeit.Password(true, true, true, true, true, 12),
			mockError:     DataBase.ErrUserExists,
			expectedError: ErrUserExist,
		},
		{
			name:          "repo error",
			id:            gofakeit.UUID(),
			email:         gofakeit.Email(),
			password:      gofakeit.Password(true, true, true, true, true, 12),
			mockError:     errors.New("something bad with repo"),
			expectedError: errors.New("service.Register: something bad with repo"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := mocks.NewISsoRepository(t)
			useCase := NewAuthUseCase(mockRepo, nil, nil, nil, hasher)

			mockRepo.EXPECT().SaveUser(ctx, tc.email, mock.Anything).Return(tc.id, 0, tc.mockError).Once()

			registerId, _, err := useCase.Register(ctx, tc.email, tc.password)

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
	hasher := mocks2.NewPasswordHasher(t)

	cases := []struct {
		name          string
		id            string
		mockError     error
		expectedOk    bool
		expectedError error
	}{
		{
			name:          "success",
			id:            gofakeit.UUID(),
			mockError:     nil,
			expectedOk:    true,
			expectedError: nil,
		},
		{
			name:          "user not found",
			id:            gofakeit.UUID(),
			mockError:     DataBase.ErrUserNotFound,
			expectedOk:    false,
			expectedError: fmt.Errorf("service.DeleteAccount: %w", ErrNoUser),
		},
		{
			name:          "repo error",
			id:            gofakeit.UUID(),
			mockError:     errors.New("repo failed"),
			expectedOk:    false,
			expectedError: fmt.Errorf("service.DeleteAccount: repo failed"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := mocks.NewISsoRepository(t)
			useCase := NewAuthUseCase(mockRepo, nil, nil, nil, hasher)

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
	hasher := mocks2.NewPasswordHasher(t)

	cases := []struct {
		name          string
		inputUser     entity.User
		password      string
		updateID      string
		updateErr     error
		getUser       entity.User
		getErr        error
		expectedUser  entity.User
		expectedError error
	}{
		{
			name: "success",
			inputUser: entity.User{
				ID:       gofakeit.UUID(),
				Email:    gofakeit.Email(),
				Name:     "John",
				Surname:  "Doe",
				Username: "jdoe",
				City:     "Berlin",
			},
			password: "strong-password",
			updateID: gofakeit.UUID(),
			getUser: entity.User{
				ID:       gofakeit.UUID(),
				Email:    "john@mail.com",
				Name:     "John",
				Surname:  "Doe",
				Username: "jdoe",
				City:     "Berlin",
			},
			expectedUser: entity.User{
				ID:       gofakeit.UUID(),
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
				ID:    gofakeit.UUID(),
				Email: gofakeit.Email(),
			},
			password:      "pass",
			updateErr:     errors.New("update failed"),
			expectedError: fmt.Errorf("service.UpdateUserInfo: update failed"),
		},
		{
			name: "get user error",
			inputUser: entity.User{
				ID:    gofakeit.UUID(),
				Email: gofakeit.Email(),
			},
			password:      "pass",
			updateID:      gofakeit.UUID(),
			getErr:        errors.New("get failed"),
			expectedError: fmt.Errorf("service.UpdateUserInfo: get failed"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := mocks.NewISsoRepository(t)
			useCase := NewAuthUseCase(mockRepo, nil, nil, nil, hasher)

			mockRepo.
				EXPECT().
				UpdateUser(ctx, mock.MatchedBy(func(u entity.User) bool {
					isP, _ := hasher.Verify(u.Password, tc.password)

					return u.ID == tc.inputUser.ID &&
						u.Email == tc.inputUser.Email && isP
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
