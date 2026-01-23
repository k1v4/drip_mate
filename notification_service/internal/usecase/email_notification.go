package usecase

import (
	"context"
)

type EmailNotificationUseCase struct {
}

func NewEmailNotificationUseCase() *EmailNotificationUseCase {
	return &EmailNotificationUseCase{}
}

func (en *EmailNotificationUseCase) SendEmailNotification(ctx context.Context, text string, email string) error {
	return nil
}
