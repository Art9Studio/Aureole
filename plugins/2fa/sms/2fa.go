package sms

import (
	"aureole/internal/configs"
	"aureole/internal/core"
	"aureole/internal/plugins"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-openapi/spec"
	"net/http"
	"os"
	"path"

	"github.com/gofiber/fiber/v2"
	"github.com/mitchellh/mapstructure"
)

const pluginID = "0509"

type (
	mfa struct {
		pluginAPI     core.PluginAPI
		rawConf       *configs.SecondFactor
		conf          *config
		sender        plugins.Sender
		tmpl, tmplExt string
		swagger       struct {
			Paths       *spec.Paths
			Definitions spec.Definitions
		}
	}

	token struct {
		Token string `json:"token"`
	}

	otp struct {
		token
		Otp string `json:"otp"`
	}
)

//go:embed docs/swagger.json
var swaggerJson []byte

func (m *mfa) Init(api core.PluginAPI) (err error) {
	m.pluginAPI = api
	m.conf, err = initConfig(&m.rawConf.Config)
	if err != nil {
		return err
	}

	_, err = m.pluginAPI.GetIDManager()
	if err != nil {
		return err
	}

	err = json.Unmarshal(swaggerJson, &m.swagger)
	if err != nil {
		fmt.Printf("sms 2fa: cannot marshal swagger docs: %v", err)
	}

	tmpl, err := os.ReadFile(m.conf.TmplPath)
	if err != nil {
		m.tmpl = defaultTmpl
		m.tmplExt = "txt"
	} else {
		m.tmpl = string(tmpl)
		m.tmplExt = path.Ext(m.conf.TmplPath)
	}

	createRoutes(m)
	return nil
}

func (m *mfa) GetMetaData() plugins.Meta {
	return plugins.Meta{
		Type: adapterName,
		Name: m.rawConf.Name,
		ID:   pluginID,
	}
}

func (m *mfa) GetHandlersSpec() (*spec.Paths, spec.Definitions) {
	return m.swagger.Paths, m.swagger.Definitions
}

func (m *mfa) IsEnabled(cred *plugins.Credential) (bool, error) {
	return m.pluginAPI.Is2FAEnabled(cred, pluginID)
}

func (m *mfa) Init2FA() plugins.MFAInitFunc {
	return func(c fiber.Ctx) (fiber.Map, error) {
		var strToken token
		if err := c.BodyParser(&strToken); err != nil {
			return nil, err
		}
		if strToken.Token == "" {
			return nil, errors.New("token are required")
		}

		var (
			provider string
			cred     plugins.Credential
		)
		t, err := m.pluginAPI.ParseJWT(strToken.Token)
		if err != nil {
			return nil, err
		}
		err = m.pluginAPI.GetFromJWT(t, "provider", &provider)
		if err != nil {
			return nil, errors.New("cannot get provider from token")
		}
		err = m.pluginAPI.GetFromJWT(t, "credential", &cred)
		if err != nil {
			return nil, errors.New("cannot get credential from token")
		}

		otp, err := m.pluginAPI.GetRandStr(m.conf.Otp.Length, m.conf.Otp.Alphabet)
		if err != nil {
			return nil, err
		}

		token, err := m.pluginAPI.CreateJWT(
			map[string]interface{}{
				"phone":    cred.Value,
				"provider": provider,
				"attempts": 0,
			},
			m.conf.Otp.Exp)
		if err != nil {
			return nil, err
		}

		encOtp, err := m.pluginAPI.Encrypt(otp)
		if err != nil {
			return nil, err
		}
		err = m.pluginAPI.SaveToService(cred.Value, encOtp, m.conf.Otp.Exp)
		if err != nil {
			return nil, err
		}

		err = m.sender.Send(cred.Value, "", m.tmpl, m.tmplExt, map[string]interface{}{"otp": otp})
		if err != nil {
			return nil, err
		}

		return fiber.Map{"token": token}, nil
	}
}

func (m *mfa) Verify() plugins.MFAVerifyFunc {
	return func(c fiber.Ctx) (*plugins.Credential, fiber.Map, error) {
		var otp otp
		if err := c.BodyParser(&otp); err != nil {
			return nil, nil, err
		}
		if otp.Token == "" || otp.Otp == "" {
			return nil, nil, errors.New("token and otp are required")
		}

		t, err := m.pluginAPI.ParseJWT(otp.Token)
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
		if err := m.pluginAPI.InvalidateJWT(t); err != nil {
			return nil, nil, err
		}

		if int(attempts.(float64)) >= m.conf.MaxAttempts {
			return nil, nil, errors.New("too much attempts")
		}

		var (
			encOtp  []byte
			decrOtp string
		)
		ok, err = m.pluginAPI.GetFromService(phone.(string), &encOtp)
		if err != nil {
			return nil, nil, err
		}
		if !ok {
			return nil, nil, errors.New("otp has expired")
		}
		err = m.pluginAPI.Decrypt(encOtp, &decrOtp)
		if err != nil {
			return nil, nil, err
		}

		if decrOtp == otp.Otp {
			return &plugins.Credential{
				Name:  "phone",
				Value: phone.(string),
			}, nil, nil
		} else {
			token, err := m.pluginAPI.CreateJWT(
				map[string]interface{}{
					"phone":    phone,
					"attempts": int(attempts.(float64)) + 1,
				},
				m.conf.Otp.Exp)
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

func createRoutes(m *mfa) {
	routes := []*core.Route{
		{
			Method:  http.MethodPost,
			Path:    resendUrl,
			Handler: resend(m),
		},
	}
	m.pluginAPI.AddAppRoutes(routes)
}
