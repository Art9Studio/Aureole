package authenticator

import (
	"aureole/internal/configs"
	"aureole/internal/core"
	"aureole/internal/plugins"
	"aureole/pkg/dgoogauth"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-openapi/spec"
	"net/http"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/mitchellh/mapstructure"
)

const pluginID = "1799"

type (
	gauth struct {
		pluginAPI core.PluginAPI
		rawConf   *configs.SecondFactor
		conf      *config
		manager   plugins.IDManager
		swagger   struct {
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

//go:embed swagger.json
var swaggerJson []byte

func (g *gauth) Init(api core.PluginAPI) (err error) {
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

func (g *gauth) GetMetaData() plugins.Meta {
	return plugins.Meta{
		Type: adapterName,
		Name: g.rawConf.Name,
		ID:   pluginID,
	}
}

func (g *gauth) GetHandlersSpec() (*spec.Paths, spec.Definitions) {
	return g.swagger.Paths, g.swagger.Definitions
}

func (g *gauth) IsEnabled(cred *plugins.Credential) (bool, error) {
	return g.pluginAPI.Is2FAEnabled(cred, pluginID)
}

func (g *gauth) Init2FA() plugins.MFAInitFunc {
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

func (g *gauth) Verify() plugins.MFAVerifyFunc {
	return func(c fiber.Ctx) (*plugins.Credential, fiber.Map, error) {
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
		cred := &plugins.Credential{}
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
		err = g.manager.On2FA(cred, &plugins.MFAData{
			PluginID:     pluginID,
			ProviderName: adapterName,
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

func initConfig(rawConf *configs.RawConfig) (*config, error) {
	adapterConf := &config{}
	if err := mapstructure.Decode(rawConf, adapterConf); err != nil {
		return nil, err
	}
	adapterConf.setDefaults()
	return adapterConf, nil
}

func createRoutes(g *gauth) {
	routes := []*core.Route{
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
	g.pluginAPI.AddAppRoutes(routes)
}
