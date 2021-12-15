package plugins

import (
	"aureole/internal/configs"
	"errors"
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

var UserNotExistError = errors.New("user doesn't exists")

type (
	IDManagerAdapter interface {
		Create(manager *configs.IDManager) IDManager
	}

	IDManager interface {
		MetaDataGetter
		Register(c *Credential, i *Identity, authnProvider string) (*Identity, error)
		OnUserAuthenticated(c *Credential, i *Identity, authnProvider string) (*Identity, error)
		GetData(c *Credential, authnProvider string, name string) (interface{}, error)
		Update(c *Credential, i *Identity, authnProvider string) (*Identity, error)

		On2FA(c *Credential, data *MFAData) error
		Get2FAData(c *Credential, mfaID string) (*MFAData, error)

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

	MFAData struct {
		PluginID     string
		ProviderName string
		Payload      map[string]interface{}
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
