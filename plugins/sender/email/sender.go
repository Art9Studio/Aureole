package email

import (
	"github.com/jordan-wright/email"
	"net/smtp"
	"strings"
)

type Email struct {
	Conf *config
}

func (e *Email) Send(recipient string, templateName string, templateData map[string]interface{}) error {
	panic("implement me")
}

func (e *Email) SendRaw(recipient string, message string) error {
	mail := &email.Email{
		From:    e.Conf.From,
		To:      []string{recipient},
		Bcc:     e.Conf.Bcc,
		Cc:      e.Conf.Cc,
		Subject: e.Conf.Subject,
		Text:    []byte(message),
	}

	hostname := strings.Split(e.Conf.Host, ":")[0]
	plainAuth := smtp.PlainAuth("", e.Conf.Username, e.Conf.Password, hostname)

	return mail.Send(e.Conf.Host, plainAuth)
}
