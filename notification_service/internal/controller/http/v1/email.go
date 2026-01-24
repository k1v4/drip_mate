package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"notification_service/internal/entity"
	"notification_service/internal/usecase"
)

type EmailController struct {
	useCase usecase.IUseCase
}

func NewEmailController(useCase usecase.IUseCase) *EmailController {
	return &EmailController{
		useCase: useCase,
	}
}

func (c *EmailController) Handle(ctx context.Context, msg entity.Message) error {
	var event entity.NotificationEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return fmt.Errorf("unmarshal event: %w", err)
	}

	if event.Text == "" {
		return fmt.Errorf("empty text")
	}

	if event.Email == "" {
		return fmt.Errorf("empty email")
	}

	err := c.useCase.SendEmailNotification(ctx, event.Text, event.Email)
	if err != nil {
		return fmt.Errorf("failed to send email notification: %w", err)
	}

	return nil
}
