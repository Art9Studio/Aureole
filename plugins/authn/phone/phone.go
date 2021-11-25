package phone

import (
	"aureole/internal/configs"
	"aureole/internal/identity"
	authzTypes "aureole/internal/plugins/authz/types"
	"aureole/internal/plugins/core"
	"aureole/internal/plugins/pwhasher/types"
	senderTypes "aureole/internal/plugins/sender/types"
	"aureole/internal/router/interface"
	app "aureole/internal/state/interface"
	"fmt"
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

func (*phone) GetPluginID() string {
	return PluginID
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
			Path:    p.conf.PathPrefix + p.conf.SendUrl,
			Handler: SendOtp(p),
		},
		{
			Method:  "POST",
			Path:    p.conf.PathPrefix + p.conf.ConfirmUrl,
			Handler: Login(p),
		},
		{
			Method:  "POST",
			Path:    p.conf.PathPrefix + p.conf.ResendUrl,
			Handler: Resend(p),
		},
	}
	p.pluginApi.GetRouter().AddAppRoutes(p.app.GetName(), routes)
}
