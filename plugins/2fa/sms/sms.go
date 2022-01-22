package sms

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

const pluginID = "0509"

type (
	sms struct {
		pluginApi core.PluginAPI
		app       *core.App
		rawConf   *configs.SecondFactor
		conf      *config
		sender    plugins.Sender
	}

	input struct {
		Phone string `json:"phone"`
		Token string `json:"token"`
		Otp   string `json:"otp"`
	}
)

func (s *sms) Init(appName string, api core.PluginAPI) (err error) {
	s.pluginApi = api
	s.conf, err = initConfig(&s.rawConf.Config)
	if err != nil {
		return err
	}

	s.app, err = s.pluginApi.GetApp(appName)
	if err != nil {
		return fmt.Errorf("app named '%s' is not declared", appName)
	}

	createRoutes(s)
	return nil
}

func (s *sms) GetMetaData() plugins.Meta {
	return plugins.Meta{
		Type: adapterName,
		Name: s.rawConf.Name,
		ID:   pluginID,
	}
}

func (s *sms) IsEnabled(cred *plugins.Credential, provider string) (bool, error) {
	enabled, id, err := s.pluginApi.Is2FAEnabled(cred, provider)
	if err != nil {
		return false, err
	}
	if !enabled {
		return false, nil
	}
	if id != pluginID {
		return false, errors.New("another 2FA is enabled")
	}
	return true, nil
}

func (s *sms) Init2FA(cred *plugins.Credential, provider string, _ fiber.Ctx) (fiber.Map, error) {
	otp, err := core.GetRandStr(s.conf.Otp.Length, s.conf.Otp.Alphabet)
	if err != nil {
		return nil, err
	}

	token, err := core.CreateJWT(
		map[string]interface{}{
			"phone":    cred.Value,
			"provider": provider,
			"attempts": 0,
		},
		s.conf.Otp.Exp)
	if err != nil {
		return nil, err
	}

	encOtp, err := core.Encrypt(otp)
	if err != nil {
		return nil, err
	}
	err = s.pluginApi.SaveToService(cred.Value, encOtp, s.conf.Otp.Exp)
	if err != nil {
		return nil, err
	}

	err = s.sender.Send(cred.Value, "", s.conf.Template, map[string]interface{}{"otp": otp})
	if err != nil {
		return nil, err
	}

	return fiber.Map{"token": token}, nil
}

func (s *sms) Verify() plugins.MFAVerifyFunc {
	return func(c fiber.Ctx) (*plugins.Credential, fiber.Map, error) {
		var input *input
		if err := c.BodyParser(input); err != nil {
			return nil, nil, err
		}
		if input.Token == "" || input.Otp == "" {
			return nil, nil, errors.New("token and otp are required")
		}

		t, err := core.ParseJWT(input.Token)
		if err != nil {
			return nil, nil, err
		}
		phone, ok := t.Get("phone")
		if !ok {
			return nil, nil, errors.New("cannot get otp from token")
		}
		attempts, ok := t.Get("attempts")
		if !ok {
			return nil, nil, errors.New("cannot get attempts from token")
		}
		if err := core.InvalidateJWT(t); err != nil {
			return nil, nil, err
		}

		if int(attempts.(float64)) >= s.conf.MaxAttempts {
			return nil, nil, errors.New("too much attempts")
		}

		var (
			encOtp  []byte
			decrOtp string
		)
		ok, err = s.pluginApi.GetFromService(phone.(string), &encOtp)
		if err != nil {
			return nil, nil, err
		}
		if !ok {
			return nil, nil, errors.New("otp has expired")
		}
		err = core.Decrypt(encOtp, &decrOtp)
		if err != nil {
			return nil, nil, err
		}

		if decrOtp == input.Otp {
			return &plugins.Credential{
				Name:  "phone",
				Value: phone.(string),
			}, nil, nil
		} else {
			token, err := core.CreateJWT(
				map[string]interface{}{
					"phone":    phone,
					"attempts": int(attempts.(float64)) + 1,
				},
				s.conf.Otp.Exp)
			if err != nil {
				return nil, nil, err
			}
			return nil, fiber.Map{"token": token}, errors.New("wrong otp")
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

func createRoutes(s *sms) {
	routes := []*core.Route{
		{
			Method:  http.MethodPost,
			Path:    resendUrl,
			Handler: resend(s),
		},
	}
	s.pluginApi.AddAppRoutes(s.app.GetName(), routes)
}
