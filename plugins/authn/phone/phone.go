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
		pluginAPI core.PluginAPI
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

func (p *phone) Init(api core.PluginAPI) (err error) {
	p.pluginAPI = api
	p.conf, err = initConfig(&p.rawConf.Config)
	if err != nil {
		return err
	}

	var ok bool
	p.sender, ok = p.pluginAPI.GetSender(p.conf.Sender)
	if !ok {
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

func (p *phone) GetLoginHandler() (string, func() plugins.AuthNLoginFunc) {
	return http.MethodPost, p.login
}

func (p *phone) login() plugins.AuthNLoginFunc {
	return func(c fiber.Ctx) (*plugins.AuthNResult, error) {
		var input input
		if err := c.BodyParser(&input); err != nil {
			return nil, err
		}
		if input.Token == "" || input.Otp == "" {
			return nil, errors.New("token and otp are required")
		}

		var (
			phone    string
			attempts int
		)
		t, err := p.pluginAPI.ParseJWT(input.Token)
		if err != nil {
			return nil, err
		}
		err = p.pluginAPI.GetFromJWT(t, "phone", &phone)
		if err != nil {
			return nil, errors.New("cannot get phone from token")
		}
		err = p.pluginAPI.GetFromJWT(t, "attempts", &attempts)
		if err != nil {
			return nil, errors.New("cannot get attempts count from token")
		}
		if err := p.pluginAPI.InvalidateJWT(t); err != nil {
			return nil, err
		}

		if attempts >= p.conf.MaxAttempts {
			return nil, errors.New("too much attempts")
		}

		var (
			encOtp  []byte
			decrOtp string
		)
		ok, err := p.pluginAPI.GetFromService(phone, &encOtp)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, errors.New("otp has expired")
		}

		err = p.pluginAPI.Decrypt(encOtp, &decrOtp)
		if err != nil {
			return nil, err
		}

		if decrOtp == input.Otp {
			return &plugins.AuthNResult{
				Cred: &plugins.Credential{
					Name:  plugins.Phone,
					Value: phone,
				},
				Identity: &plugins.Identity{
					Email:         &phone,
					PhoneVerified: true,
				},
				Provider: adapterName,
			}, nil
		} else {
			token, err := p.pluginAPI.CreateJWT(
				map[string]interface{}{
					"phone":    phone,
					"attempts": attempts + 1,
				},
				p.conf.Otp.Exp)
			if err != nil {
				return nil, err
			}
			return &plugins.AuthNResult{ErrorData: fiber.Map{"token": token}}, errors.New("wrong otp")
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
			Method:  http.MethodPost,
			Path:    sendUrl,
			Handler: sendOTP(p),
		},
		{
			Method:  http.MethodPost,
			Path:    resendUrl,
			Handler: resendOTP(p),
		},
	}
	p.pluginAPI.AddAppRoutes(routes)
}
