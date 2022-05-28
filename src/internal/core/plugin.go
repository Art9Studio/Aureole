package core

import (
	"aureole/internal/configs"
	"errors"
	"github.com/fatih/structs"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gofiber/fiber/v2"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/mitchellh/mapstructure"
)

type plugin interface {
	MetaDataGetter
	PathsGetter
}

type MetaDataGetter interface {
	GetMetaData() Meta
}

type (
	Authenticator interface {
		plugin
		LoginWrapper() AuthNLoginFunc
		GetAuthRoute() *Route
	}

	AuthNResult struct {
		Cred       *Credential
		Identity   *Identity
		Provider   string
		Additional map[string]interface{}
		ErrorData  map[string]interface{}
	}

	AuthNLoginFunc func(fiber.Ctx) (*AuthNResult, error)
)

const (
	Private KeyType = "private"
	Public  KeyType = "public"
)

type (
	// CryptoKeyPluginCreator defines methods for authentication plugin
	CryptoKeyPluginCreator interface {
		// Create returns desired crypto key depends on the given config
		Create(config *configs.PluginConfig) CryptoKey
	}

	CryptoKey interface {
		plugin
		GetPrivateSet() jwk.Set
		GetPublicSet() jwk.Set
	}

	KeyType string
)

type (
	CryptoStorageCreator interface {
		Create(*configs.PluginConfig) CryptoStorage
	}

	CryptoStorage interface {
		plugin
		Read(v *[]byte) (ok bool, err error)
		Write(v []byte) error
	}
)

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
	IDManagerCreator interface {
		Create(manager *configs.PluginConfig) IDManager
	}

	IDManager interface {
		plugin
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

type (
	Issuer interface {
		plugin
		GetResponseData() *openapi3.Responses
		GetNativeQueries() map[string]string
		Authorize(*fiber.Ctx, *IssuerPayload) error
	}

	IssuerPayload struct {
		ID            interface{}            `mapstructure:"id,omitempty"`
		Username      *string                `mapstructure:"username,omitempty"`
		Phone         *string                `mapstructure:"phone,omitempty"`
		Email         *string                `mapstructure:"email,omitempty"`
		EmailVerified bool                   `mapstructure:"email_verified"`
		PhoneVerified bool                   `mapstructure:"phone_verified"`
		Additional    map[string]interface{} `mapstructure:"additional,omitempty"`
		// NativeQ    func(queryName string, args ...interface{}) string
	}
)

func NewIssuerPayload(data map[string]interface{}) (*IssuerPayload, error) {
	p := &IssuerPayload{}
	if err := mapstructure.Decode(data, p); err != nil {
		return nil, err
	}
	return p, nil
}

type (
	MFA interface {
		plugin
		IsEnabled(cred *Credential) (bool, error)
		Init2FA() MFAInitFunc
		Verify() MFAVerifyFunc
	}

	MFAVerifyFunc func(fiber.Ctx) (cred *Credential, errorData fiber.Map, err error)
	MFAInitFunc   func(fiber.Ctx) (mfaData fiber.Map, err error)
)

type (
	// RootPluginCreator defines methods for admin plugin
	RootPluginCreator interface {
		// Create returns desired admin plugin depends on the given config
		Create(admin *configs.PluginConfig) RootPlugin
	}

	RootPlugin interface {
		plugin
	}
)

type (
	SenderCreator interface {
		// Create returns desired messenger depends on the given config
		Create(*configs.PluginConfig) Sender
	}

	Sender interface {
		plugin
		Send(recipient, subject, tmpl, tmplExtension string, tmplCtx map[string]interface{}) error
		SendRaw(recipient, subject, message string) error
	}
)

type (
	StorageCreator interface {
		Create(*configs.PluginConfig) Storage
	}

	Storage interface {
		plugin
		Set(k string, v interface{}, exp int) error
		Get(k string, v interface{}) (ok bool, err error)
		Delete(k string) error
		Exists(k string) (found bool, err error)
		Close() error
	}
)
