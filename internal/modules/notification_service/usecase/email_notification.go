package usecase

import (
	"context"
	"fmt"
	"net/http"

	"github.com/k1v4/drip_mate/internal/modules/notification_service"
	"github.com/k1v4/drip_mate/internal/modules/notification_service/entity"
	"github.com/labstack/gommon/log"
)

type EmailNotificationUseCase struct {
	emailAdapter EmailClient
	welcomeTmpl  *notification_service.Templates
}

func NewEmailNotificationUseCase(emailAdapter EmailClient, welcomeTmpl *notification_service.Templates) *EmailNotificationUseCase {
	return &EmailNotificationUseCase{
		emailAdapter: emailAdapter,
		welcomeTmpl:  welcomeTmpl,
	}
}

func (en *EmailNotificationUseCase) SendEmailNotification(ctx context.Context, email string) error {
	var html, plainText string

	welcomeMsg, err := en.welcomeTmpl.RenderWelcome("google.com")
	if err != nil {
		// TODO прокинуть общий логгер
		log.Errorf("failed to render welcome template: %v", err)
		plainText = "Добро пожаловать в drip mate! Мы рады видеть вас."
	} else {
		html = welcomeMsg
	}

	emailInfo := &entity.Email{
		To:      []string{email},
		Subject: "Добро пожаловать в drip mate",
		HTML:    html,
		Text:    plainText,
	}

	result, err := en.emailAdapter.Send(emailInfo)
	if err != nil {
		return fmt.Errorf("send email: %w", err)
	}

	if result.StatusCode != http.StatusOK {
		return fmt.Errorf("send email: unexpected status code %d", result.StatusCode)
	}

	return nil
}
