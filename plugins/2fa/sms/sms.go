package sms

import (
	"aureole/internal/configs"
	"aureole/internal/plugins/core"
	senderT "aureole/internal/plugins/sender/types"
	"aureole/internal/router/interface"
	app "aureole/internal/state/interface"
	"fmt"
	"github.com/mitchellh/mapstructure"
)

const PluginID = "0509"

type (
	sms struct {
		pluginApi core.PluginAPI
		app       app.AppState
		rawConf   *configs.SecondFactor
		conf      *config
		sender    senderT.Sender
	}

	input struct {
		Phone string `json:"phone"`
		Token string `json:"token"`
		Otp   string `json:"otp"`
	}
)

func (s *sms) Init(appName string, api core.PluginAPI) (err error) {
	s.pluginApi = api
	s.conf, err = initConfig(&s.rawConf.Config)
	if err != nil {
		return err
	}

	s.app, err = s.pluginApi.GetApp(appName)
	if err != nil {
		return fmt.Errorf("app named '%s' is not declared", appName)
	}

	createRoutes(s)
	return nil
}

func (*sms) GetPluginID() string {
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

func createRoutes(s *sms) {
	routes := []*_interface.Route{
		{
			Method:  "POST",
			Path:    s.conf.PathPrefix + s.conf.SendUrl,
			Handler: SendOtp(s),
		},
		{
			Method:  "POST",
			Path:    s.conf.PathPrefix + s.conf.VerifyUrl,
			Handler: Verify(s),
		},
		{
			Method:  "POST",
			Path:    s.conf.PathPrefix + s.conf.ResendUrl,
			Handler: Resend(s),
		},
	}
	s.pluginApi.GetRouter().AddAppRoutes(s.app.GetName(), routes)
}
