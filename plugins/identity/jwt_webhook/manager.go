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
	pluginAPI core.PluginAPI
	rawConf   *configs.IDManager
	conf      *config
	client    http.Client
}

func (j *manager) Init(api core.PluginAPI) (err error) {
	j.pluginAPI = api
	j.conf, err = initConfig(&j.rawConf.Config)
	if err != nil {
		return err
	}
	j.client = http.Client{Timeout: time.Duration(j.conf.Timeout) * time.Millisecond}
	return nil
}

func (j *manager) GetMetaData() plugins.Meta {
	return plugins.Meta{
		Type: adapterName,
		ID:   pluginID,
	}
}

func (j *manager) Register(c *plugins.Credential, i *plugins.Identity, authnProvider string) (*plugins.Identity, error) {
	requestToken, err := j.pluginAPI.CreateJWT(map[string]interface{}{
		"event":          "Register",
		"credential":     map[string]string{c.Name: c.Value},
		"identity":       i.AsMap(),
		"authn_provider": authnProvider,
	},
		j.pluginAPI.GetAuthSessionExp())
	if err != nil {
		return nil, err
	}

	respData, err := makeRequest(j, requestToken)
	if err != nil {
		return nil, err
	}
	rawToken, err := getJWT(j.pluginAPI, respData)
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
	requestToken, err := j.pluginAPI.CreateJWT(map[string]interface{}{
		"event":          "OnUserAuthenticated",
		"credential":     map[string]string{c.Name: c.Value},
		"identity":       i.AsMap(),
		"authn_provider": authnProvider,
	},
		j.pluginAPI.GetAuthSessionExp())
	if err != nil {
		return nil, err
	}

	respData, err := makeRequest(j, requestToken)
	if err != nil {
		return nil, err
	}
	rawToken, err := getJWT(j.pluginAPI, respData)
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

func (j *manager) On2FA(c *plugins.Credential, mfaData *plugins.MFAData) error {
	requestToken, err := j.pluginAPI.CreateJWT(map[string]interface{}{
		"event":      "On2FA",
		"credential": map[string]string{c.Name: c.Value},
		"2fa_data":   mfaData,
	},
		j.pluginAPI.GetAuthSessionExp())
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
	requestToken, err := j.pluginAPI.CreateJWT(map[string]interface{}{
		"event":          "GetData",
		"credential":     map[string]string{c.Name: c.Value},
		"name":           name,
		"authn_provider": authnProvider,
	},
		j.pluginAPI.GetAuthSessionExp())
	if err != nil {
		return nil, err
	}

	respData, err := makeRequest(j, requestToken)
	if err != nil {
		return nil, err
	}
	rawToken, err := getJWT(j.pluginAPI, respData)
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

func (j *manager) Get2FAData(c *plugins.Credential, mfaID string) (*plugins.MFAData, error) {
	requestToken, err := j.pluginAPI.CreateJWT(map[string]interface{}{
		"event":      "Get2FAData",
		"credential": map[string]string{c.Name: c.Value},
		"2fa_id":     mfaID,
	},
		j.pluginAPI.GetAuthSessionExp())
	if err != nil {
		return nil, err
	}

	respData, err := makeRequest(j, requestToken)
	if err != nil {
		return nil, err
	}
	rawToken, err := getJWT(j.pluginAPI, respData)
	if err != nil {
		return nil, err
	}

	token, err := jwt.ParseString(rawToken)
	if err != nil {
		return nil, err
	}

	var data plugins.MFAData
	err = j.pluginAPI.GetFromJWT(token, "2fa_data", &data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (j *manager) Update(c *plugins.Credential, i *plugins.Identity, authnProvider string) (*plugins.Identity, error) {
	requestToken, err := j.pluginAPI.CreateJWT(map[string]interface{}{
		"event":          "Update",
		"credential":     map[string]string{c.Name: c.Value},
		"identity":       i.AsMap(),
		"authn_provider": authnProvider,
	},
		j.pluginAPI.GetAuthSessionExp())
	if err != nil {
		return nil, err
	}

	respData, err := makeRequest(j, requestToken)
	if err != nil {
		return nil, err
	}
	rawToken, err := getJWT(j.pluginAPI, respData)
	if err != nil {
		return nil, err
	}

	token, err := jwt.ParseString(rawToken)
	if err != nil {
		return nil, err
	}

	var rawIdent map[string]interface{}
	err = j.pluginAPI.GetFromJWT(token, "identity", &rawIdent)
	if err != nil {
		return nil, err
	}
	return plugins.NewIdentity(rawIdent)
}

func (j *manager) CheckFeaturesAvailable(features []string) error {
	requestToken, err := j.pluginAPI.CreateJWT(map[string]interface{}{
		"event":    "CheckFeaturesAvailable",
		"features": features,
	},
		j.pluginAPI.GetAuthSessionExp())
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
		retry.Attempts(uint(j.conf.RetriesNum)),
	)
	if err != nil {
		return nil, err
	}
	return respBytes, nil
}

func getJWT(api core.PluginAPI, data []byte) (string, error) {
	var respData map[string]string
	err := json.Unmarshal(data, &respData)
	if err != nil {
		return "", err
	}

	if requestToken, ok := respData["request_token"]; ok {
		t, err := jwt.ParseString(requestToken)
		if err == nil {
			_ = api.InvalidateJWT(t)
		}
	}

	token, ok := respData["token"]
	if ok {
		return token, nil
	}
	return "", errors.New("cannot found token")
}
