package authenticator

import (
	"aureole/internal/configs"
	"aureole/internal/core"
	"aureole/pkg/dgoogauth"
	_ "embed"
	"errors"
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3gen"
	"net/http"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/mitchellh/mapstructure"
)

//go:embed meta.yaml
var rawMeta []byte

var meta core.Metadata

// init initializes package by register pluginCreator
func init() {
	meta = core.MFARepo.Register(rawMeta, Create)
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

	Init2FAReqBody struct {
		token
	}

	VerifyReqBody struct {
		otp
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

	return nil
}

func (g otpAuth) GetMetadata() core.Metadata {
	return meta
}

// func (g otpAuth) GetCustomAppRoutes() *openapi3.Paths {
// 	return g.swagger.Paths
// }

func (g *otpAuth) IsEnabled(cred *core.Credential) (bool, error) {
	// TODO: Что вместо ID?
	return g.pluginAPI.Is2FAEnabled(cred, fmt.Sprintf("%d", meta.PluginID))
}

func (g *otpAuth) Init2FA() core.MFAInitFunc {
	return func(c fiber.Ctx) (fiber.Map, error) {
		strToken := &Init2FAReqBody{}
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

func (g *otpAuth) GetOAS3AuthRequestBody() *openapi3.RequestBody {
	schema, _ := openapi3gen.NewSchemaRefForValue(&Init2FAReqBody{}, nil)
	return &openapi3.RequestBody{
		Description: "Token",
		Required:    true,
		Content: map[string]*openapi3.MediaType{
			fiber.MIMEApplicationJSON: {
				Schema: schema,
			},
		},
	}
}

func (g *otpAuth) GetOAS3AuthParameters() openapi3.Parameters {
	return openapi3.Parameters{}
}

func (g *otpAuth) Verify() core.MFAVerifyFunc {
	return func(c fiber.Ctx) (*core.Credential, fiber.Map, error) {
		otp := &VerifyReqBody{}
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

		manager, ok := g.pluginAPI.GetIDManager()
		if !ok {
			return nil, nil, errors.New("cannot get IDManager")
		}
		secret, err := manager.GetData(cred, provider, "secret")
		if err != nil {
			return nil, nil, err
		}
		scratchCodes, err := manager.GetData(cred, provider, "scratch_codes")
		if err != nil {
			return nil, nil, err
		}

		var counter int
		if g.conf.Alg == "hotp" {
			rawCounter, err := manager.GetData(cred, provider, "counter")
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
		err = manager.On2FA(cred, &core.MFAData{
			PluginID: fmt.Sprintf("%d", meta.PluginID),
			// todo (Talgat) : get provider name
			ProviderName: "name",
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

func (g *otpAuth) GetCustomAppRoutes() []*core.Route {
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
