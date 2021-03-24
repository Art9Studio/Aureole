package email

import (
	"bytes"
	"github.com/jordan-wright/email"
	"net/smtp"
	"path"
	"strings"
	"text/template"
)

type Email struct {
	Conf *config
}

func (e *Email) Send(recipient string, subject string, tmplFileName string, tmplCtx map[string]interface{}) error {
	baseName := path.Base(tmplFileName)
	tmpl := template.Must(template.New(baseName).ParseFiles(tmplFileName))

	message := &bytes.Buffer{}
	if err := tmpl.Execute(message, tmplCtx); err != nil {
		return err
	}

	mail := &email.Email{
		From:    e.Conf.From,
		To:      []string{recipient},
		Bcc:     e.Conf.Bcc,
		Cc:      e.Conf.Cc,
		Subject: subject,
		Text:    message.Bytes(),
	}

	// todo: test custom ports support
	hostname := strings.Split(e.Conf.Host, ":")[0]
	plainAuth := smtp.PlainAuth("", e.Conf.Username, e.Conf.Password, hostname)

	return mail.Send(e.Conf.Host, plainAuth)
}

func (e *Email) SendRaw(recipient string, subject string, message string) error {
	mail := &email.Email{
		From:    e.Conf.From,
		To:      []string{recipient},
		Bcc:     e.Conf.Bcc,
		Cc:      e.Conf.Cc,
		Subject: subject,
		Text:    []byte(message),
	}

	// todo: test custom ports support
	//hostname := strings.Split(e.Conf.Host, ":")[0]
	plainAuth := smtp.PlainAuth("", e.Conf.Username, e.Conf.Password, e.Conf.Host)

	return mail.Send(e.Conf.Host, plainAuth)
}
