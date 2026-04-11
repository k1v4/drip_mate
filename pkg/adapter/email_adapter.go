package adapter

import (
	"crypto/tls"
	"fmt"

	"github.com/k1v4/drip_mate/internal/config"
	"github.com/k1v4/drip_mate/internal/modules/notification_service/entity"
	"gopkg.in/gomail.v2"
)

type GoMailClient struct {
	dialer *gomail.Dialer
	from   string
}

func NewGoMailClient(cfg config.SMTP) *GoMailClient {
	dialer := gomail.NewDialer(cfg.Host, cfg.Port, cfg.Username, cfg.Password)
	dialer.TLSConfig = &tls.Config{ServerName: cfg.Host}

	return &GoMailClient{
		dialer: dialer,
		from:   cfg.Username,
	}
}

func (g *GoMailClient) Send(email *entity.Email) (*entity.SendResult, error) {
	if email == nil {
		return nil, fmt.Errorf("email information is nil")
	}
	if len(email.To) == 0 {
		return nil, fmt.Errorf("no recipient specified")
	}

	m := gomail.NewMessage()
	m.SetHeader("From", g.from)
	m.SetHeader("To", email.To...)
	m.SetHeader("Subject", email.Subject)

	if email.HTML != "" {
		m.SetBody("text/html", email.HTML)
	} else if email.Text != "" {
		m.SetBody("text/plain", email.Text)
	}

	if err := g.dialer.DialAndSend(m); err != nil {
		return nil, fmt.Errorf("smtp send error: %w", err)
	}

	return &entity.SendResult{
		StatusCode: 200,
	}, nil
}
