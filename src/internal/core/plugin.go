package core

import (
	"aureole/internal/configs"
	"github.com/fatih/structs"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gofiber/fiber/v2"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/mitchellh/mapstructure"
)

type Plugin interface {
	MetadataGetter
	RoutesGetter
}

type MetadataGetter interface {
	GetMetadata() Metadata
}

type (
	User struct {
		ID            string  `json:"id,omitempty" mapstructure:"id,omitempty"`
		Email         *string `json:"email,omitempty" mapstructure:"email,omitempty"`
		Phone         *string `json:"phone,omitempty" mapstructure:"phone,omitempty"`
		Username      *string `json:"username,omitempty" mapstructure:"username,omitempty"`
		EmailVerified bool    `json:"email_verified,omitempty" mapstructure:"email_verified,omitempty"`
		PhoneVerified bool    `json:"phone_verified,omitempty" mapstructure:"phone_verified,omitempty"`
		IsMFAEnabled  *bool   `json:"is_mfa_enabled,omitempty" mapstructure:"is_mfa_enabled,omitempty"`
		EnabledMFAs   []int   `json:"enabled_mfas,omitempty" mapstructure:"enabled_mfas,omitempty"`
	}

	ImportedUser struct {
		PluginID     string                 `json:"plugin_id,omitempty" mapstructure:"plugin_id,omitempty"`
		ProviderName string                 `json:"provider_name,omitempty" mapstructure:"provider_name,omitempty"`
		ProviderId   string                 `json:"provider_id,omitempty" mapstructure:"provider_id,omitempty"`
		UserId       string                 `json:"user_id,omitempty" mapstructure:"user_id,omitempty"`
		Additional   map[string]interface{} `json:"additional,omitempty" mapstructure:"additional,omitempty"`
	}

	Secrets map[string]Secret
	Secret  *string
)

type (
	AuthenticatorCreator interface {
		Create(*configs.PluginConfig) CryptoStorage
	}
	Authenticator interface {
		Plugin
		GetAuthHandler() AuthHandlerFunc
		GetAuthHTTPMethod() string
		GetOAS3AuthRequestBody() *openapi3.RequestBody
		GetOAS3AuthParameters() openapi3.Parameters
	}

	AuthResult struct {
		User         *User
		UserFromDB   *User
		ImportedUser *ImportedUser
		Secrets      *Secrets
		Cred         *Credential
		Provider     string
		ProviderId   string
		Additional   map[string]interface{}
		ErrorData    map[string]interface{}
	}

	AuthHandlerFunc func(fiber.Ctx) (*AuthResult, error)
)

const (
	Private KeyType = "private"
	Public  KeyType = "public"
)

