package usecase

import (
	"context"
	"notification_service/internal/entity"
)

// IHandler интерфейс для чтения сообщения из кафки
type IHandler interface {
	Handle(ctx context.Context, msg entity.Message) error
}

type IUseCase interface {
	SendEmailNotification(ctx context.Context, text string, email string) error
}

// EmailClient интерфейс для отправки email
type EmailClient interface {
	Send(email *entity.Email) (*entity.SendResult, error)
}
