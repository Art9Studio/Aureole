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
	"path"
	"strconv"
	"strings"
	txtTmpl "text/template"

	"github.com/mitchellh/mapstructure"
)

const pluginID = "5116"

type (
	twilio struct {
		pluginApi core.PluginAPI
		rawConf   *configs.Sender
		conf      *config
	}

	exception struct {
		Status  int
		Message string
	}
)

func (t *twilio) Init(api core.PluginAPI) error {
	t.pluginApi = api
	adapterConf := &config{}
	if err := mapstructure.Decode(t.rawConf.Config, adapterConf); err != nil {
		return err
	}
	t.conf = adapterConf

	return nil
}

func (t *twilio) GetMetaData() plugins.Meta {
	return plugins.Meta{
		Type: adapterName,
		Name: t.rawConf.Name,
		ID:   pluginID,
	}
}

func (t *twilio) Send(recipient, subject, tmplName string, tmplCtx map[string]interface{}) error {
	tmplFileName := t.conf.Templates[tmplName]
	baseName := path.Base(tmplFileName)
	message := &bytes.Buffer{}

	tmpl := txtTmpl.Must(txtTmpl.New(baseName).ParseFiles(tmplFileName))
	if err := tmpl.Execute(message, tmplCtx); err != nil {
		return err
	}

	return t.SendRaw(recipient, subject, message.String())
}

func (t *twilio) SendRaw(recipient, subject, message string) error {
	endpoint := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", t.conf.AccountSid)
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
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

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