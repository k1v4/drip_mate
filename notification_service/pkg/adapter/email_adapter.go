package adapter

import (
	"fmt"
	"notification_service/internal/entity"
	"os"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// SendGridClient реализация EmailClient для SendGrid API.
type SendGridClient struct {
	client *sendgrid.Client
}

// NewSendGridClient создаёт нового клиента.
// apiKey можно передать явно или через переменную окружения SENDGRID_API_KEY.
func NewSendGridClient(apiKey string) *SendGridClient {
	if apiKey == "" {
		apiKey = os.Getenv("SENDGRID_API_KEY")
	}

	if apiKey == "" {
		return nil
	}

	return &SendGridClient{
		client: sendgrid.NewSendClient(apiKey),
	}
}

// Send отправляет письмо через SendGrid.
func (s *SendGridClient) Send(email *entity.Email) (*entity.SendResult, error) {
	if email == nil {
		return nil, fmt.Errorf("email information is nil")
	}
	if len(email.To) == 0 {
		return nil, fmt.Errorf("no recipient specified")
	}

	from := mail.NewEmail("", email.From)
	message := mail.NewV3Mail()
	message.SetFrom(from)
	message.Subject = email.Subject

	personalization := mail.NewPersonalization()
	for _, toAddr := range email.To {
		to := mail.NewEmail("", toAddr)
		personalization.AddTos(to)
	}
	message.AddPersonalizations(personalization)

	if email.Text != "" {
		message.AddContent(mail.NewContent("text/plain", email.Text))
	}
	if email.HTML != "" {
		message.AddContent(mail.NewContent("text/html", email.HTML))
	}

	resp, err := s.client.Send(message)
	if err != nil {
		return nil, fmt.Errorf("sendgrid send error: %w", err)
	}

	return &entity.SendResult{
		StatusCode: resp.StatusCode,
		Body:       resp.Body,
		Headers:    resp.Headers,
	}, nil
}
