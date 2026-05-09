package usecase_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	notificationSvc "github.com/k1v4/drip_mate/internal/modules/notification_service"
	notificationEntity "github.com/k1v4/drip_mate/internal/modules/notification_service/entity"
	"github.com/k1v4/drip_mate/internal/modules/notification_service/usecase"
	mockEmail "github.com/k1v4/drip_mate/mocks/internal_/modules/notification_service/usecase"
	mockLogger "github.com/k1v4/drip_mate/mocks/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newTemplates(t *testing.T) *notificationSvc.Templates {
	t.Helper()
	tmpl, err := notificationSvc.NewTemplates()
	require.NoError(t, err)
	return tmpl
}

func TestEmailNotificationUseCase_SendEmailNotification(t *testing.T) {
	tests := []struct {
		name            string
		email           string
		setupClient     func(c *mockEmail.EmailClient)
		setupLogger     func(l *mockLogger.Logger)
		wantErr         bool
		wantErrContains string
	}{
		{
			name:  "success — email sent with rendered html",
			email: "user@example.com",
			setupClient: func(c *mockEmail.EmailClient) {
				c.On("Send", mock.MatchedBy(func(e *notificationEntity.Email) bool {
					return len(e.To) == 1 &&
						e.To[0] == "user@example.com" &&
						e.Subject == "Добро пожаловать в drip mate" &&
						e.HTML != "" &&
						e.Text == ""
				})).Return(&notificationEntity.SendResult{StatusCode: http.StatusOK}, nil)
			},
			wantErr: false,
		},
		{
			name:  "html contains appURL substitution",
			email: "user@example.com",
			setupClient: func(c *mockEmail.EmailClient) {
				c.On("Send", mock.MatchedBy(func(e *notificationEntity.Email) bool {
					return e.HTML != "" &&
						assert.Contains(t, e.HTML, "google.com") &&
						assert.Contains(t, e.HTML, "google.com/unsubscribe") &&
						assert.Contains(t, e.HTML, "google.com/privacy")
				})).Return(&notificationEntity.SendResult{StatusCode: http.StatusOK}, nil)
			},
			wantErr: false,
		},
		{
			name:  "send error — returns wrapped error",
			email: "user@example.com",
			setupClient: func(c *mockEmail.EmailClient) {
				c.On("Send", mock.Anything).Return(nil, errors.New("smtp unavailable"))
			},
			wantErr:         true,
			wantErrContains: "send email",
		},
		{
			name:  "unexpected status 500 — returns error",
			email: "user@example.com",
			setupClient: func(c *mockEmail.EmailClient) {
				c.On("Send", mock.Anything).Return(
					&notificationEntity.SendResult{StatusCode: http.StatusInternalServerError}, nil,
				)
			},
			wantErr:         true,
			wantErrContains: "unexpected status code 500",
		},
		{
			name:  "unexpected status 202 — returns error",
			email: "user@example.com",
			setupClient: func(c *mockEmail.EmailClient) {
				c.On("Send", mock.Anything).Return(
					&notificationEntity.SendResult{StatusCode: http.StatusAccepted}, nil,
				)
			},
			wantErr:         true,
			wantErrContains: "unexpected status code 202",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := mockEmail.NewEmailClient(t)
			log := mockLogger.NewLogger(t)

			if tc.setupClient != nil {
				tc.setupClient(client)
			}
			if tc.setupLogger != nil {
				tc.setupLogger(log)
			}

			uc := usecase.NewEmailNotificationUseCase(client, log, newTemplates(t))

			err := uc.SendEmailNotification(context.Background(), tc.email)

			if tc.wantErr {
				assert.Error(t, err)
				if tc.wantErrContains != "" {
					assert.ErrorContains(t, err, tc.wantErrContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTemplates_RenderWelcome(t *testing.T) {
	tmpl := newTemplates(t)

	tests := []struct {
		name         string
		appURL       string
		wantContains []string
	}{
		{
			name:   "substitutes all three URLs",
			appURL: "https://dripmate.app",
			wantContains: []string{
				"https://dripmate.app",
				"https://dripmate.app/unsubscribe",
				"https://dripmate.app/privacy",
			},
		},
		{
			name:   "empty appURL produces relative paths",
			appURL: "",
			wantContains: []string{
				"/unsubscribe",
				"/privacy",
			},
		},
		{
			name:   "contains static content from template",
			appURL: "https://dripmate.app",
			wantContains: []string{
				"drip mate",
				"Добро пожаловать",
				"Начать подбор",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := tmpl.RenderWelcome(tc.appURL)

			assert.NoError(t, err)
			assert.NotEmpty(t, result)
			for _, s := range tc.wantContains {
				assert.Contains(t, result, s)
			}
		})
	}
}
