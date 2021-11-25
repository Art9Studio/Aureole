package phone

import (
	"aureole/internal/configs"
	"aureole/internal/encrypt"
	"aureole/internal/identity"
	"aureole/internal/jwt"
	"aureole/internal/plugins"
	authnT "aureole/internal/plugins/authn/types"
	authzTypes "aureole/internal/plugins/authz/types"
	"aureole/internal/plugins/core"
	"aureole/internal/plugins/pwhasher/types"
	senderTypes "aureole/internal/plugins/sender/types"
	"aureole/internal/router"
	app "aureole/internal/state/interface"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/mitchellh/mapstructure"
)

const PluginID = "6937"

type (
	phone struct {
		pluginApi  core.PluginAPI
		app        app.AppState
		rawConf    *configs.Authn
		conf       *config
		manager    identity.ManagerI
		hasher     types.PwHasher
		authorizer authzTypes.Authorizer
		sender     senderTypes.Sender
	}

	input struct {
		Phone string `json:"phone"`
		Token string `json:"token"`
		Otp   string `json:"otp"`
	}
)

func (p *phone) Init(appName string, api core.PluginAPI) (err error) {
	p.pluginApi = api
	p.conf, err = initConfig(&p.rawConf.Config)
	if err != nil {
		return err
	}

	p.app, err = p.pluginApi.GetApp(appName)
	if err != nil {
		return fmt.Errorf("app named '%s' is not declared", appName)
	}

	p.manager, err = p.app.GetIdentityManager()
	if err != nil {
		fmt.Printf("manager for app '%s' is not declared", appName)
	}

	p.hasher, err = p.pluginApi.GetHasher(p.conf.Hasher)
	if err != nil {
		return fmt.Errorf("hasher named '%s' is not declared", p.conf.Hasher)
	}

	p.sender, err = p.pluginApi.GetSender(p.conf.Sender)
	if err != nil {
		return fmt.Errorf("sender named '%s' is not declared", p.conf.Sender)
	}

	p.authorizer, err = p.app.GetAuthorizer()
	if err != nil {
		return fmt.Errorf("authorizer named for app '%s' is not declared", appName)
	}

	createRoutes(p)
	return nil
}

func (*phone) GetMetaData() plugins.Meta {
	return plugins.Meta{
		Type: AdapterName,
		ID:   PluginID,
	}
}

func (p *phone) Login() authnT.AuthFunc {
	return func(c fiber.Ctx) (*identity.Credential, fiber.Map, error) {
		var input *input
		if err := c.BodyParser(input); err != nil {
			return nil, nil, err
		}
		if input.Token == "" || input.Otp == "" {
			return nil, nil, errors.New("token and otp are required")
		}

		t, err := jwt.ParseJWT(input.Token)
		if err != nil {
			return nil, nil, err
		}
		phone, ok := t.Get("phone")
		if !ok {
			return nil, nil, errors.New("cannot get phone from token")
		}
		attempts, ok := t.Get("attempts")
		if !ok {
			return nil, nil, errors.New("cannot get attempts from token")
		}
		if err := jwt.InvalidateJWT(t); err != nil {
			return nil, nil, err
		}

		if int(attempts.(float64)) >= p.conf.MaxAttempts {
			return nil, nil, errors.New("too much attempts")
		}

		var (
			encOtp  []byte
			decrOtp string
		)
		ok, err = p.pluginApi.GetFromService(phone.(string), &encOtp)
		if err != nil {
			return nil, nil, err
		}
		if !ok {
			return nil, nil, errors.New("otp has expired")
		}
		err = encrypt.Decrypt(encOtp, &decrOtp)
		if err != nil {
			return nil, nil, err
		}

		if decrOtp == input.Otp {
			return &identity.Credential{
					Name:  identity.Phone,
					Value: phone.(string),
				},
				fiber.Map{
					identity.Phone:         phone,
					identity.PhoneVerified: true,
					identity.AuthnProvider: AdapterName,
				}, nil
		} else {
			token, err := jwt.CreateJWT(
				map[string]interface{}{
					"phone":    phone,
					"attempts": int(attempts.(float64)) + 1,
				},
				p.conf.Otp.Exp)
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

func createRoutes(p *phone) {
	routes := []*router.Route{
		{
			Method:  router.MethodGET,
			Path:    p.conf.PathPrefix + p.conf.SendUrl,
			Handler: SendOtp(p),
		},
		{
			Method:  router.MethodGET,
			Path:    p.conf.PathPrefix + p.conf.ResendUrl,
			Handler: Resend(p),
		},
	}
	router.GetRouter().AddAppRoutes(p.app.GetName(), routes)
}
