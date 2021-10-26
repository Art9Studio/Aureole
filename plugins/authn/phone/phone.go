package phone

import (
	"aureole/internal/configs"
	"aureole/internal/identity"
	"aureole/internal/plugins/authn"
	authzTypes "aureole/internal/plugins/authz/types"
	cKeyTypes "aureole/internal/plugins/cryptokey/types"
	"aureole/internal/plugins/pwhasher/types"
	senderTypes "aureole/internal/plugins/sender/types"
	"aureole/internal/router/interface"
	app "aureole/internal/state/interface"
	"errors"
	"fmt"
	"github.com/mitchellh/mapstructure"
)

type (
	phone struct {
		app        app.AppState
		rawConf    *configs.Authn
		conf       *config
		manager    identity.ManagerI
		hasher     types.PwHasher
		authorizer authzTypes.Authorizer
		serviceKey cKeyTypes.CryptoKey
		sender     senderTypes.Sender
	}

	input struct {
		Phone string `json:"phone"`
		Token string `json:"token"`
		Otp   string `json:"otp"`
	}
)

func (p *phone) Init(app app.AppState) (err error) {
	p.app = app
	p.rawConf.PathPrefix = "/" + AdapterName

	p.conf, err = initConfig(&p.rawConf.Config)
	if err != nil {
		return err
	}

	pluginApi := authn.Repository.PluginApi
	p.manager, err = app.GetIdentityManager()
	if err != nil {
		fmt.Printf("manager for app '%s' is not declared, persis layer is not available", app.GetName())
	}

	p.hasher, err = pluginApi.Project.GetHasher(p.conf.Hasher)
	if err != nil {
		return fmt.Errorf("hasher named '%s' is not declared", p.conf.Hasher)
	}

	p.serviceKey, err = pluginApi.Project.GetCryptoKey("service_internal_key")
	if err != nil {
		return errors.New("cryptokey named 'service_internal_key' is not declared")
	}

	p.sender, err = pluginApi.Project.GetSender(p.conf.Sender)
	if err != nil {
		return fmt.Errorf("sender named '%s' is not declared", p.conf.Sender)
	}

	p.authorizer, err = p.app.GetAuthorizer()
	if err != nil {
		return fmt.Errorf("authorizer named for app '%s' is not declared", app.GetName())
	}

	createRoutes(p)
	return nil
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
	routes := []*_interface.Route{
		{
			Method:  "POST",
			Path:    p.rawConf.PathPrefix + p.conf.SendUrl,
			Handler: SendOtp(p),
		},
		{
			Method:  "POST",
			Path:    p.rawConf.PathPrefix + p.conf.ConfirmUrl,
			Handler: Login(p),
		},
		{
			Method:  "POST",
			Path:    p.rawConf.PathPrefix + p.conf.ResendUrl,
			Handler: Resend(p),
		},
	}
	authn.Repository.PluginApi.Router.AddAppRoutes(p.app.GetName(), routes)
}
