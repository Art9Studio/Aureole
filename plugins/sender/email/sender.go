package email

import (
	"aureole/internal/configs"
	"aureole/internal/core"
	"aureole/internal/plugins"
	"bytes"
	"crypto/tls"
	htmlTmpl "html/template"
	"net/smtp"
	"strings"
	txtTmpl "text/template"

	"github.com/jordan-wright/email"
	"github.com/mitchellh/mapstructure"
)

const pluginID = "5151"

type sender struct {
	pluginApi core.PluginAPI
	rawConf   *configs.Sender
	conf      *config
}

func (s *sender) Init(api core.PluginAPI) error {
	s.pluginApi = api
	adapterConf := &config{}
	if err := mapstructure.Decode(s.rawConf.Config, adapterConf); err != nil {
		return err
	}
	s.conf = adapterConf

	return nil
}

func (s *sender) GetMetaData() plugins.Meta {
	return plugins.Meta{
		Type: adapterName,
		Name: s.rawConf.Name,
		ID:   pluginID,
	}
}

func (s *sender) Send(recipient, subject, tmplStr, tmplExt string, tmplCtx map[string]interface{}) error {
	mail := &email.Email{
		From:    s.conf.From,
		To:      []string{recipient},
		Bcc:     s.conf.Bcc,
		Cc:      s.conf.Cc,
		Subject: subject,
	}

	message := &bytes.Buffer{}

	if tmplExt == ".html" {
		tmpl, err := htmlTmpl.New("message").Parse(tmplStr)
		if err != nil {
			return err
		}

		err = tmpl.Execute(message, tmplCtx)
		if err != nil {
			return err
		}
		mail.HTML = message.Bytes()
	} else {
		tmpl, err := txtTmpl.New("message").Parse(tmplStr)
		if err != nil {
			return err
		}

		err = tmpl.Execute(message, tmplCtx)
		if err != nil {
			return err
		}
		mail.Text = message.Bytes()
	}

	// todo: test custom ports support
	hostname := strings.Split(s.conf.Host, ":")[0]
	plainAuth := smtp.PlainAuth("", s.conf.Username, s.conf.Password, hostname)

	if s.conf.InsecureSkipVerify {
		return mail.SendWithStartTLS(s.conf.Host, plainAuth, &tls.Config{InsecureSkipVerify: true})
	}
	return mail.Send(s.conf.Host, plainAuth)
}

func (s *sender) SendRaw(recipient, subject, message string) error {
	mail := &email.Email{
		From:    s.conf.From,
		To:      []string{recipient},
		Bcc:     s.conf.Bcc,
		Cc:      s.conf.Cc,
		Subject: subject,
		Text:    []byte(message),
	}

	// todo: test custom ports support
	hostname := strings.Split(s.conf.Host, ":")[0]
	plainAuth := smtp.PlainAuth("", s.conf.Username, s.conf.Password, hostname)

	if s.conf.InsecureSkipVerify {
		return mail.SendWithStartTLS(s.conf.Host, plainAuth, &tls.Config{InsecureSkipVerify: true})
	}
	return mail.Send(s.conf.Host, plainAuth)
}
