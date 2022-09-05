package twilio

import (
	"aureole/internal/configs"
	"aureole/internal/core"
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	txtTmpl "text/template"

	"github.com/mitchellh/mapstructure"
)

// const pluginID = "5116"
//go:embed meta.yaml
var rawMeta []byte

var meta core.Metadata

func init() {
	meta = core.SenderRepo.Register(rawMeta, Create)
}

type (
	twilio struct {
		pluginApi core.PluginAPI
		rawConf   configs.PluginConfig
		conf      *config
	}

	exception struct {
		Status  int
		Message string
	}
)

func Create(conf configs.PluginConfig) core.Sender {
	return &twilio{rawConf: conf}
}

func (t *twilio) Init(api core.PluginAPI) error {
	t.pluginApi = api
	PluginConf := &config{}
	if err := mapstructure.Decode(t.rawConf.Config, PluginConf); err != nil {
		return err
	}
	t.conf = PluginConf

	return nil
}

func (t *twilio) GetMetadata() core.Metadata {
	return meta
}

func (t *twilio) Send(recipient, subject, tmplStr, _ string, tmplCtx map[string]interface{}) error {
	message := &bytes.Buffer{}

	tmpl, err := txtTmpl.New("message").Parse(tmplStr)
	if err != nil {
		return err
	}

	err = tmpl.Execute(message, tmplCtx)
	if err != nil {
		return err
	}
	return t.SendRaw(recipient, subject, message.String())
}

func (t *twilio) SendRaw(recipient, _, message string) error {
	endpoint := fmt.Sprintf("%s/%s/%s", t.conf.Endpoint, t.conf.AccountSid, t.conf.MessageType)
	data := url.Values{}
	data.Set("Body", message)
	data.Set("From", t.conf.From)
	data.Set("To", recipient)

	ctx := context.Background()
	r, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	r.SetBasicAuth(t.conf.AccountSid, t.conf.AuthToken)
	r.Header.Set("Content-Type", fiber.MIMEApplicationForm)
	r.Header.Set("Content-Length", strconv.Itoa(len(data.Encode())))

	res, err := http.DefaultClient.Do(r)
	if err != nil {
		return err
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		defer res.Body.Close()
		e := &exception{}
		if err := json.NewDecoder(res.Body).Decode(e); err != nil {
			return err
		}
		return fmt.Errorf("twilio error occurred: status: %d; message: %s", e.Status, e.Message)
	}

	return nil
}
func (t *twilio) GetCustomAppRoutes() []*core.Route {
	return []*core.Route{}
}
