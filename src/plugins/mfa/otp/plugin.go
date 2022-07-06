package authenticator

import (
	"aureole/internal/configs"
	"aureole/internal/core"
	"aureole/pkg/dgoogauth"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/mitchellh/mapstructure"
)

// name is the internal name of the plugin
const name = "otp"
const ID = "1799"

// init initializes package by register plugin
func init() {
	core.Repo.Register([]byte(name), Create)
}

type (
	otpAuth struct {
		pluginAPI core.PluginAPI
		rawConf   configs.PluginConfig
		conf      *config
	}

	token struct {
		Token string `json:"token"`
	}

	otp struct {
		token
		Otp string `json:"otp"`
	}
)

func Create(conf configs.PluginConfig) core.MFA {
	return &otpAuth{rawConf: conf}
}

func (g *otpAuth) Init(api core.PluginAPI) (err error) {
	g.pluginAPI = api
	g.conf, err = initConfig(&g.rawConf.Config)
	if err != nil {
		return err
	}

	_, ok := g.pluginAPI.GetIDManager()
	if !ok {
		return fmt.Errorf("manager for app '%s' is not declared", g.pluginAPI.GetAppName())
	}

	err = json.Unmarshal(swaggerJson, &g.swagger)
	if err != nil {
		fmt.Printf("google-auth 2fa: cannot marshal swagger docs: %v", err)
	}

	createRoutes(g)
	return nil
}

func (g otpAuth) GetMetaData() core.Meta {
	return core.Meta{
		Type: name,
		Name: g.rawConf.Name,
	}
}

// func (g otpAuth) GetPaths() *openapi3.Paths {
// 	return g.swagger.Paths
// }

func (g *otpAuth) IsEnabled(cred *core.Credential) (bool, error) {
	return g.pluginAPI.Is2FAEnabled(cred, ID)
}

func (g *otpAuth) Init2FA() core.MFAInitFunc {
	return func(c fiber.Ctx) (fiber.Map, error) {
		var strToken *token
		if err := c.BodyParser(strToken); err != nil {
			return nil, err
		}
		if strToken.Token == "" {
			return nil, errors.New("token are required")
		}

		token, err := g.pluginAPI.ParseJWT(strToken.Token)
		if err != nil {
			return nil, err
		}
		_, ok := token.Get("provider")
		if !ok {
			return nil, errors.New("cannot get provider from token")
		}
		_, ok = token.Get("credential")
		if !ok {
			return nil, errors.New("cannot get credential from token")
		}

		return fiber.Map{"token": strToken.Token}, nil
	}
}

func (g *otpAuth) Verify() core.MFAVerifyFunc {
	return func(c fiber.Ctx) (*core.Credential, fiber.Map, error) {
		var otp *otp
		if err := c.BodyParser(otp); err != nil {
			return nil, nil, err
		}
		if otp.Token == "" || otp.Otp == "" {
			return nil, nil, errors.New("token and otp are required")
		}

		t, err := g.pluginAPI.ParseJWT(otp.Token)
		if err != nil {
			return nil, nil, err
		}
		rawProvider, ok := t.Get("provider")
		if !ok {
			return nil, nil, errors.New("cannot get provider from token")
		}
		rawCred, ok := t.Get("credential")
		if !ok {
			return nil, nil, errors.New("cannot get credential from token")
		}
		if err := g.pluginAPI.InvalidateJWT(t); err != nil {
			return nil, nil, err
		}

		provider := rawProvider.(string)
		cred := &core.Credential{}
		if err := mapstructure.Decode(rawCred, cred); err != nil {
			return nil, nil, err
		}

		secret, err := g.manager.GetData(cred, provider, "secret")
		if err != nil {
			return nil, nil, err
		}
		scratchCodes, err := g.manager.GetData(cred, provider, "scratch_codes")
		if err != nil {
			return nil, nil, err
		}

		var counter int
		if g.conf.Alg == "hotp" {
			rawCounter, err := g.manager.GetData(cred, provider, "counter")
			if err != nil {
				return nil, nil, err
			}
			counter = rawCounter.(int)
		}

		var usedOtp []int
		_, err = g.pluginAPI.GetFromService(cred.Value, &usedOtp)
		if err != nil {
			return nil, nil, err
		}

		otpConf := &dgoogauth.OTPConfig{
			Secret:        secret.(string),
			WindowSize:    g.conf.WindowSize,
			HotpCounter:   counter,
			DisallowReuse: usedOtp,
			ScratchCodes:  scratchCodes.([]string),
		}
		ok, err = otpConf.Authenticate(strings.TrimSpace(otp.Otp))
		if err != nil {
			return nil, nil, err
		}
		if !ok {
			return nil, nil, errors.New("wrong otp")
		}
		err = g.manager.On2FA(cred, &core.MFAData{
			PluginID:     ID,
			ProviderName: name,
			Payload:      map[string]interface{}{"counter": otpConf.HotpCounter, "scratch_code": otpConf.ScratchCodes},
		})
		if err != nil {
			return nil, nil, err
		}

		if g.conf.DisallowReuse {
			if usedOtp == nil {
				usedOtp = make([]int, 1)
			}
			intOtp, err := strconv.Atoi(otp.Otp)
			if err != nil {
				return nil, nil, err
			}

			usedOtp = append(usedOtp, intOtp)
			if err := g.pluginAPI.SaveToService(cred.Value, usedOtp, 1); err != nil {
				return nil, nil, err
			}
		}

		return cred, nil, nil
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

func (g *otpAuth) GetPaths() []*core.Route {
	return []*core.Route{
		{
			Method:  http.MethodPost,
			Path:    getQRUrl,
			Handler: getQR(g),
		},
		{
			Method:  http.MethodPost,
			Path:    getScratchesUrl,
			Handler: getScratchCodes(g),
		},
	}
}
