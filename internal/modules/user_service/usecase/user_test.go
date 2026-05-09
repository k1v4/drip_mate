package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/k1v4/drip_mate/internal/config"
	"github.com/k1v4/drip_mate/internal/entity"
	userEntity "github.com/k1v4/drip_mate/internal/modules/user_service/entity"
	"github.com/k1v4/drip_mate/internal/modules/user_service/usecase"
	mockRepo "github.com/k1v4/drip_mate/mocks/internal_/modules/user_service/usecase"
	mockAuth "github.com/k1v4/drip_mate/mocks/pkg/auth"
	mockKafka "github.com/k1v4/drip_mate/mocks/pkg/kafkaPkg"
	mockLogger "github.com/k1v4/drip_mate/mocks/pkg/logger"
	"github.com/k1v4/drip_mate/pkg/DataBase"
	"github.com/k1v4/drip_mate/pkg/kafkaPkg"
	"github.com/stretchr/testify/require"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func defaultCfg() *config.Token {
	return &config.Token{
		TTL:    time.Hour,
		Secret: "test-secret",
		Issuer: "test-issuer",
	}
}

func buildUC(
	repo *mockRepo.ISsoRepository,
	hasher *mockAuth.PasswordHasher,
	log *mockLogger.Logger,
) *usecase.AuthUseCase {
	return usecase.NewAuthUseCase(
		repo,
		log,
		(*kafkaPkg.Producer[entity.NotificationEvent])(nil),
		defaultCfg(),
		hasher,
		nil,
	)
}

func TestAuthUseCase_Login(t *testing.T) {
	validUserID := uuid.New()
	validUser := &userEntity.User{
		ID:       validUserID,
		Email:    "user@example.com",
		Password: "hashed_secret",
		AccessID: 1,
	}

	tests := []struct {
		name        string
		email       string
		password    string
		setupRepo   func(r *mockRepo.ISsoRepository)
		setupHasher func(h *mockAuth.PasswordHasher)
		wantRole    entity.Role
		wantErr     error
	}{
		{
			name:     "success",
			email:    "user@example.com",
			password: "secret",
			setupRepo: func(r *mockRepo.ISsoRepository) {
				r.On("GetUser", mock.Anything, "user@example.com").Return(validUser, nil)
			},
			setupHasher: func(h *mockAuth.PasswordHasher) {
				h.On("Verify", "secret", "hashed_secret").Return(true, nil)
			},
			wantRole: entity.Role(1),
			wantErr:  nil,
		},
		{
			name:     "user not found",
			email:    "ghost@example.com",
			password: "secret",
			setupRepo: func(r *mockRepo.ISsoRepository) {
				r.On("GetUser", mock.Anything, "ghost@example.com").Return(nil, DataBase.ErrUserNotFound)
			},
			wantErr: usecase.ErrNoUser,
		},
		{
			name:     "invalid password",
			email:    "user@example.com",
			password: "wrong",
			setupRepo: func(r *mockRepo.ISsoRepository) {
				r.On("GetUser", mock.Anything, "user@example.com").Return(validUser, nil)
			},
			setupHasher: func(h *mockAuth.PasswordHasher) {
				h.On("Verify", "wrong", "hashed_secret").Return(false, nil)
			},
			wantErr: usecase.ErrInvalidCredentials,
		},
		{
			name:     "verify returns error",
			email:    "user@example.com",
			password: "secret",
			setupRepo: func(r *mockRepo.ISsoRepository) {
				r.On("GetUser", mock.Anything, "user@example.com").Return(validUser, nil)
			},
			setupHasher: func(h *mockAuth.PasswordHasher) {
				h.On("Verify", "secret", "hashed_secret").Return(false, errors.New("hasher error"))
			},
			wantErr: errors.New("hasher error"),
		},
		{
			name:     "repo unexpected error",
			email:    "user@example.com",
			password: "secret",
			setupRepo: func(r *mockRepo.ISsoRepository) {
				r.On("GetUser", mock.Anything, "user@example.com").Return(nil, errors.New("db down"))
			},
			wantErr: errors.New("db down"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := mockRepo.NewISsoRepository(t)
			hasher := mockAuth.NewPasswordHasher(t)
			log := mockLogger.NewLogger(t)

			if tc.setupRepo != nil {
				tc.setupRepo(repo)
			}
			if tc.setupHasher != nil {
				tc.setupHasher(hasher)
			}

			uc := buildUC(repo, hasher, log)
			role, token, err := uc.Login(context.Background(), tc.email, tc.password)

			if tc.wantErr != nil {
				assert.Error(t, err)
				if errors.Is(tc.wantErr, usecase.ErrNoUser) || errors.Is(tc.wantErr, usecase.ErrInvalidCredentials) {
					assert.ErrorIs(t, err, tc.wantErr)
				}
				assert.Empty(t, token)
				assert.Zero(t, role)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantRole, role)
				assert.NotEmpty(t, token)
			}
		})
	}
}

