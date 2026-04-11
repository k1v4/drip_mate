package v1

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/k1v4/drip_mate/internal/modules/notification_service/entity"
	"github.com/k1v4/drip_mate/internal/modules/notification_service/usecase"
)

type EmailController struct {
	useCase usecase.IUseCase
}

func NewEmailController(useCase usecase.IUseCase) *EmailController {
	return &EmailController{
		useCase: useCase,
	}
}

func (c *EmailController) Handle(ctx context.Context, msg *entity.Message) error {
	var event entity.NotificationEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return fmt.Errorf("unmarshal event: %w", err)
	}

	if event.Email == "" {
		return fmt.Errorf("empty email")
	}

	err := c.useCase.SendEmailNotification(ctx, event.Email)
	if err != nil {
		return fmt.Errorf("failed to send email notification: %w", err)
	}

	return nil
}
