package phone

import (
	"aureole/internal/configs"
	"aureole/internal/core"
	"aureole/internal/plugins"
	"errors"
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/mitchellh/mapstructure"
)

const pluginID = "6937"

type (
	phone struct {
		pluginApi core.PluginAPI
		appName   string
		rawConf   *configs.Authn
		conf      *config
		sender    plugins.Sender
	}

	input struct {
		Phone string `json:"phone"`
		Token string `json:"token"`
		Otp   string `json:"otp"`
	}
)

func (p *phone) Init(appName string, api core.PluginAPI) (err error) {
	p.pluginApi = api
	p.appName = appName
	p.conf, err = initConfig(&p.rawConf.Config)
	if err != nil {
		return err
	}

	p.sender, err = p.pluginApi.GetSender(p.conf.Sender)
	if err != nil {
		return fmt.Errorf("sender named '%s' is not declared", p.conf.Sender)
	}

	createRoutes(p)
	return nil
}

func (*phone) GetMetaData() plugins.Meta {
	return plugins.Meta{
		Type: adapterName,
		ID:   pluginID,
	}
}

func (p *phone) Login() plugins.AuthNLoginFunc {
	return func(c fiber.Ctx) (*plugins.AuthNResult, error) {
		var input *input
		if err := c.BodyParser(input); err != nil {
			return nil, err
		}
		if input.Token == "" || input.Otp == "" {
			return nil, errors.New("token and otp are required")
		}

		t, err := core.ParseJWT(input.Token)
		if err != nil {
			return nil, err
		}
		phone, ok := t.Get("phone")
		if !ok {
			return nil, errors.New("cannot get phone from token")
		}
		attempts, ok := t.Get("attempts")
		if !ok {
			return nil, errors.New("cannot get attempts from token")
		}
		if err := core.InvalidateJWT(t); err != nil {
			return nil, err
		}

		if int(attempts.(float64)) >= p.conf.MaxAttempts {
			return nil, errors.New("too much attempts")
		}

		var (
			encOtp  []byte
			decrOtp string
		)
		ok, err = p.pluginApi.GetFromService(phone.(string), &encOtp)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, errors.New("otp has expired")
		}
		err = core.Decrypt(encOtp, &decrOtp)
		if err != nil {
			return nil, err
		}

		if decrOtp == input.Otp {
			return &plugins.AuthNResult{
				Cred: &plugins.Credential{
					Name:  plugins.Phone,
					Value: phone.(string),
				},
				Identity: &plugins.Identity{
					Email:         phone.(*string),
					PhoneVerified: true,
				},
				Provider: adapterName,
			}, nil
		} else {
			token, err := core.CreateJWT(
				map[string]interface{}{
					"phone":    phone,
					"attempts": int(attempts.(float64)) + 1,
				},
				p.conf.Otp.Exp)
			if err != nil {
				return nil, err
			}
			return &plugins.AuthNResult{
				Additional: map[string]interface{}{"token": token},
			}, errors.New("wrong otp")
		}
	}
}

func initConfig(rawConf *configs.RawConfig) (*config, error) {
	adapterConf := &config{}
	if err := mapstructure.Decode(rawConf, adapterConf); err != nil {
		return nil, err
	}
	adapterConf.setDefaults()

	return adapterConf, nil
}

func createRoutes(p *phone) {
	routes := []*core.Route{
		{
			Method:  http.MethodGet,
			Path:    sendUrl,
			Handler: sendOTP(p),
		},
		{
			Method:  http.MethodGet,
			Path:    resendUrl,
			Handler: resendOTP(p),
		},
	}
	p.pluginApi.AddAppRoutes(p.appName, routes)
}