func TestAuthUseCase_Register(t *testing.T) {
	newID := uuid.New().String()

	tests := []struct {
		name        string
		email       string
		password    string
		setupRepo   func(r *mockRepo.ISsoRepository)
		setupHasher func(h *mockAuth.PasswordHasher)
		setupWriter func(w *mockKafka.KafkaWriter)
		wantRole    entity.Role
		wantErr     error
	}{
		{
			name:     "success",
			email:    "new@example.com",
			password: "pass123",
			setupRepo: func(r *mockRepo.ISsoRepository) {
				r.On("SaveUser", mock.Anything, "new@example.com", "hashed_pass123").
					Return(newID, 1, nil)
			},
			setupHasher: func(h *mockAuth.PasswordHasher) {
				h.On("Hash", "pass123").Return("hashed_pass123", nil)
			},
			setupWriter: func(w *mockKafka.KafkaWriter) {
				w.On("WriteMessages", mock.Anything, mock.Anything).Return(nil)
			},
			wantRole: entity.Role(1),
			wantErr:  nil,
		},
		{
			name:     "success — kafka send fails, does not affect result",
			email:    "new@example.com",
			password: "pass123",
			setupRepo: func(r *mockRepo.ISsoRepository) {
				r.On("SaveUser", mock.Anything, "new@example.com", "hashed_pass123").
					Return(newID, 1, nil)
			},
			setupHasher: func(h *mockAuth.PasswordHasher) {
				h.On("Hash", "pass123").Return("hashed_pass123", nil)
			},
			setupWriter: func(w *mockKafka.KafkaWriter) {
				w.On("WriteMessages", mock.Anything, mock.Anything).Return(errors.New("kafka unavailable"))
			},
			wantRole: entity.Role(1),
			wantErr:  nil,
		},
		{
			name:     "user already exists",
			email:    "exist@example.com",
			password: "pass123",
			setupRepo: func(r *mockRepo.ISsoRepository) {
				r.On("SaveUser", mock.Anything, "exist@example.com", "hashed_pass123").
					Return("", 0, DataBase.ErrUserExists)
			},
			setupHasher: func(h *mockAuth.PasswordHasher) {
				h.On("Hash", "pass123").Return("hashed_pass123", nil)
			},
			wantErr: usecase.ErrUserExist,
		},
		{
			name:     "hash error",
			email:    "new@example.com",
			password: "pass123",
			setupHasher: func(h *mockAuth.PasswordHasher) {
				h.On("Hash", "pass123").Return("", errors.New("hash failed"))
			},
			wantErr: errors.New("hash failed"),
		},
		{
			name:     "repo unexpected error",
			email:    "new@example.com",
			password: "pass123",
			setupRepo: func(r *mockRepo.ISsoRepository) {
				r.On("SaveUser", mock.Anything, "new@example.com", "hashed_pass123").
					Return("", 0, errors.New("db error"))
			},
			setupHasher: func(h *mockAuth.PasswordHasher) {
				h.On("Hash", "pass123").Return("hashed_pass123", nil)
			},
			wantErr: errors.New("db error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := mockRepo.NewISsoRepository(t)
			hasher := mockAuth.NewPasswordHasher(t)
			log := mockLogger.NewLogger(t)
			writer := mockKafka.NewKafkaWriter(t)

			if tc.setupRepo != nil {
				tc.setupRepo(repo)
			}
			if tc.setupHasher != nil {
				tc.setupHasher(hasher)
			}
			if tc.setupWriter != nil {
				tc.setupWriter(writer)
			}

			if tc.setupWriter != nil {
				log.On("Error", mock.Anything, mock.AnythingOfType("string")).Maybe()
			}

			producer := kafkaPkg.NewProducer[entity.NotificationEvent](writer)

			uc := usecase.NewAuthUseCase(
				repo,
				log,
				producer,
				defaultCfg(),
				hasher,
				nil,
			)

			role, token, err := uc.Register(context.Background(), tc.email, tc.password)

			if tc.wantErr != nil {
				assert.Error(t, err)
				if errors.Is(tc.wantErr, usecase.ErrUserExist) {
					assert.ErrorIs(t, err, tc.wantErr)
				}
				assert.Empty(t, token)
				assert.Zero(t, role)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantRole, role)
				assert.NotEmpty(t, token)
			}
		})
	}
}