type (
	// CryptoKeyPluginCreator defines methods for authentication Plugin
	CryptoKeyPluginCreator interface {
		// Create returns desired crypto key depends on the given config
		Create(config *configs.PluginConfig) CryptoKey
	}

	CryptoKey interface {
		Plugin
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
		Plugin
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
	Attempts       = "attempts"
	OAuth2Data     = "oauth2"
	AdditionalData = "additional"
	Link           = "link"
	Token          = "token"
	State          = "state"
	Sub            = "sub"
	Hotp           = "hotp"
	MIMEImagePNG   = "image/png"
	MIMECredential = "credential"
)

type (
	IDManagerCreator interface {
		Create(manager *configs.PluginConfig) IDManager
	}

	IDManager interface {
		Plugin
		RegisterOrUpdate(authRes *AuthResult) (*AuthResult, error)
		SetSecret(cred *Credential, pluginId string, secret Secret) error
		GetSecret(cred *Credential, pluginId, secret string) (Secret, error)
		SetSecrets(cred *Credential, pluginId string, payload *Secrets) error
		GetSecrets(userId, pluginId string) (*Secrets, error)
		GetUser(cred *Credential) (*User, error)
		GetData(c *Credential, authnProvider string, name string) (interface{}, error)
		IsMFAEnabled(c *Credential) (bool, error)
		//todo get rid of
		OnMFA(c *Credential, data *MFAData) error
	}

	Credential struct {
		Name  string
		Value string
	}

	Identity struct {
		ID            interface{}            `mapstructure:"id,omitempty" json:"id,omitempty"`
		Email         *string                `mapstructure:"email,omitempty" json:"email,omitempty"`
		Phone         *string                `mapstructure:"phone,omitempty" json:"phone,omitempty"`
		Username      *string                `mapstructure:"username,omitempty" json:"username,omitempty"`
		EmailVerified bool                   `mapstructure:"email_verified" json:"email_verified"`
		PhoneVerified bool                   `mapstructure:"phone_verified" json:"phone_verified"`
		Additional    map[string]interface{} `mapstructure:"additional,omitempty" json:"additional,omitempty"`
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

func NewUser(data map[string]interface{}) (*User, error) {
	user := &User{}
	err := mapstructure.Decode(data, user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u *User) AsMap() map[string]interface{} {
	structsConf := structs.New(u)
	structsConf.TagName = "mapstructure"
	return structsConf.Map()
}

func (iu *ImportedUser) AsMap() map[string]interface{} {
	structsConf := structs.New(iu)
	structsConf.TagName = "mapstructure"
	return structsConf.Map()
}

func (i *Identity) AsMap() map[string]interface{} {
	structsConf := structs.New(i)
	structsConf.TagName = "mapstructure"
	return structsConf.Map()
}

type (
	IssuerCreator interface {
		Create(*configs.PluginConfig) Issuer
	}
	Issuer interface {
		Plugin
		GetOAS3SuccessResponse() (*openapi3.Response, error)
		GetNativeQueries() map[string]string
		GetVerifyKeys() map[string]CryptoKey
		Authorize(*fiber.Ctx, *IssuerPayload) error
	}

	IssuerPayload struct {
		ID            interface{}            `mapstructure:"id,omitempty" json:"id,omitempty"`
		AureoleID     *string                `mapstructure:"aureole_id,omitempty" json:"aureole_id,omitempty"`
		Username      *string                `mapstructure:"username,omitempty" json:"username,omitempty"`
		Phone         *string                `mapstructure:"phone,omitempty" json:"phone,omitempty"`
		Email         *string                `mapstructure:"email,omitempty" json:"email,omitempty"`
		EmailVerified bool                   `mapstructure:"email_verified" json:"email_verified"`
		PhoneVerified bool                   `mapstructure:"phone_verified" json:"phone_verified"`
		Additional    map[string]interface{} `mapstructure:"additional,omitempty" json:"additional,omitempty"`
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
	MFACreator interface {
		Create(*configs.PluginConfig) MFA
	}

	MFA interface {
		Plugin
		IsEnabled(cred *Credential) (bool, error)
		InitMFA() MFAInitFunc
		Verify() MFAVerifyFunc
		GetOAS3AuthRequestBody() *openapi3.RequestBody
		GetOAS3AuthParameters() openapi3.Parameters
	}

	VerifyRequest interface {
		GetOAS3VerifyRequestBody() *openapi3.RequestBody
		GetOAS3VerifyParameters() openapi3.Parameters
	}

	MFAVerifyFunc func(fiber.Ctx) (cred *Credential, errorData MFAResMap, err error)
	MFAInitFunc   func(fiber.Ctx) (mfaData MFAResMap, err error)
	MFAResMap     map[string]interface{}
)

type (
	// RootPluginCreator defines methods for admin Plugin
	RootPluginCreator interface {
		// Create returns desired admin Plugin depends on the given config
		Create(admin *configs.PluginConfig) RootPlugin
	}

	RootPlugin interface {
		Plugin
	}
)

type (
	SenderCreator interface {
		// Create returns desired messenger depends on the given config
		Create(*configs.PluginConfig) Sender
	}

	Sender interface {
		Plugin
		Send(recipient, subject, tmpl, tmplExtension string, tmplCtx map[string]interface{}) error
		SendRaw(recipient, subject, message string) error
	}
)

type (
	StorageCreator interface {
		Create(*configs.PluginConfig) Storage
	}

	Storage interface {
		Plugin
		Set(k string, v interface{}, exp int) error
		Get(k string, v interface{}) (ok bool, err error)
		Delete(k string) error
		Exists(k string) (found bool, err error)
		Close() error
	}
)
