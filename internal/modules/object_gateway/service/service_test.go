package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/k1v4/drip_mate/internal/modules/object_gateway/service"
	mockRepo "github.com/k1v4/drip_mate/mocks/internal_/modules/object_gateway/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUploadServer_UploadImage(t *testing.T) {
	imageData := []byte("fake-image-bytes")

	tests := []struct {
		name            string
		fileName        string
		setupRepo       func(r *mockRepo.IUploadRepository)
		wantErr         bool
		wantURLContains []string
	}{
		{
			name:     "success — simple filename",
			fileName: "photo.jpg",
			setupRepo: func(r *mockRepo.IUploadRepository) {
				r.On("UploadImage", mock.Anything,
					mock.MatchedBy(func(name string) bool {
						return strings.HasPrefix(name, "photo_") &&
							strings.HasSuffix(name, ".jpg")
					}),
					imageData,
				).Return("https://s3.example.com/photo_20240101_120000.jpg", nil)
			},
			wantErr:         false,
			wantURLContains: []string{"https://s3.example.com"},
		},
		{
			name:     "success — filename with dots",
			fileName: "my.photo.file.png",
			setupRepo: func(r *mockRepo.IUploadRepository) {
				r.On("UploadImage", mock.Anything,
					mock.MatchedBy(func(name string) bool {
						return strings.Contains(name, "my.photo.file_") &&
							strings.HasSuffix(name, ".png")
					}),
					imageData,
				).Return("https://s3.example.com/my.photo.file_20240101.png", nil)
			},
			wantErr: false,
		},
		{
			name:     "success — timestamp format in generated name",
			fileName: "image.jpg",
			setupRepo: func(r *mockRepo.IUploadRepository) {
				r.On("UploadImage", mock.Anything,
					mock.MatchedBy(func(name string) bool {
						parts := strings.Split(name, "_")
						if len(parts) < 2 {
							return false
						}
						return !strings.Contains(name, " ") &&
							strings.HasSuffix(name, ".jpg")
					}),
					imageData,
				).Return("https://s3.example.com/result.jpg", nil)
			},
			wantErr: false,
		},
		{
			name:     "repo error",
			fileName: "photo.jpg",
			setupRepo: func(r *mockRepo.IUploadRepository) {
				r.On("UploadImage", mock.Anything, mock.AnythingOfType("string"), imageData).
					Return("", errors.New("s3 unavailable"))
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := mockRepo.NewIUploadRepository(t)
			tc.setupRepo(repo)

			svc := service.NewUploadService(repo)
			url, err := svc.UploadImage(context.Background(), tc.fileName, imageData)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Empty(t, url)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, url)
				for _, part := range tc.wantURLContains {
					assert.Contains(t, url, part)
				}
			}
		})
	}
}

func TestUploadServer_DeleteImage(t *testing.T) {
	tests := []struct {
		name         string
		url          string
		setupRepo    func(r *mockRepo.IUploadRepository)
		wantOk       bool
		wantErr      bool
		wantFileName string
	}{
		{
			name:         "success — extracts filename from URL",
			url:          "https://s3.example.com/bucket/photo_20240101_12-00-00.jpg",
			wantFileName: "photo_20240101_12-00-00.jpg",
			setupRepo: func(r *mockRepo.IUploadRepository) {
				r.On("DeleteImage", mock.Anything, "photo_20240101_12-00-00.jpg").Return(nil)
			},
			wantOk:  true,
			wantErr: false,
		},
		{
			name:         "success — URL without path segments",
			url:          "https://s3.example.com/image.png",
			wantFileName: "image.png",
			setupRepo: func(r *mockRepo.IUploadRepository) {
				r.On("DeleteImage", mock.Anything, "image.png").Return(nil)
			},
			wantOk:  true,
			wantErr: false,
		},
		{
			name:         "success — deep path in URL",
			url:          "https://cdn.example.com/a/b/c/d/file.webp",
			wantFileName: "file.webp",
			setupRepo: func(r *mockRepo.IUploadRepository) {
				r.On("DeleteImage", mock.Anything, "file.webp").Return(nil)
			},
			wantOk:  true,
			wantErr: false,
		},
		{
			name:         "repo error",
			url:          "https://s3.example.com/photo.jpg",
			wantFileName: "photo.jpg",
			setupRepo: func(r *mockRepo.IUploadRepository) {
				r.On("DeleteImage", mock.Anything, "photo.jpg").Return(errors.New("s3 error"))
			},
			wantOk:  false,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := mockRepo.NewIUploadRepository(t)
			tc.setupRepo(repo)

			svc := service.NewUploadService(repo)
			ok, err := svc.DeleteImage(context.Background(), tc.url)

			assert.Equal(t, tc.wantOk, ok)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUploadImage_FilenameGeneration(t *testing.T) {
	tests := []struct {
		name         string
		inputName    string
		checkGenName func(t *testing.T, generated string)
	}{
		{
			name:      "no spaces in generated name",
			inputName: "photo.jpg",
			checkGenName: func(t *testing.T, generated string) {
				assert.NotContains(t, generated, " ")
			},
		},
		{
			name:      "no colons in generated name",
			inputName: "photo.jpg",
			checkGenName: func(t *testing.T, generated string) {
				assert.NotContains(t, generated, ":")
			},
		},
		{
			name:      "extension preserved",
			inputName: "photo.jpeg",
			checkGenName: func(t *testing.T, generated string) {
				assert.True(t, strings.HasSuffix(generated, ".jpeg"))
			},
		},
		{
			name:      "base name preserved with dots",
			inputName: "my.cool.photo.png",
			checkGenName: func(t *testing.T, generated string) {
				assert.True(t, strings.HasPrefix(generated, "my.cool.photo_"))
				assert.True(t, strings.HasSuffix(generated, ".png"))
			},
		},
		{
			name:      "underscore separator between name and timestamp",
			inputName: "image.jpg",
			checkGenName: func(t *testing.T, generated string) {
				withoutExt := strings.TrimSuffix(generated, ".jpg")
				parts := strings.SplitN(withoutExt, "_", 2)
				assert.Len(t, parts, 2)
				assert.Equal(t, "image", parts[0])
				assert.NotEmpty(t, parts[1])
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var capturedName string
			repo := mockRepo.NewIUploadRepository(t)
			repo.On("UploadImage", mock.Anything,
				mock.MatchedBy(func(name string) bool {
					capturedName = name
					return true
				}),
				mock.Anything,
			).Return("https://s3.example.com/result", nil)

			svc := service.NewUploadService(repo)
			_, err := svc.UploadImage(context.Background(), tc.inputName, []byte("data"))
			assert.NoError(t, err)

			tc.checkGenName(t, capturedName)
		})
	}
}
