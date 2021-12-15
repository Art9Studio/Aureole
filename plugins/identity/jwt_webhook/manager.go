package jwt_webhook

import (
	"aureole/internal/configs"
	"aureole/internal/core"
	"aureole/internal/plugins"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/mitchellh/mapstructure"
)

const pluginID = "8483"

type manager struct {
	pluginApi core.PluginAPI
	app       *core.App
	rawConf   *configs.IDManager
	conf      *config
	client    http.Client
}

func (j *manager) Init(appName string, api core.PluginAPI) (err error) {
	j.pluginApi = api
	j.conf, err = initConfig(&j.rawConf.Config)
	if err != nil {
		return err
	}

	j.client = http.Client{Timeout: time.Duration(j.conf.Timeout) * time.Millisecond}
	j.app, err = j.pluginApi.GetApp(appName)
	if err != nil {
		return fmt.Errorf("app named '%s' is not declared", appName)
	}

	return nil
}

func (j *manager) GetMetaData() plugins.Meta {
	return plugins.Meta{
		Type: adapterName,
		Name: j.rawConf.Name,
		ID:   pluginID,
	}
}

func (j *manager) Register(c *plugins.Credential, i *plugins.Identity, authnProvider string) (*plugins.Identity, error) {
	requestToken, err := core.CreateJWT(map[string]interface{}{
		"event":          "Register",
		"credential":     map[string]string{c.Name: c.Value},
		"identity":       i.AsMap(),
		"authn_provider": authnProvider,
	},
		j.app.GetAuthSessionExp())
	if err != nil {
		return nil, err
	}

	respData, err := makeRequest(j, requestToken)
	if err != nil {
		return nil, err
	}
	rawToken, err := getJWT(respData)
	if err != nil {
		return nil, err
	}

	token, err := jwt.ParseString(rawToken)
	if err != nil {
		return nil, err
	}
	payload, err := token.AsMap(context.Background())
	if err != nil {
		return nil, err
	}

	return plugins.NewIdentity(payload)
}

func (j *manager) OnUserAuthenticated(c *plugins.Credential, i *plugins.Identity, authnProvider string) (*plugins.Identity, error) {
	requestToken, err := core.CreateJWT(map[string]interface{}{
		"event":          "OnUserAuthenticated",
		"credential":     map[string]string{c.Name: c.Value},
		"identity":       i.AsMap(),
		"authn_provider": authnProvider,
	},
		j.app.GetAuthSessionExp())
	if err != nil {
		return nil, err
	}

	respData, err := makeRequest(j, requestToken)
	if err != nil {
		return nil, err
	}
	rawToken, err := getJWT(respData)
	if err != nil {
		return nil, err
	}

	token, err := jwt.ParseString(rawToken)
	if err != nil {
		return nil, err
	}
	payload, err := token.AsMap(context.Background())
	if err != nil {
		return nil, err
	}

	return plugins.NewIdentity(payload)
}

func (j *manager) On2FA(c *plugins.Credential, mfaProvider string, data map[string]interface{}) error {
	requestToken, err := core.CreateJWT(map[string]interface{}{
		"event":        "On2FA",
		"credential":   map[string]string{c.Name: c.Value},
		"2fa_provider": mfaProvider,
		"2fa_data":     data,
	},
		j.app.GetAuthSessionExp())
	if err != nil {
		return err
	}

	_, err = makeRequest(j, requestToken)
	if err != nil {
		return err
	}
	return nil
}

func (j *manager) GetData(c *plugins.Credential, authnProvider, name string) (interface{}, error) {
	requestToken, err := core.CreateJWT(map[string]interface{}{
		"event":          "GetData",
		"credential":     map[string]string{c.Name: c.Value},
		"name":           name,
		"authn_provider": authnProvider,
	},
		j.app.GetAuthSessionExp())
	if err != nil {
		return nil, err
	}

	respData, err := makeRequest(j, requestToken)
	if err != nil {
		return nil, err
	}
	rawToken, err := getJWT(respData)
	if err != nil {
		return nil, err
	}

	token, err := jwt.ParseString(rawToken)
	if err != nil {
		return nil, err
	}
	data, ok := token.Get(name)
	if !ok {
		return nil, fmt.Errorf("cannot get '%s' field", name)
	}
	return data, nil
}

