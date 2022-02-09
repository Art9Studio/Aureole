package phone

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

const pluginID = "6937"

type (
	authn struct {
		pluginAPI     core.PluginAPI
		rawConf       *configs.Authn
		conf          *config
		sender        plugins.Sender
		tmpl, tmplExt string
		swagger       struct {
			Paths       *spec.Paths
			Definitions spec.Definitions
		}
	}

	phone struct {
		Phone string `json:"phone"`
	}

	otp struct {
		Token string `json:"token"`
		Otp   string `json:"otp"`
	}
)

//go:embed docs/swagger.json
var swaggerJson []byte

func (a *authn) Init(api core.PluginAPI) (err error) {
	a.pluginAPI = api
	a.conf, err = initConfig(&a.rawConf.Config)
	if err != nil {
		return err
	}

	a.sender, err = a.pluginAPI.GetSender(a.conf.Sender)
	if err != nil {
		return err
	}

	err = json.Unmarshal(swaggerJson, &a.swagger)
	if err != nil {
		fmt.Printf("phone authn: cannot marshal swagger docs: %v", err)
	}

	tmpl, err := os.ReadFile(a.conf.TmplPath)
	if err != nil {
		a.tmpl = defaultTmpl
		a.tmplExt = "txt"
	} else {
		a.tmpl = string(tmpl)
		a.tmplExt = path.Ext(a.conf.TmplPath)
	}

	createRoutes(a)
	return nil
}

func (*authn) GetMetaData() plugins.Meta {
	return plugins.Meta{
		Type: adapterName,
		ID:   pluginID,
	}
}

func (a *authn) GetHandlersSpec() (*spec.Paths, spec.Definitions) {
	return a.swagger.Paths, a.swagger.Definitions
}

func (a *authn) LoginWrapper() plugins.AuthNLoginFunc {
	return func(c fiber.Ctx) (*plugins.AuthNResult, error) {
		var otp otp
		if err := c.BodyParser(&otp); err != nil {
			return nil, err
		}
		if otp.Token == "" || otp.Otp == "" {
			return nil, errors.New("token and otp are required")
		}

		var (
			phone    string
			attempts int
		)
		t, err := a.pluginAPI.ParseJWT(otp.Token)
		if err != nil {
			return nil, err
		}
		err = a.pluginAPI.GetFromJWT(t, "phone", &phone)
		if err != nil {
			return nil, errors.New("cannot get phone from token")
		}
		err = a.pluginAPI.GetFromJWT(t, "attempts", &attempts)
		if err != nil {
			return nil, errors.New("cannot get attempts count from token")
		}
		if err := a.pluginAPI.InvalidateJWT(t); err != nil {
			return nil, err
		}

		if attempts >= a.conf.MaxAttempts {
			return nil, errors.New("too much attempts")
		}

		var (
			encOtp  []byte
			decrOtp string
		)
		ok, err := a.pluginAPI.GetFromService(phone, &encOtp)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, errors.New("otp has expired")
		}

		err = a.pluginAPI.Decrypt(encOtp, &decrOtp)
		if err != nil {
			return nil, err
		}

		if decrOtp == otp.Otp {
			return &plugins.AuthNResult{
				Cred: &plugins.Credential{
					Name:  plugins.Phone,
					Value: phone,
				},
				Identity: &plugins.Identity{
					Phone:         &phone,
					PhoneVerified: true,
				},
				Provider: adapterName,
			}, nil
		} else {
			token, err := a.pluginAPI.CreateJWT(
				map[string]interface{}{
					"phone":    phone,
					"attempts": attempts + 1,
				},
				a.conf.Otp.Exp)
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

func createRoutes(p *authn) {
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
