package grpc

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/k1v4/drip_mate/file_upload_service/pkg/logger"
	uploaderv1 "github.com/k1v4/protos/gen/file_uploader"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

type Server struct {
	grpcServer *grpc.Server
	listener   net.Listener
}

func NewServer(ctx context.Context, grpcPort int, uploadService IUploadService) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			ContextWithLogger(logger.GetLoggerFromContext(ctx)),
		),
	}

	grpcServer := grpc.NewServer(opts...)
	uploaderv1.RegisterFileUploaderServer(grpcServer, NewUploadTransport(uploadService))

	return &Server{grpcServer, listener}, nil
}

func (s *Server) Start(ctx context.Context) error {
	eg := errgroup.Group{}

	eg.Go(func() error {
		lg := logger.GetLoggerFromContext(ctx)
		if lg != nil {
			lg.Info(ctx, "starting grpc server", zap.Int("port", s.listener.Addr().(*net.TCPAddr).Port))
		}

		return s.grpcServer.Serve(s.listener)
	})

	return eg.Wait()
}

func (s *Server) Stop(ctx context.Context) error {
	s.grpcServer.GracefulStop()

	l := logger.GetLoggerFromContext(ctx)
	if l != nil {
		l.Info(ctx, "grpc server stopped")
	}

	return nil
}
