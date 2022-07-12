package phone

import (
	"aureole/internal/configs"
	"aureole/internal/core"
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3gen"
	"net/http"
	"os"
	"path"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/mitchellh/mapstructure"

	_ "embed"
	"errors"
)

//go:embed meta.yaml
var rawMeta []byte

var meta core.Metadata

// init initializes package by register pluginCreator
func init() {
	meta = core.AuthenticatorRepo.Register(rawMeta, Create)
}

type (
	authn struct {
		pluginAPI     core.PluginAPI
		rawConf       configs.AuthPluginConfig
		conf          *config
		sender        core.Sender
		tmpl, tmplExt string
	}

	phone struct {
		Phone string `json:"phone"`
	}

	otp struct {
		Token string `json:"token"`
		Otp   string `json:"otp"`
	}

	token struct {
		Token string `json:"token"`
	}
)

func (p *authn) GetAuthHTTPMethod() string {
	return http.MethodGet
}

func Create(conf configs.AuthPluginConfig) core.Authenticator {
	return &authn{rawConf: conf}
}

func (a *authn) Init(api core.PluginAPI) (err error) {
	a.pluginAPI = api
	a.conf, err = initConfig(&a.rawConf.Config)
	if err != nil {
		return err
	}

	var ok bool
	a.sender, ok = a.pluginAPI.GetSender(a.conf.Sender)
	if !ok {
		return fmt.Errorf("sender named '%s' is not declared", a.conf.Sender)
	}

	tmpl, err := os.ReadFile(a.conf.TmplPath)
	if err != nil {
		a.tmpl = defaultTmpl
		a.tmplExt = "txt"
	} else {
		a.tmpl = string(tmpl)
		a.tmplExt = path.Ext(a.conf.TmplPath)
	}

	return nil
}

func (authn) GetMetadata() core.Metadata {
	return meta
}

func (a *authn) GetAuthHandler() core.AuthHandlerFunc {
	return func(c fiber.Ctx) (*core.AuthResult, error) {
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
			return &core.AuthResult{
				Cred: &core.Credential{
					Name:  core.Phone,
					Value: phone,
				},
				Identity: &core.Identity{
					Email:         &phone,
					PhoneVerified: true,
				},
				Provider: meta.ShortName,
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
			return &core.AuthResult{ErrorData: fiber.Map{"token": token}}, errors.New("wrong otp")
		}
	}
}

func initConfig(conf *configs.RawConfig) (*config, error) {
	PluginConf := &config{}
	if err := mapstructure.Decode(conf, PluginConf); err != nil {
		return nil, err
	}
	PluginConf.setDefaults()

	return PluginConf, nil
}

func (a *authn) GetCustomAppRoutes() []*core.Route {
	phoneSchema, _ := openapi3gen.NewSchemaRefForValue(phone{}, nil)
	otpSchema, _ := openapi3gen.NewSchemaRefForValue(otp{}, nil)
	resSchema, _ := openapi3gen.NewSchemaRefForValue(token{}, nil)
	return []*core.Route{
		{
			Method:  http.MethodPost,
			Path:    sendUrl,
			Handler: sendOTP(a),
			OAS3Operation: core.NewOA3Operation(meta, phoneSchema, nil, map[string]*openapi3.SchemaRef{
				strconv.Itoa(http.StatusOK):                  resSchema,
				strconv.Itoa(http.StatusBadRequest):          nil,
				strconv.Itoa(http.StatusInternalServerError): nil,
			}),
		},
		{
			Method:  http.MethodPost,
			Path:    resendUrl,
			Handler: resendOTP(a),
			OAS3Operation: core.NewOA3Operation(meta, otpSchema, nil, map[string]*openapi3.SchemaRef{
				strconv.Itoa(http.StatusOK):                  resSchema,
				strconv.Itoa(http.StatusBadRequest):          nil,
				strconv.Itoa(http.StatusInternalServerError): nil,
			}),
		},
	}
}