func (j *manager) Get2FAData(c *plugins.Credential, mfaProvider string) (map[string]interface{}, error) {
	requestToken, err := core.CreateJWT(map[string]interface{}{
		"event":        "Get2FAData",
		"credential":   map[string]string{c.Name: c.Value},
		"2fa_provider": mfaProvider,
	},
		j.app.GetAuthSessionExp())
	if err != nil {
		return nil, err
	}

	respData, err := makeRequest(j, requestToken)
	if err != nil {
		return nil, err
	}
	rawToken, err := getJWT(respData)
	if err != nil {
		return nil, err
	}

	token, err := jwt.ParseString(rawToken)
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	err = core.GetFromJWT(token, "2fa_data", &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (j *manager) Update(c *plugins.Credential, i *plugins.Identity, authnProvider string) (*plugins.Identity, error) {
	requestToken, err := core.CreateJWT(map[string]interface{}{
		"event":          "Update",
		"credential":     map[string]string{c.Name: c.Value},
		"identity":       i.AsMap(),
		"authn_provider": authnProvider,
	},
		j.app.GetAuthSessionExp())
	if err != nil {
		return nil, err
	}

	respData, err := makeRequest(j, requestToken)
	if err != nil {
		return nil, err
	}
	rawToken, err := getJWT(respData)
	if err != nil {
		return nil, err
	}

	token, err := jwt.ParseString(rawToken)
	if err != nil {
		return nil, err
	}

	var rawIdent map[string]interface{}
	err = core.GetFromJWT(token, "identity", &rawIdent)
	if err != nil {
		return nil, err
	}
	return plugins.NewIdentity(rawIdent)
}

func (j *manager) CheckFeaturesAvailable(features []string) error {
	requestToken, err := core.CreateJWT(map[string]interface{}{
		"event":    "CheckFeaturesAvailable",
		"features": features,
	},
		j.app.GetAuthSessionExp())
	if err != nil {
		return err
	}

	_, err = makeRequest(j, requestToken)
	if err != nil {
		return err
	}
	return nil
}

func initConfig(rawConf *configs.RawConfig) (*config, error) {
	adapterConf := &config{}
	if err := mapstructure.Decode(rawConf, adapterConf); err != nil {
		return nil, err
	}
	adapterConf.setDefaults()

	return adapterConf, nil
}

func makeRequest(j *manager, token string) ([]byte, error) {
	var respBytes []byte

	body, err := json.Marshal(map[string]string{"token": token})
	if err != nil {
		return nil, err
	}
	r, err := http.NewRequestWithContext(context.Background(), http.MethodPost, j.conf.Address, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	for k, v := range j.conf.Headers {
		r.Header.Set(k, v)
	}
	r.Header.Set("Content-Type", "application/json")

	err = retry.Do(
		func() error {
			resp, err := j.client.Do(r)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			_, err = resp.Body.Read(respBytes)
			return err
		},
		retry.DelayType(retry.FixedDelay),
		retry.Delay(time.Duration(j.conf.RetryInterval)*time.Millisecond),
		retry.Attempts(j.conf.RetriesNum),
	)
	if err != nil {
		return nil, err
	}
	return respBytes, nil
}

func getJWT(data []byte) (string, error) {
	var respData map[string]string
	err := json.Unmarshal(data, &respData)
	if err != nil {
		return "", err
	}

	if requestToken, ok := respData["request_token"]; ok {
		t, err := jwt.ParseString(requestToken)
		if err == nil {
			_ = core.InvalidateJWT(t)
		}
	}

	token, ok := respData["token"]
	if ok {
		return token, nil
	}
	return "", errors.New("cannot found token")
}
