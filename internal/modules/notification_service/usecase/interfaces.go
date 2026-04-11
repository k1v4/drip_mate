package usecase

import (
	"context"

	"github.com/k1v4/drip_mate/internal/entity"
	notificationEntity "github.com/k1v4/drip_mate/internal/modules/notification_service/entity"
)

// IHandler интерфейс для чтения сообщения из кафки
type IHandler interface {
	Handle(ctx context.Context, event *entity.NotificationEvent) error
}

type IUseCase interface {
	SendEmailNotification(ctx context.Context, email string) error
}

// EmailClient интерфейс для отправки email
type EmailClient interface {
	Send(email *notificationEntity.Email) (*notificationEntity.SendResult, error)
}
