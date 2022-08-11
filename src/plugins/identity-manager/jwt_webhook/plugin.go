package jwt_webhook

import (
	"aureole/internal/configs"
	"aureole/internal/core"
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"net/http"
	"time"

	"github.com/avast/retry-go/v4"

	"github.com/lestrrat-go/jwx/jwt"
	"github.com/mitchellh/mapstructure"
)

// const pluginID = "8483"

//go:embed meta.yaml
var rawMeta []byte

var meta core.Metadata

// init initializes package by register pluginCreator
func init() {
	meta = core.IDManagerRepo.Register(rawMeta, Create)
}

type webhook struct {
	pluginAPI core.PluginAPI
	rawConf   configs.PluginConfig
	conf      *config
	client    http.Client
}

func Create(conf configs.PluginConfig) core.IDManager {
	return &webhook{rawConf: conf}
}

func (j *webhook) Init(api core.PluginAPI) (err error) {
	j.pluginAPI = api
	j.conf, err = initConfig(&j.rawConf.Config)
	if err != nil {
		return err
	}
	j.client = http.Client{Timeout: time.Duration(j.conf.Timeout) * time.Millisecond}
	return nil
}

func (webhook) GetMetadata() core.Metadata {
	return meta
}

func (j *webhook) GetCustomAppRoutes() []*core.Route {
	return []*core.Route{}
}

func (j *webhook) Register(c *core.Credential, i *core.Identity, u *core.User, authnProvider string) (*core.User, error) {
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

	return core.NewUser(payload)
}

func (j *webhook) OnUserAuthenticated(authRes *core.AuthResult) (*core.AuthResult, error) {
	requestToken, err := j.pluginAPI.CreateJWT(map[string]interface{}{
		"event":          "OnUserAuthenticated",
		"credential":     map[string]string{authRes.Cred.Name: authRes.Cred.Value},
		"identity":       authRes.Identity.AsMap(),
		"authn_provider": authRes.Provider,
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
	newUser, err := core.NewUser(payload)
	authRes.User = newUser

	return authRes, nil
}

func (j *webhook) OnMFA(c *core.Credential, mfaData *core.MFAData) error {
	requestToken, err := j.pluginAPI.CreateJWT(map[string]interface{}{
		"event":      "OnMFA",
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

func (j *webhook) GetData(c *core.Credential, authnProvider, name string) (interface{}, error) {
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

func (j *webhook) GetMFAData(c *core.Credential, mfaID string) (*core.MFAData, error) {
	requestToken, err := j.pluginAPI.CreateJWT(map[string]interface{}{
		"event":      "GetMFAData",
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

	var data core.MFAData
	err = j.pluginAPI.GetFromJWT(token, "2fa_data", &data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (j *webhook) Update(c *core.Credential, i *core.Identity, authnProvider string) (*core.Identity, error) {
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
	return core.NewIdentity(rawIdent)
}

func (j *webhook) CheckFeaturesAvailable(features []string) error {
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

func initConfig(conf *configs.RawConfig) (*config, error) {
	PluginConf := &config{}
	if err := mapstructure.Decode(conf, PluginConf); err != nil {
		return nil, err
	}
	PluginConf.setDefaults()

	return PluginConf, nil
}

func makeRequest(j *webhook, token string) ([]byte, error) {
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
	r.Header.Set("Content-Type", fiber.MIMEApplicationJSON)

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
