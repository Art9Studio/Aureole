package email

import (
	"aureole/internal/configs"
	"aureole/internal/core"
	"bytes"
	"crypto/tls"
	htmlTmpl "html/template"
	"net/smtp"
	"strings"
	txtTmpl "text/template"

	emailLib "github.com/jordan-wright/email"
	"github.com/mitchellh/mapstructure"
)

// const pluginID = "5151"
var rawMeta []byte

var meta core.Meta

func init() {
	meta = core.SenderRepo.Register(rawMeta, Create)
}

type email struct {
	pluginApi core.PluginAPI
	rawConf   configs.PluginConfig
	conf      *config
}

func Create(conf configs.PluginConfig) core.Sender {
	return &email{rawConf: conf}
}

func (e *email) Init(api core.PluginAPI) error {
	e.pluginApi = api
	PluginConf := &config{}
	if err := mapstructure.Decode(e.rawConf.Config, PluginConf); err != nil {
		return err
	}
	e.conf = PluginConf

	return nil
}

func (e email) GetMetaData() core.Meta {
	return meta
}

func (e *email) Send(recipient, subject, tmplStr, tmplExt string, tmplCtx map[string]interface{}) error {
	mail := &emailLib.Email{
		From:    e.conf.From,
		To:      []string{recipient},
		Bcc:     e.conf.Bcc,
		Cc:      e.conf.Cc,
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
	hostname := strings.Split(e.conf.Host, ":")[0]
	plainAuth := smtp.PlainAuth("", e.conf.Username, e.conf.Password, hostname)

	if e.conf.InsecureSkipVerify {
		return mail.SendWithStartTLS(e.conf.Host, plainAuth, &tls.Config{InsecureSkipVerify: true})
	}
	return mail.Send(e.conf.Host, plainAuth)
}

func (e *email) SendRaw(recipient, subject, message string) error {
	mail := &emailLib.Email{
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

	if e.conf.InsecureSkipVerify {
		return mail.SendWithStartTLS(e.conf.Host, plainAuth, &tls.Config{InsecureSkipVerify: true})
	}
	return mail.Send(e.conf.Host, plainAuth)
}
func (e *email) GetAppRoutes() []*core.Route {
	return []*core.Route{}

}
