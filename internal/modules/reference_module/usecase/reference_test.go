package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/k1v4/drip_mate/internal/entity"
	"github.com/k1v4/drip_mate/internal/modules/reference_module/usecase"
	mockRepo "github.com/k1v4/drip_mate/mocks/internal_/modules/reference_module/usecase"
)

func TestGetColors(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := mockRepo.NewIReferenceRepository(t)

		want := []entity.ColorType{
			{ID: 1, Name: "red"},
			{ID: 2, Name: "blue"},
		}

		repo.On("GetColors", ctx).Return(want, nil)

		uc := usecase.NewReferenceUseCase(repo)

		got, err := uc.GetColors(ctx)

		require.NoError(t, err)
		require.Equal(t, want, got)

		repo.AssertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		repo := mockRepo.NewIReferenceRepository(t)

		repo.On("GetColors", ctx).Return(nil, errors.New("db error"))

		uc := usecase.NewReferenceUseCase(repo)

		got, err := uc.GetColors(ctx)

		require.Error(t, err)
		require.Nil(t, got)

		repo.AssertExpectations(t)
	})
}

func TestGetMusics(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := mockRepo.NewIReferenceRepository(t)

		want := []entity.MusicType{
			{ID: 1, Name: "rock"},
			{ID: 2, Name: "jazz"},
		}

		repo.On("GetMusics", ctx).Return(want, nil)

		uc := usecase.NewReferenceUseCase(repo)

		got, err := uc.GetMusics(ctx)

		require.NoError(t, err)
		require.Equal(t, want, got)

		repo.AssertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		repo := mockRepo.NewIReferenceRepository(t)

		repo.On("GetMusics", ctx).Return(nil, errors.New("db error"))

		uc := usecase.NewReferenceUseCase(repo)

		got, err := uc.GetMusics(ctx)

		require.Error(t, err)
		require.Nil(t, got)

		repo.AssertExpectations(t)
	})
}

func TestGetCategories(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := mockRepo.NewIReferenceRepository(t)

		want := []entity.Category{
			{ID: 1, Name: "casual"},
			{ID: 2, Name: "sport"},
		}

		repo.On("GetCategories", ctx).Return(want, nil)

		uc := usecase.NewReferenceUseCase(repo)

		got, err := uc.GetCategories(ctx)

		require.NoError(t, err)
		require.Equal(t, want, got)

		repo.AssertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		repo := mockRepo.NewIReferenceRepository(t)

		repo.On("GetCategories", ctx).Return(nil, errors.New("db error"))

		uc := usecase.NewReferenceUseCase(repo)

		got, err := uc.GetCategories(ctx)

		require.Error(t, err)
		require.Nil(t, got)

		repo.AssertExpectations(t)
	})
}

func TestGetSeasons(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := mockRepo.NewIReferenceRepository(t)

		want := []entity.Season{
			{ID: 1, Name: "summer"},
			{ID: 2, Name: "winter"},
		}

		repo.On("GetSeasons", ctx).Return(want, nil)

		uc := usecase.NewReferenceUseCase(repo)

		got, err := uc.GetSeasons(ctx)

		require.NoError(t, err)
		require.Equal(t, want, got)

		repo.AssertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		repo := mockRepo.NewIReferenceRepository(t)

		repo.On("GetSeasons", ctx).Return(nil, errors.New("db error"))

		uc := usecase.NewReferenceUseCase(repo)

		got, err := uc.GetSeasons(ctx)

		require.Error(t, err)
		require.Nil(t, got)

		repo.AssertExpectations(t)
	})
}

func TestGetStyles(t *testing.T) {
	cases := []struct {
		name      string
		ctx       context.Context
		want      []entity.StyleType
		isWantErr bool
		wantErr   error
		setupMock func()
	}{
		{
			name: "success",
			ctx:  context.Background(),
			want: []entity.StyleType{
				{
					ID:   1,
					Name: "1",
				},
				{
					ID:   2,
					Name: "2",
				},
			},
			isWantErr: false,
			wantErr:   nil,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			repo := mockRepo.NewIReferenceRepository(t)

			uc := usecase.NewReferenceUseCase(repo)
			repo.On("GetStyles", tt.ctx).Return(tt.want, tt.wantErr)

			got, err := uc.GetStyles(tt.ctx)

			if tt.isWantErr {
				require.Error(t, err)
				require.Nil(t, got)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}

			repo.AssertExpectations(t)
		})
	}
}
