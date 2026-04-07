package logger

import (
	"context"

	"go.uber.org/zap"
)

type loggerKey string

const (
	LoggerKey   = loggerKey("logger")
	RequestID   = "requestID"
	ServiceName = "service"
)

type logger struct {
	serviceName string
	logger      *zap.Logger
}

type Logger interface {
	Info(ctx context.Context, msg string, fields ...zap.Field)
	Error(ctx context.Context, msg string, fields ...zap.Field)
}

func (l *logger) Info(ctx context.Context, msg string, fields ...zap.Field) {
	fields = append(fields, zap.String(ServiceName, l.serviceName))

	if ctx.Value(RequestID) != nil {
		fields = append(fields, zap.String(RequestID, ctx.Value(RequestID).(string)))
	}

	l.logger.Info(msg, fields...)
}

func (l *logger) Error(ctx context.Context, msg string, fields ...zap.Field) {
	fields = append(fields, zap.String(ServiceName, l.serviceName))

	if ctx.Value(RequestID) != nil {
		fields = append(fields, zap.String(RequestID, ctx.Value(RequestID).(string)))
	}

	l.logger.Error(msg, fields...)
}

func GetLoggerFromContext(ctx context.Context) Logger {
	return ctx.Value(LoggerKey).(Logger)
}

func NewLogger() Logger {
	l, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer l.Sync()

	return &logger{
		serviceName: ServiceName,
		logger:      l,
	}
}
