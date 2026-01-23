package v1

import (
	"context"
	"notification_service/internal/entity"
	"notification_service/internal/usecase"
)

type Controller struct {
	useCase usecase.IUseCase
}

func (c *Controller) Handle(ctx context.Context, msg entity.Message) error {
	return c.useCase.SendEmailNotification(ctx, "", "")
}
