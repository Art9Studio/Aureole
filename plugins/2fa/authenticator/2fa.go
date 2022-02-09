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
	mfa struct {
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
		Token string `json:"token"`
		Otp   string `json:"otp"`
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

	m.manager, err = m.pluginAPI.GetIDManager()
	if err != nil {
		return err
	}

	err = json.Unmarshal(swaggerJson, &m.swagger)
	if err != nil {
		fmt.Printf("google-auth 2fa: cannot marshal swagger docs: %v", err)
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

		token, err := m.pluginAPI.ParseJWT(strToken.Token)
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
		rawCred, ok := t.Get("credential")
		if !ok {
			return nil, nil, errors.New("cannot get credential from token")
		}
		if err := m.pluginAPI.InvalidateJWT(t); err != nil {
			return nil, nil, err
		}

		cred := &plugins.Credential{}
		if err := mapstructure.Decode(rawCred, cred); err != nil {
			return nil, nil, err
		}

		mfaData, err := m.manager.Get2FAData(cred, pluginID)
		if err != nil {
			return nil, nil, err
		}

		secret := mfaData.Payload["secret"]
		rawScratchCodes := mfaData.Payload["scratch_codes"]

		var scratchCodes []string
		for _, code := range rawScratchCodes.([]interface{}) {
			scratchCodes = append(scratchCodes, code.(string))
		}

		var counter int
		if m.conf.Alg == "hotp" {
			rawCounter := mfaData.Payload["counter"]
			counter = int(rawCounter.(float64))
		}

		var usedOtp []int
		_, err = m.pluginAPI.GetFromService(cred.Value, &usedOtp)
		if err != nil {
			return nil, nil, err
		}

		otpConf := &dgoogauth.OTPConfig{
			Secret:        secret.(string),
			WindowSize:    m.conf.WindowSize,
			HotpCounter:   counter,
			DisallowReuse: usedOtp,
			ScratchCodes:  scratchCodes,
		}
		ok, err = otpConf.Authenticate(strings.TrimSpace(otp.Otp))
		if err != nil {
			return nil, nil, err
		}
		if !ok {
			return nil, nil, errors.New("wrong otp")
		}
		err = m.manager.On2FA(cred, &plugins.MFAData{
			PluginID:     pluginID,
			ProviderName: adapterName,
			Payload:      map[string]interface{}{"counter": otpConf.HotpCounter, "scratch_code": otpConf.ScratchCodes},
		})
		if err != nil {
			return nil, nil, err
		}

		if m.conf.DisallowReuse {
			if usedOtp == nil {
				usedOtp = make([]int, 1)
			}
			intOtp, err := strconv.Atoi(otp.Otp)
			if err != nil {
				return nil, nil, err
			}

			usedOtp = append(usedOtp, intOtp)
			if err := m.pluginAPI.SaveToService(cred.Value, usedOtp, 1); err != nil {
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

func createRoutes(m *mfa) {
	routes := []*core.Route{
		{
			Method:  http.MethodPost,
			Path:    getQRUrl,
			Handler: getQR(m),
		},
		{
			Method:  http.MethodPost,
			Path:    getScratchesUrl,
			Handler: getScratchCodes(m),
		},
	}
	m.pluginAPI.AddAppRoutes(routes)
}
