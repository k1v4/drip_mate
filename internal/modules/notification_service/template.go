package notification_service

import (
	"bytes"
	"embed"
	"html/template"
)

//go:embed templates/welcome.html
var templateFS embed.FS

type Templates struct {
	welcome *template.Template
}

func NewTemplates() (*Templates, error) {
	welcome, err := template.ParseFS(templateFS, "templates/welcome.html")
	if err != nil {
		return nil, err
	}
	return &Templates{welcome: welcome}, nil
}

type welcomeData struct {
	AppURL         string
	UnsubscribeURL string
	PrivacyURL     string
}

func (t *Templates) RenderWelcome(appURL string) (string, error) {
	var buf bytes.Buffer
	err := t.welcome.Execute(&buf, welcomeData{
		AppURL:         appURL,
		UnsubscribeURL: appURL + "/unsubscribe",
		PrivacyURL:     appURL + "/privacy",
	})
	return buf.String(), err
}
