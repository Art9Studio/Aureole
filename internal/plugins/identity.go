package plugins

import (
	"aureole/internal/configs"
	"fmt"

	"github.com/fatih/structs"

	"github.com/mitchellh/mapstructure"
)

var IDManagerRepo = createRepository()

const (
	ID             = "id"
	SocialID       = "social_id"
	Email          = "email"
	Phone          = "phone"
	Username       = "username"
	EmailVerified  = "email_verified"
	PhoneVerified  = "phone_verified"
	Password       = "password"
	SecondFactorID = "2fa_id"
	AuthNProvider  = "provider"
	OAuth2Data     = "oauth2"
	AdditionalData = "additional"
)

type (
	IDManagerAdapter interface {
		Create(manager *configs.IDManager) IDManager
	}

	IDManager interface {
		Register(c *Credential, i *Identity, authnProvider string) (*Identity, error)
		OnUserAuthenticated(c *Credential, i *Identity, authnProvider string) (*Identity, error)
		On2FA(c *Credential, mfaProvider string, data map[string]interface{}) error

		GetData(c *Credential, authnProvider string, name string) (interface{}, error)
		Get2FAData(c *Credential, mfaProvider string) (map[string]interface{}, error)

		Update(c *Credential, i *Identity, authnProvider string) (*Identity, error)
		CheckFeaturesAvailable(features []string) error
	}

	Credential struct {
		Name  string
		Value string
	}

	Identity struct {
		ID            interface{}            `mapstructure:"id,omitempty"`
		Email         *string                `mapstructure:"email,omitempty"`
		Phone         *string                `mapstructure:"phone,omitempty"`
		Username      *string                `mapstructure:"username,omitempty"`
		EmailVerified bool                   `mapstructure:"email_verified"`
		PhoneVerified bool                   `mapstructure:"phone_verified"`
		Additional    map[string]interface{} `mapstructure:"additional,omitempty"`
	}
)

func NewIDManager(conf *configs.IDManager) (IDManager, error) {
	a, err := IDManagerRepo.Get(conf.Type)
	if err != nil {
		return nil, err
	}

	adapter, ok := a.(IDManagerAdapter)
	if !ok {
		return nil, fmt.Errorf("trying to cast adapter was failed: %v", err)
	}

	return adapter.Create(conf), nil
}

func NewIdentity(data map[string]interface{}) (*Identity, error) {
	ident := &Identity{}
	err := mapstructure.Decode(data, ident)
	if err != nil {
		return nil, err
	}
	return ident, nil
}

func (i *Identity) AsMap() map[string]interface{} {
	structsConf := structs.New(i)
	structsConf.TagName = "mapstructure"
	return structsConf.Map()
}
