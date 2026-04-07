package httpserver

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestServerOptions(t *testing.T) {
	handler := http.NewServeMux()
	handler.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name             string
		options          []Option
		expectedAddr     string
		expectedRead     time.Duration
		expectedWrite    time.Duration
		expectedShutdown time.Duration
	}{
		{
			name:             "defaults",
			options:          nil,
			expectedAddr:     ":8080",
			expectedRead:     5 * time.Second,
			expectedWrite:    5 * time.Second,
			expectedShutdown: 3 * time.Second,
		},
		{
			name: "custom port",
			options: []Option{
				Port("9090"),
			},
			expectedAddr:     ":9090",
			expectedRead:     5 * time.Second,
			expectedWrite:    5 * time.Second,
			expectedShutdown: 3 * time.Second,
		},
		{
			name: "custom read/write timeout",
			options: []Option{
				ReadTimeout(10 * time.Second),
				WriteTimeout(15 * time.Second),
			},
			expectedAddr:     ":8080",
			expectedRead:     10 * time.Second,
			expectedWrite:    15 * time.Second,
			expectedShutdown: 3 * time.Second,
		},
		{
			name: "custom shutdown timeout",
			options: []Option{
				ShutdownTimeout(7 * time.Second),
			},
			expectedAddr:     ":8080",
			expectedRead:     5 * time.Second,
			expectedWrite:    5 * time.Second,
			expectedShutdown: 7 * time.Second,
		},
		{
			name: "all custom",
			options: []Option{
				Port("9091"),
				ReadTimeout(8 * time.Second),
				WriteTimeout(12 * time.Second),
				ShutdownTimeout(20 * time.Second),
			},
			expectedAddr:     ":9091",
			expectedRead:     8 * time.Second,
			expectedWrite:    12 * time.Second,
			expectedShutdown: 20 * time.Second,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := New(handler, tc.options...)

			assert.Equal(t, tc.expectedAddr, s.server.Addr)
			assert.Equal(t, tc.expectedRead, s.server.ReadTimeout)
			assert.Equal(t, tc.expectedWrite, s.server.WriteTimeout)
			assert.Equal(t, tc.expectedShutdown, s.shutdownTimeout)
		})
	}
}
