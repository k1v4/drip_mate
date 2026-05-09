package grpc_test

import (
	"context"
	"errors"
	"testing"

	grpcTransport "github.com/k1v4/drip_mate/internal/modules/object_gateway/transport/grpc"
	mockSvc "github.com/k1v4/drip_mate/mocks/internal_/modules/object_gateway/transport/grpc"
	uploaderv1 "github.com/k1v4/protos/gen/file_uploader"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestUploadTransport_UploadFile(t *testing.T) {
	imageData := []byte("fake-image-bytes")

	tests := []struct {
		name     string
		req      *uploaderv1.ImageUploadRequest
		setupSvc func(s *mockSvc.IUploadService)
		wantCode codes.Code
		wantURL  string
	}{
		{
			name: "success",
			req: &uploaderv1.ImageUploadRequest{
				FileName:  "photo.jpg",
				ImageData: imageData,
			},
			setupSvc: func(s *mockSvc.IUploadService) {
				s.On("UploadImage", mock.Anything, "photo.jpg", imageData).
					Return("https://s3.example.com/photo.jpg", nil)
			},
			wantCode: codes.OK,
			wantURL:  "https://s3.example.com/photo.jpg",
		},
		{
			name: "empty image data",
			req: &uploaderv1.ImageUploadRequest{
				FileName:  "photo.jpg",
				ImageData: []byte{},
			},
			wantCode: codes.InvalidArgument,
		},
		{
			name: "nil image data",
			req: &uploaderv1.ImageUploadRequest{
				FileName:  "photo.jpg",
				ImageData: nil,
			},
			wantCode: codes.InvalidArgument,
		},
		{
			name: "empty filename",
			req: &uploaderv1.ImageUploadRequest{
				FileName:  "",
				ImageData: imageData,
			},
			wantCode: codes.InvalidArgument,
		},
		{
			name: "whitespace only filename",
			req: &uploaderv1.ImageUploadRequest{
				FileName:  "   ",
				ImageData: imageData,
			},
			wantCode: codes.InvalidArgument,
		},
		{
			name: "service error",
			req: &uploaderv1.ImageUploadRequest{
				FileName:  "photo.jpg",
				ImageData: imageData,
			},
			setupSvc: func(s *mockSvc.IUploadService) {
				s.On("UploadImage", mock.Anything, "photo.jpg", imageData).
					Return("", errors.New("s3 unavailable"))
			},
			wantCode: codes.Internal,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := mockSvc.NewIUploadService(t)
			if tc.setupSvc != nil {
				tc.setupSvc(svc)
			}

			transport := grpcTransport.NewUploadTransport(svc)
			resp, err := transport.UploadFile(context.Background(), tc.req)

			if tc.wantCode == codes.OK {
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.Equal(t, tc.wantURL, resp.GetUrl())
			} else {
				require.Error(t, err)
				assert.Nil(t, resp)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tc.wantCode, st.Code())
			}
		})
	}
}

func TestUploadTransport_DeleteFile(t *testing.T) {
	tests := []struct {
		name        string
		req         *uploaderv1.ImageDeleteRequest
		setupSvc    func(s *mockSvc.IUploadService)
		wantCode    codes.Code
		wantDeleted bool
	}{
		{
			name: "success",
			req:  &uploaderv1.ImageDeleteRequest{Url: "https://s3.example.com/photo.jpg"},
			setupSvc: func(s *mockSvc.IUploadService) {
				s.On("DeleteImage", mock.Anything, "https://s3.example.com/photo.jpg").
					Return(true, nil)
			},
			wantCode:    codes.OK,
			wantDeleted: true,
		},
		{
			name:     "empty url",
			req:      &uploaderv1.ImageDeleteRequest{Url: ""},
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "whitespace only url",
			req:      &uploaderv1.ImageDeleteRequest{Url: "   "},
			wantCode: codes.InvalidArgument,
		},
		{
			name: "service error",
			req:  &uploaderv1.ImageDeleteRequest{Url: "https://s3.example.com/photo.jpg"},
			setupSvc: func(s *mockSvc.IUploadService) {
				s.On("DeleteImage", mock.Anything, "https://s3.example.com/photo.jpg").
					Return(false, errors.New("s3 error"))
			},
			wantCode: codes.Internal,
		},
		{
			name: "service returns false without error",
			req:  &uploaderv1.ImageDeleteRequest{Url: "https://s3.example.com/missing.jpg"},
			setupSvc: func(s *mockSvc.IUploadService) {
				s.On("DeleteImage", mock.Anything, "https://s3.example.com/missing.jpg").
					Return(false, nil)
			},
			wantCode:    codes.OK,
			wantDeleted: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := mockSvc.NewIUploadService(t)
			if tc.setupSvc != nil {
				tc.setupSvc(svc)
			}

			transport := grpcTransport.NewUploadTransport(svc)
			resp, err := transport.DeleteFile(context.Background(), tc.req)

			if tc.wantCode == codes.OK {
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.Equal(t, tc.wantDeleted, resp.GetIsDeleted())
			} else {
				require.Error(t, err)
				assert.Nil(t, resp)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tc.wantCode, st.Code())
			}
		})
	}
}
