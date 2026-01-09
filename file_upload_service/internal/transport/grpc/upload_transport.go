package grpc

import (
	"context"
	"strings"

	uploaderv1 "github.com/k1v4/protos/gen/file_uploader"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type IUploadService interface {
	UploadImage(ctx context.Context, fileName string, imageData []byte) (string, error)
	DeleteImage(ctx context.Context, url string) (bool, error)
}

type UploadTransport struct {
	uploaderv1.UnimplementedFileUploaderServer
	service IUploadService
}

func NewUploadTransport(service IUploadService) *UploadTransport {
	return &UploadTransport{service: service}
}

func (u *UploadTransport) UploadFile(ctx context.Context, req *uploaderv1.ImageUploadRequest) (*uploaderv1.ImageUploadResponse, error) {
	const op = "transport.UploadImage"

	imageData := req.GetImageData()
	if len(imageData) == 0 {
		return nil, status.Error(codes.InvalidArgument, "image data is empty")
	}

	fileName := req.GetFileName()
	if len(strings.TrimSpace(fileName)) == 0 {
		return nil, status.Error(codes.InvalidArgument, "fileName is empty")
	}

	url, err := u.service.UploadImage(ctx, fileName, imageData)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &uploaderv1.ImageUploadResponse{
		Url: url,
	}, nil
}

func (u *UploadTransport) DeleteFile(ctx context.Context, req *uploaderv1.ImageDeleteRequest) (*uploaderv1.ImageDeleteResponse, error) {
	const op = "uploader.DeleteImage"

	url := req.GetUrl()
	if len(strings.TrimSpace(url)) == 0 {
		return nil, status.Error(codes.InvalidArgument, "url is empty")
	}

	res, err := u.service.DeleteImage(ctx, url)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &uploaderv1.ImageDeleteResponse{
		IsDeleted: res,
	}, nil
}
