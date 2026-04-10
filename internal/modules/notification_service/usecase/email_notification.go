package usecase

import (
	"context"
	"fmt"
	"net/http"

	"github.com/k1v4/drip_mate/internal/modules/notification_service/entity"
)

type EmailNotificationUseCase struct {
	emailAdapter EmailClient
}

func NewEmailNotificationUseCase(emailAdapter EmailClient) *EmailNotificationUseCase {
	return &EmailNotificationUseCase{
		emailAdapter: emailAdapter,
	}
}

func (en *EmailNotificationUseCase) SendEmailNotification(ctx context.Context, text string, email string) error {
	emailInfo := &entity.Email{
		From:    "myanymail@mail.ru",
		To:      []string{email},
		Subject: "Welcome letter!",
		HTML:    text,
		Text:    text,
	}

	result, err := en.emailAdapter.Send(emailInfo)
	if err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	if result.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send email via not success code response")
	}

	return nil
}
