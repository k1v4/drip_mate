package usecase

import (
	"context"
	"notification_service/internal/entity"
)

type IHandler interface {
	Handle(ctx context.Context, msg entity.Message) error
}

type IUseCase interface {
	SendEmailNotification(ctx context.Context, text string, email string) error
}
