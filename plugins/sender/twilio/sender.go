package twilio

import (
	"aureole/internal/configs"
	"aureole/internal/core"
	"aureole/internal/plugins"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	txtTmpl "text/template"

	"github.com/mitchellh/mapstructure"
)

const pluginID = "5116"

type (
	sender struct {
		pluginApi core.PluginAPI
		rawConf   *configs.Sender
		conf      *config
	}

	exception struct {
		Status  int
		Message string
	}
)

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

func (s *sender) Send(recipient, subject, tmplStr, _ string, tmplCtx map[string]interface{}) error {
	message := &bytes.Buffer{}

	tmpl, err := txtTmpl.New("message").Parse(tmplStr)
	if err != nil {
		return err
	}

	err = tmpl.Execute(message, tmplCtx)
	if err != nil {
		return err
	}
	return s.SendRaw(recipient, subject, message.String())
}

func (s *sender) SendRaw(recipient, _, message string) error {
	endpoint := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", s.conf.AccountSid)
	data := url.Values{}
	data.Set("Body", message)
	data.Set("From", s.conf.From)
	data.Set("To", recipient)

	ctx := context.Background()
	r, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	r.SetBasicAuth(s.conf.AccountSid, s.conf.AuthToken)
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
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
