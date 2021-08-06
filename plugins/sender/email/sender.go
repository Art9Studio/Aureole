package email

import (
	"aureole/internal/configs"
	"aureole/internal/plugins/sender"
	"bytes"
	"crypto/tls"
	htmlTmpl "html/template"
	"net/smtp"
	"path"
	"strings"
	txtTmpl "text/template"

	"github.com/jordan-wright/email"
	"github.com/mitchellh/mapstructure"
)

type Email struct {
	rawConf *configs.Sender
	conf    *config
}

func (e *Email) Init() error {
	adapterConf := &config{}
	if err := mapstructure.Decode(e.rawConf.Config, adapterConf); err != nil {
		return err
	}
	e.conf = adapterConf

	return nil
}

func (e *Email) Send(recipient, subject, tmplName string, tmplCtx map[string]interface{}) error {
	mail := &email.Email{
		From:    e.conf.From,
		To:      []string{recipient},
		Bcc:     e.conf.Bcc,
		Cc:      e.conf.Cc,
		Subject: subject,
	}

	tmplFileName := e.conf.Templates[tmplName]
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
	hostname := strings.Split(e.conf.Host, ":")[0]
	plainAuth := smtp.PlainAuth("", e.conf.Username, e.conf.Password, hostname)

	if sender.Repository.PluginApi.Project.IsTestRun() {
		return mail.SendWithStartTLS(e.conf.Host, plainAuth, &tls.Config{InsecureSkipVerify: true})
	}
	return mail.Send(e.conf.Host, plainAuth)
}

func (e *Email) SendRaw(recipient, subject, message string) error {
	mail := &email.Email{
		From:    e.conf.From,
		To:      []string{recipient},
		Bcc:     e.conf.Bcc,
		Cc:      e.conf.Cc,
		Subject: subject,
		Text:    []byte(message),
	}

	// todo: test custom ports support
	hostname := strings.Split(e.conf.Host, ":")[0]
	plainAuth := smtp.PlainAuth("", e.conf.Username, e.conf.Password, hostname)

	if sender.Repository.PluginApi.Project.IsTestRun() {
		return mail.SendWithStartTLS(e.conf.Host, plainAuth, &tls.Config{InsecureSkipVerify: true})
	}
	return mail.Send(e.conf.Host, plainAuth)
}
