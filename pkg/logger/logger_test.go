package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func newTestLogger(buf *bytes.Buffer) Logger {
	encoderCfg := zap.NewProductionEncoderConfig()
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		zapcore.AddSync(buf),
		zapcore.InfoLevel,
	)

	zl := zap.New(core)

	return &logger{
		serviceName: ServiceName,
		logger:      zl,
	}
}

func TestLogger_Info(t *testing.T) {
	var buf bytes.Buffer

	l := newTestLogger(&buf)

	//nolint:staticcheck
	ctx := context.WithValue(context.Background(), RequestID, "req-123")

	l.Info(ctx, "info message", zap.String("key", "value"))

	var result map[string]any

	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal log output: %v", err)
	}

	if result["msg"] != "info message" {
		t.Errorf("expected msg 'info message', got %v", result["msg"])
	}

	if result["key"] != "value" {
		t.Errorf("expected key 'value', got %v", result["key"])
	}

	if result[RequestID] != "req-123" {
		t.Errorf("expected requestID 'req-123', got %v", result[RequestID])
	}

	if result[ServiceName] != ServiceName {
		t.Errorf("expected service '%s', got %v", ServiceName, result[ServiceName])
	}

	if _, ok := result["time"]; !ok {
		t.Error("expected time field to exist")
	}
}

func TestLogger_Error(t *testing.T) {
	var buf bytes.Buffer

	l := newTestLogger(&buf)

	//nolint:staticcheck
	ctx := context.WithValue(context.Background(), RequestID, "err-456")

	l.Error(ctx, "error message", zap.String("errorKey", "errorValue"))

	var result map[string]any

	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal log output: %v", err)
	}

	if result["msg"] != "error message" {
		t.Errorf("expected msg 'error message', got %v", result["msg"])
	}

	if result["errorKey"] != "errorValue" {
		t.Errorf("expected errorKey 'errorValue', got %v", result["errorKey"])
	}

	if result[RequestID] != "err-456" {
		t.Errorf("expected requestID 'err-456', got %v", result[RequestID])
	}

	if result[ServiceName] != ServiceName {
		t.Errorf("expected service '%s', got %v", ServiceName, result[ServiceName])
	}

	if _, ok := result["time"]; !ok {
		t.Error("expected time field to exist")
	}
}

func TestGetLoggerFromContext(t *testing.T) {
	var buf bytes.Buffer

	l := newTestLogger(&buf)

	ctx := context.WithValue(context.Background(), LoggerKey, l)

	got := GetLoggerFromContext(ctx)

	if got == nil {
		t.Fatal("expected logger, got nil")
	}
}

func TestInfo_WithoutRequestID(t *testing.T) {
	var buf bytes.Buffer

	l := newTestLogger(&buf)

	ctx := context.Background()

	l.Info(ctx, "message without request id")

	var result map[string]any

	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal log output: %v", err)
	}

	if _, ok := result[RequestID]; ok {
		t.Error("did not expect requestID field")
	}
}

func TestNewLogger(t *testing.T) {
	l := NewLogger()

	if l == nil {
		t.Fatal("expected logger instance, got nil")
	}
}
