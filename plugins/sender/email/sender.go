package email

import (
	"bytes"
	"github.com/jordan-wright/email"
	htmlTmpl "html/template"
	"net/smtp"
	"path"
	"strings"
	txtTmpl "text/template"
)

type Email struct {
	Conf *config
}

func (e *Email) Send(recipient, subject, tmplName string, tmplCtx map[string]interface{}) error {
	mail := &email.Email{
		From:    e.Conf.From,
		To:      []string{recipient},
		Bcc:     e.Conf.Bcc,
		Cc:      e.Conf.Cc,
		Subject: subject,
	}

	tmplFileName := e.Conf.Templates[tmplName]
	baseName := path.Base(tmplFileName)
	extension := path.Ext(tmplFileName)
	message := &bytes.Buffer{}

	if extension == ".html" {
		tmpl := htmlTmpl.Must(htmlTmpl.New(baseName).ParseFiles(tmplFileName))
		if err := tmpl.Execute(message, tmplCtx); err != nil {
			return err
		}

		mail.HTML = message.Bytes()
	} else {
		tmpl := txtTmpl.Must(txtTmpl.New(baseName).ParseFiles(tmplFileName))
		if err := tmpl.Execute(message, tmplCtx); err != nil {
			return err
		}

		mail.Text = message.Bytes()
	}

	// todo: test custom ports support
	hostname := strings.Split(e.Conf.Host, ":")[0]
	plainAuth := smtp.PlainAuth("", e.Conf.Username, e.Conf.Password, hostname)

	return mail.Send(e.Conf.Host, plainAuth)
}

func (e *Email) SendRaw(recipient, subject, message string) error {
	mail := &email.Email{
		From:    e.Conf.From,
		To:      []string{recipient},
		Bcc:     e.Conf.Bcc,
		Cc:      e.Conf.Cc,
		Subject: subject,
		Text:    []byte(message),
	}

	// todo: test custom ports support
	hostname := strings.Split(e.Conf.Host, ":")[0]
	plainAuth := smtp.PlainAuth("", e.Conf.Username, e.Conf.Password, hostname)

	return mail.Send(e.Conf.Host, plainAuth)
}
