package v1_test

import (
	"context"
	"errors"
	"testing"

	"github.com/k1v4/drip_mate/internal/entity"
	v1 "github.com/k1v4/drip_mate/internal/modules/notification_service/controller/http/v1"
	mockUseCase "github.com/k1v4/drip_mate/mocks/internal_/modules/notification_service/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestEmailController_Handle(t *testing.T) {
	tests := []struct {
		name            string
		event           *entity.NotificationEvent
		setupUseCase    func(u *mockUseCase.IUseCase)
		wantErr         bool
		wantErrContains string
	}{
		{
			name:  "success",
			event: &entity.NotificationEvent{Email: "user@example.com"},
			setupUseCase: func(u *mockUseCase.IUseCase) {
				u.On("SendEmailNotification", mock.Anything, "user@example.com").Return(nil)
			},
			wantErr: false,
		},
		{
			name:            "empty email — returns error without calling usecase",
			event:           &entity.NotificationEvent{Email: ""},
			wantErr:         true,
			wantErrContains: "empty email",
		},
		{
			name:  "usecase error — wrapped and returned",
			event: &entity.NotificationEvent{Email: "user@example.com"},
			setupUseCase: func(u *mockUseCase.IUseCase) {
				u.On("SendEmailNotification", mock.Anything, "user@example.com").
					Return(errors.New("smtp error"))
			},
			wantErr:         true,
			wantErrContains: "failed to send email notification",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			uc := mockUseCase.NewIUseCase(t)

			if tc.setupUseCase != nil {
				tc.setupUseCase(uc)
			}

			controller := v1.NewEmailController(uc)
			err := controller.Handle(context.Background(), tc.event)

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
