package sms

import (
	"aureole/internal/configs"
	"aureole/internal/core"
	"aureole/internal/plugins"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/gofiber/fiber/v2"
	"github.com/mitchellh/mapstructure"
)

const pluginID = "0509"

type (
	sms struct {
		pluginAPI     core.PluginAPI
		rawConf       *configs.SecondFactor
		conf          *config
		sender        plugins.Sender
		tmpl, tmplExt string
	}

	input struct {
		Phone string `json:"phone"`
		Token string `json:"token"`
		Otp   string `json:"otp"`
	}
)

func (s *sms) Init(api core.PluginAPI) (err error) {
	s.pluginAPI = api
	s.conf, err = initConfig(&s.rawConf.Config)
	if err != nil {
		return err
	}

	_, ok := s.pluginAPI.GetIDManager()
	if !ok {
		return fmt.Errorf("manager for app '%s' is not declared", s.pluginAPI.GetAppName())
	}

	tmpl, err := os.ReadFile(s.conf.TmplPath)
	if err != nil {
		s.tmpl = defaultTmpl
		s.tmplExt = "txt"
	} else {
		s.tmpl = string(tmpl)
		s.tmplExt = path.Ext(s.conf.TmplPath)
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

func (s *sms) Clone() interface{} {
	return &sms{rawConf: s.rawConf}
}

func (s *sms) IsEnabled(cred *plugins.Credential) (bool, error) {
	return s.pluginAPI.Is2FAEnabled(cred, pluginID)
}

func (s *sms) Init2FA() plugins.MFAInitFunc {
	return func(c fiber.Ctx) (fiber.Map, error) {
		var input *input
		if err := c.BodyParser(input); err != nil {
			return nil, err
		}
		if input.Token == "" {
			return nil, errors.New("token are required")
		}

		var (
			provider string
			cred     plugins.Credential
		)
		t, err := s.pluginAPI.ParseJWT(input.Token)
		if err != nil {
			return nil, err
		}
		err = s.pluginAPI.GetFromJWT(t, "provider", &provider)
		if err != nil {
			return nil, errors.New("cannot get provider from token")
		}
		err = s.pluginAPI.GetFromJWT(t, "credential", &cred)
		if err != nil {
			return nil, errors.New("cannot get credential from token")
		}

		otp, err := s.pluginAPI.GetRandStr(s.conf.Otp.Length, s.conf.Otp.Alphabet)
		if err != nil {
			return nil, err
		}

		token, err := s.pluginAPI.CreateJWT(
			map[string]interface{}{
				"phone":    cred.Value,
				"provider": provider,
				"attempts": 0,
			},
			s.conf.Otp.Exp)
		if err != nil {
			return nil, err
		}

		encOtp, err := s.pluginAPI.Encrypt(otp)
		if err != nil {
			return nil, err
		}
		err = s.pluginAPI.SaveToService(cred.Value, encOtp, s.conf.Otp.Exp)
		if err != nil {
			return nil, err
		}

		err = s.sender.Send(cred.Value, "", s.tmpl, s.tmplExt, map[string]interface{}{"otp": otp})
		if err != nil {
			return nil, err
		}

		return fiber.Map{"token": token}, nil
	}
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

		t, err := s.pluginAPI.ParseJWT(input.Token)
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
		if err := s.pluginAPI.InvalidateJWT(t); err != nil {
			return nil, nil, err
		}

		if int(attempts.(float64)) >= s.conf.MaxAttempts {
			return nil, nil, errors.New("too much attempts")
		}

		var (
			encOtp  []byte
			decrOtp string
		)
		ok, err = s.pluginAPI.GetFromService(phone.(string), &encOtp)
		if err != nil {
			return nil, nil, err
		}
		if !ok {
			return nil, nil, errors.New("otp has expired")
		}
		err = s.pluginAPI.Decrypt(encOtp, &decrOtp)
		if err != nil {
			return nil, nil, err
		}

		if decrOtp == input.Otp {
			return &plugins.Credential{
				Name:  "phone",
				Value: phone.(string),
			}, nil, nil
		} else {
			token, err := s.pluginAPI.CreateJWT(
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
	s.pluginAPI.AddAppRoutes(routes)
}