func TestAuthUseCase_DeleteAccount(t *testing.T) {
	validID := uuid.New().String()
	invalidID := "invalidUUID"

	tests := []struct {
		name        string
		id          string
		setupRepo   func(r *mockRepo.ISsoRepository)
		setupLogger func(l *mockLogger.Logger)
		needsRedis  bool
		wantOk      bool
		wantErr     error
	}{
		{
			name: "success",
			id:   validID,
			setupRepo: func(r *mockRepo.ISsoRepository) {
				r.On("DeleteUser", mock.Anything, validID).Return(nil)
			},
			setupLogger: func(l *mockLogger.Logger) {
				l.On("Error", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()
			},
			needsRedis: true,
			wantOk:     true,
			wantErr:    nil,
		},
		{
			name: "invalidID",
			id:   invalidID,
			setupRepo: func(r *mockRepo.ISsoRepository) {
				r.On("DeleteUser", mock.Anything, invalidID).Return(nil)
			},
			setupLogger: func(l *mockLogger.Logger) {
				l.On("Error", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()
			},
			needsRedis: true,
			wantOk:     true,
			wantErr:    nil,
		},
		{
			name: "user not found",
			id:   validID,
			setupRepo: func(r *mockRepo.ISsoRepository) {
				r.On("DeleteUser", mock.Anything, validID).Return(DataBase.ErrUserNotFound)
			},
			needsRedis: false,
			wantOk:     false,
			wantErr:    usecase.ErrNoUser,
		},
		{
			name: "repo unexpected error",
			id:   validID,
			setupRepo: func(r *mockRepo.ISsoRepository) {
				r.On("DeleteUser", mock.Anything, validID).Return(errors.New("db error"))
			},
			needsRedis: false,
			wantOk:     false,
			wantErr:    errors.New("db error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := mockRepo.NewISsoRepository(t)
			hasher := mockAuth.NewPasswordHasher(t)
			log := mockLogger.NewLogger(t)

			tc.setupRepo(repo)
			if tc.setupLogger != nil {
				tc.setupLogger(log)
			}

			var redisClient *redis.Client
			if tc.needsRedis {
				mr, err := miniredis.Run()
				require.NoError(t, err)
				t.Cleanup(mr.Close)

				redisClient = redis.NewClient(&redis.Options{Addr: mr.Addr()})
				t.Cleanup(func() { _ = redisClient.Close() })
			}

			uc := usecase.NewAuthUseCase(
				repo,
				log,
				(*kafkaPkg.Producer[entity.NotificationEvent])(nil),
				defaultCfg(),
				hasher,
				redisClient,
			)

			ok, err := uc.DeleteAccount(context.Background(), tc.id)

			if tc.needsRedis {
				time.Sleep(10 * time.Millisecond)
			}

			assert.Equal(t, tc.wantOk, ok)
			if tc.wantErr != nil {
				assert.Error(t, err)
				if errors.Is(tc.wantErr, usecase.ErrNoUser) {
					assert.ErrorIs(t, err, usecase.ErrNoUser)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthUseCase_UpdateUserInfo(t *testing.T) {
	userID := uuid.New()
	updatedUser := &userEntity.User{
		ID:       userID,
		Name:     "Ivan",
		Surname:  "Petrov",
		Username: "ivan_p",
	}

	tests := []struct {
		name       string
		id         string
		firstName  string
		surname    string
		username   string
		gender     string
		setupRepo  func(r *mockRepo.ISsoRepository)
		needsRedis bool
		wantUser   *userEntity.User
		wantErr    bool
	}{
		{
			name:      "success",
			id:        userID.String(),
			firstName: "Ivan",
			surname:   "Petrov",
			username:  "ivan_p",
			gender:    "male",
			setupRepo: func(r *mockRepo.ISsoRepository) {
				r.On("UpdateUserPersonal", mock.Anything, &userEntity.UpdatePersonal{
					ID:       userID.String(),
					Name:     "Ivan",
					Surname:  "Petrov",
					Username: "ivan_p",
					Gender:   "male",
				}).Return(userID.String(), nil)
				r.On("GetUserById", mock.Anything, userID.String()).Return(updatedUser, nil)
			},
			needsRedis: true,
			wantUser:   updatedUser,
			wantErr:    false,
		},
		{
			name:      "UpdateUserPersonal repo error",
			id:        userID.String(),
			firstName: "Ivan",
			surname:   "Petrov",
			username:  "ivan_p",
			gender:    "male",
			setupRepo: func(r *mockRepo.ISsoRepository) {
				r.On("UpdateUserPersonal", mock.Anything, mock.Anything).
					Return("", errors.New("db error"))
			},
			needsRedis: false,
			wantErr:    true,
		},
		{
			name:      "GetUserById error after update",
			id:        userID.String(),
			firstName: "Ivan",
			surname:   "Petrov",
			username:  "ivan_p",
			gender:    "male",
			setupRepo: func(r *mockRepo.ISsoRepository) {
				r.On("UpdateUserPersonal", mock.Anything, mock.Anything).
					Return(userID.String(), nil)
				r.On("GetUserById", mock.Anything, userID.String()).
					Return(nil, errors.New("db error"))
			},
			needsRedis: true,
			wantErr:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := mockRepo.NewISsoRepository(t)
			hasher := mockAuth.NewPasswordHasher(t)
			log := mockLogger.NewLogger(t)

			tc.setupRepo(repo)

			var redisClient *redis.Client
			if tc.needsRedis {
				mr, err := miniredis.Run()
				require.NoError(t, err)
				t.Cleanup(mr.Close)

				redisClient = redis.NewClient(&redis.Options{Addr: mr.Addr()})
				t.Cleanup(func() { _ = redisClient.Close() })

				log.On("Error", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()
			}

			uc := usecase.NewAuthUseCase(
				repo,
				log,
				(*kafkaPkg.Producer[entity.NotificationEvent])(nil),
				defaultCfg(),
				hasher,
				redisClient,
			)

			user, err := uc.UpdateUserInfo(context.Background(), tc.id, tc.firstName, tc.surname, tc.username, tc.gender)

			if tc.needsRedis {
				time.Sleep(10 * time.Millisecond)
			}

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantUser, user)
			}
		})
	}
}

func TestAuthUseCase_UpdatePassword(t *testing.T) {
	userID := uuid.New()
	storedUser := &userEntity.User{
		ID:       userID,
		Password: "hashed_oldpass",
	}

	tests := []struct {
		name        string
		pass        *userEntity.UpdatePass
		setupRepo   func(r *mockRepo.ISsoRepository)
		setupHasher func(h *mockAuth.PasswordHasher)
		needsRedis  bool
		wantErr     error
	}{
		{
			name: "success",
			pass: &userEntity.UpdatePass{CurrPassword: "oldpass", NewPassword: "newpass"},
			setupRepo: func(r *mockRepo.ISsoRepository) {
				r.On("GetUserById", mock.Anything, userID.String()).Return(storedUser, nil)
				r.On("UpdatePassword", mock.Anything, userID, "hashed_newpass").Return(nil)
			},
			setupHasher: func(h *mockAuth.PasswordHasher) {
				h.On("Verify", "oldpass", "hashed_oldpass").Return(true, nil)
				h.On("Hash", "newpass").Return("hashed_newpass", nil)
			},
			needsRedis: true,
			wantErr:    nil,
		},
		{
			name: "wrong current password",
			pass: &userEntity.UpdatePass{CurrPassword: "wrongpass", NewPassword: "newpass"},
			setupRepo: func(r *mockRepo.ISsoRepository) {
				r.On("GetUserById", mock.Anything, userID.String()).Return(storedUser, nil)
			},
			setupHasher: func(h *mockAuth.PasswordHasher) {
				h.On("Verify", "wrongpass", "hashed_oldpass").Return(false, nil)
			},
			needsRedis: false,
			wantErr:    usecase.ErrInvalidCredentials,
		},
		{
			name: "verify returns error",
			pass: &userEntity.UpdatePass{CurrPassword: "oldpass", NewPassword: "newpass"},
			setupRepo: func(r *mockRepo.ISsoRepository) {
				r.On("GetUserById", mock.Anything, userID.String()).Return(storedUser, nil)
			},
			setupHasher: func(h *mockAuth.PasswordHasher) {
				h.On("Verify", "oldpass", "hashed_oldpass").Return(false, errors.New("hasher error"))
			},
			needsRedis: false,
			wantErr:    errors.New("hasher error"),
		},
		{
			name: "get user error",
			pass: &userEntity.UpdatePass{CurrPassword: "oldpass", NewPassword: "newpass"},
			setupRepo: func(r *mockRepo.ISsoRepository) {
				r.On("GetUserById", mock.Anything, userID.String()).Return(nil, errors.New("db error"))
			},
			needsRedis: false,
			wantErr:    errors.New("db error"),
		},
		{
			name: "update password repo error",
			pass: &userEntity.UpdatePass{CurrPassword: "oldpass", NewPassword: "newpass"},
			setupRepo: func(r *mockRepo.ISsoRepository) {
				r.On("GetUserById", mock.Anything, userID.String()).Return(storedUser, nil)
				r.On("UpdatePassword", mock.Anything, userID, "hashed_newpass").Return(errors.New("db error"))
			},
			setupHasher: func(h *mockAuth.PasswordHasher) {
				h.On("Verify", "oldpass", "hashed_oldpass").Return(true, nil)
				h.On("Hash", "newpass").Return("hashed_newpass", nil)
			},
			needsRedis: false,
			wantErr:    errors.New("db error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := mockRepo.NewISsoRepository(t)
			hasher := mockAuth.NewPasswordHasher(t)
			log := mockLogger.NewLogger(t)

			if tc.setupRepo != nil {
				tc.setupRepo(repo)
			}
			if tc.setupHasher != nil {
				tc.setupHasher(hasher)
			}

			var redisClient *redis.Client
			if tc.needsRedis {
				mr, err := miniredis.Run()
				require.NoError(t, err)
				t.Cleanup(mr.Close)

				redisClient = redis.NewClient(&redis.Options{Addr: mr.Addr()})
				t.Cleanup(func() { _ = redisClient.Close() })

				log.On("Error", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()
			}

			uc := usecase.NewAuthUseCase(
				repo,
				log,
				(*kafkaPkg.Producer[entity.NotificationEvent])(nil),
				defaultCfg(),
				hasher,
				redisClient,
			)

			err := uc.UpdatePassword(context.Background(), userID, tc.pass)

			if tc.needsRedis {
				time.Sleep(10 * time.Millisecond)
			}

			if tc.wantErr != nil {
				assert.Error(t, err)
				if errors.Is(tc.wantErr, usecase.ErrInvalidCredentials) {
					assert.ErrorIs(t, err, usecase.ErrInvalidCredentials)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthUseCase_UpdateContext(t *testing.T) {
	userID := uuid.New()
	styles := []int{1, 2}
	colors := []int{3, 4}
	music := []int{5}

	tests := []struct {
		name      string
		req       *userEntity.UpdateContext
		setupRepo func(r *mockRepo.ISsoRepository)
		wantErr   bool
	}{
		{
			name: "success with all fields",
			req: &userEntity.UpdateContext{
				City:   "Moscow",
				Styles: &styles,
				Colors: &colors,
				Music:  &music,
			},
			setupRepo: func(r *mockRepo.ISsoRepository) {
				r.On("UpdateUserContext", mock.Anything, mock.Anything).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "success with empty optional fields",
			req: &userEntity.UpdateContext{
				City: "Perm",
			},
			setupRepo: func(r *mockRepo.ISsoRepository) {
				r.On("UpdateUserContext", mock.Anything, mock.Anything).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "repo error",
			req: &userEntity.UpdateContext{
				City: "Moscow",
			},
			setupRepo: func(r *mockRepo.ISsoRepository) {
				r.On("UpdateUserContext", mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := mockRepo.NewISsoRepository(t)
			hasher := mockAuth.NewPasswordHasher(t)
			log := mockLogger.NewLogger(t)

			tc.setupRepo(repo)

			uc := buildUC(repo, hasher, log)
			err := uc.UpdateContext(context.Background(), userID, tc.req)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// проверяем, что usecase проставил ID в req
				assert.Equal(t, userID, tc.req.ID)
			}
		})
	}
}

func TestAuthUseCase_SaveOutfit(t *testing.T) {
	userID := uuid.New()
	outfitID := uuid.New()
	catalogIDs := []uuid.UUID{uuid.New(), uuid.New()}

	tests := []struct {
		name      string
		req       userEntity.SaveOutfitRequest
		setupRepo func(r *mockRepo.ISsoRepository)
		wantID    uuid.UUID
		wantErr   bool
	}{
		{
			name: "success without log update",
			req: userEntity.SaveOutfitRequest{
				Name:           "Summer outfit",
				CatalogItemIDs: catalogIDs,
				LogID:          0,
			},
			setupRepo: func(r *mockRepo.ISsoRepository) {
				r.On("SaveOutfit", mock.Anything, userID, mock.Anything).Return(outfitID, nil)
			},
			wantID:  outfitID,
			wantErr: false,
		},
		{
			name: "success with log update",
			req: userEntity.SaveOutfitRequest{
				Name:           "Winter outfit",
				CatalogItemIDs: catalogIDs,
				LogID:          42,
			},
			setupRepo: func(r *mockRepo.ISsoRepository) {
				r.On("SaveOutfit", mock.Anything, userID, mock.Anything).Return(outfitID, nil)
				r.On("UpdateUserOutfitLog", mock.Anything, 42).Return(nil).Maybe()
			},
			wantID:  outfitID,
			wantErr: false,
		},
		{
			name: "repo error",
			req: userEntity.SaveOutfitRequest{
				Name:           "Outfit",
				CatalogItemIDs: catalogIDs,
			},
			setupRepo: func(r *mockRepo.ISsoRepository) {
				r.On("SaveOutfit", mock.Anything, userID, mock.Anything).Return(uuid.Nil, errors.New("db error"))
			},
			wantID:  uuid.Nil,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := mockRepo.NewISsoRepository(t)
			hasher := mockAuth.NewPasswordHasher(t)
			log := mockLogger.NewLogger(t)

			tc.setupRepo(repo)

			uc := buildUC(repo, hasher, log)
			id, err := uc.SaveOutfit(context.Background(), userID, tc.req)

			assert.Equal(t, tc.wantID, id)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthUseCase_GetOutfits(t *testing.T) {
	userID := uuid.New()
	outfits := []userEntity.Outfit{
		{
			ID:    uuid.New(),
			Name:  "Summer outfit",
			Items: []userEntity.OutfitItem{{ID: uuid.New(), Name: "T-shirt"}},
		},
		{
			ID:    uuid.New(),
			Name:  "Winter outfit",
			Items: []userEntity.OutfitItem{{ID: uuid.New(), Name: "Jacket"}},
		},
	}

	tests := []struct {
		name      string
		setupRepo func(r *mockRepo.ISsoRepository)
		wantLen   int
		wantErr   bool
	}{
		{
			name: "success",
			setupRepo: func(r *mockRepo.ISsoRepository) {
				r.On("GetUserOutfits", mock.Anything, userID).Return(outfits, nil)
			},
			wantLen: 2,
			wantErr: false,
		},
		{
			name: "empty list",
			setupRepo: func(r *mockRepo.ISsoRepository) {
				r.On("GetUserOutfits", mock.Anything, userID).Return([]userEntity.Outfit{}, nil)
			},
			wantLen: 0,
			wantErr: false,
		},
		{
			name: "repo error",
			setupRepo: func(r *mockRepo.ISsoRepository) {
				r.On("GetUserOutfits", mock.Anything, userID).Return(nil, errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := mockRepo.NewISsoRepository(t)
			hasher := mockAuth.NewPasswordHasher(t)
			log := mockLogger.NewLogger(t)

			tc.setupRepo(repo)

			uc := buildUC(repo, hasher, log)
			result, err := uc.GetOutfits(context.Background(), userID)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, tc.wantLen)
			}
		})
	}
}

func TestAuthUseCase_DeleteOutfit(t *testing.T) {
	userID := uuid.New()
	outfitID := uuid.New()

	tests := []struct {
		name      string
		setupRepo func(r *mockRepo.ISsoRepository)
		wantErr   bool
	}{
		{
			name: "success",
			setupRepo: func(r *mockRepo.ISsoRepository) {
				r.On("DeleteOutfit", mock.Anything, userID, outfitID).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "repo error",
			setupRepo: func(r *mockRepo.ISsoRepository) {
				r.On("DeleteOutfit", mock.Anything, userID, outfitID).Return(errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := mockRepo.NewISsoRepository(t)
			hasher := mockAuth.NewPasswordHasher(t)
			log := mockLogger.NewLogger(t)

			tc.setupRepo(repo)

			uc := buildUC(repo, hasher, log)
			err := uc.DeleteOutfit(context.Background(), userID, outfitID)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
