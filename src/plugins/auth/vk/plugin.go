package vk

import (
	"aureole/internal/configs"
	"aureole/internal/core"
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3gen"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/mitchellh/mapstructure"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/endpoints"

	_ "embed"
	"errors"
)

//go:embed meta.yaml
var rawMeta []byte

var meta core.Metadata

// init initializes package by register pluginCreator
func init() {
	meta = core.AuthenticatorRepo.Register(rawMeta, Create)
}

type (
	GetAuthHandlerQuery struct {
		State string `query:"state"`
		Code  string `query:"code"`
	}
)

type vk struct {
	pluginAPI core.PluginAPI
	rawConf   configs.AuthPluginConfig
	conf      *config
	provider  *oauth2.Config
}

func (v *vk) GetAuthHTTPMethod() string {
	return http.MethodGet
}

func Create(conf configs.AuthPluginConfig) core.Authenticator {
	return &vk{rawConf: conf}
}

func (v *vk) Init(api core.PluginAPI) (err error) {
	v.pluginAPI = api
	v.conf, err = initConfig(&v.rawConf.Config)
	if err != nil {
		return err
	}

	if err := initProvider(v); err != nil {
		return err
	}

	return nil
}

func (vk) GetMetadata() core.Metadata {
	return meta
}

func (v *vk) GetAuthHandler() core.AuthHandlerFunc {
	return func(c fiber.Ctx) (*core.AuthResult, error) {
		input := &GetAuthHandlerQuery{}
		if err := c.QueryParser(input); err != nil {
			return nil, err
		}

		if input.State != "state" {
			return nil, errors.New("invalid state")
		}
		if input.Code == "" {
			return nil, errors.New("code not found")
		}

		userData, err := getUserData(v, input.Code)
		if err != nil {
			return nil, err
		}

		var (
			pluginId   = fmt.Sprintf("%d", meta.PluginID)
			email      = userData["email"].(string)
			providerId = userData["id"].(string)
		)

		ok, err := v.pluginAPI.Filter(convertUserData(userData), v.conf.Filter)
		if err != nil {
			return nil, err
		} else if !ok {
			return nil, errors.New("input data doesn't pass filters")
		}

		return &core.AuthResult{
			User: &core.User{
				Email:         &email,
				EmailVerified: true,
			},
			ImportedUser: &core.ImportedUser{
				ProviderId:   &providerId,
				PluginID:     &pluginId,
				ProviderName: &meta.ShortName,
				Additional:   userData,
			},
			Cred: &core.Credential{
				Name:  core.Email,
				Value: email,
			},
			Provider: "social_provider$" + v.GetMetadata().ShortName,
		}, nil
	}
}

func (v *vk) GetOAS3AuthRequestBody() *openapi3.RequestBody {
	return nil
}

func (v *vk) GetOAS3AuthParameters() openapi3.Parameters {
	stateSchema, _ := openapi3gen.NewSchemaRefForValue(GetAuthHandlerQuery{}.State, nil)
	codeSchema, _ := openapi3gen.NewSchemaRefForValue(GetAuthHandlerQuery{}.Code, nil)
	return openapi3.Parameters{
		{
			Value: &openapi3.Parameter{
				Name:     "State",
				In:       "query",
				Required: true,
				Schema:   stateSchema,
			},
		},
		{
			Value: &openapi3.Parameter{
				Name:     "Code",
				In:       "query",
				Required: true,
				Schema:   codeSchema,
			},
		},
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

func initProvider(v *vk) error {
	redirectUri := v.pluginAPI.GetAppUrl()
	redirectUri.Path = v.pluginAPI.GetAuthRoute(meta.ShortName)
	v.provider = &oauth2.Config{
		ClientID:     v.conf.ClientId,
		ClientSecret: v.conf.ClientSecret,
		Endpoint:     endpoints.Vk,
		RedirectURL:  redirectUri.String(),
		Scopes:       v.conf.Scopes,
	}
	return nil
}

func (v *vk) GetCustomAppRoutes() []*core.Route {
	return []*core.Route{
		{
			Method:        http.MethodGet,
			Path:          core.GetOAuthPathPostfix(),
			Handler:       getAuthCode(v),
			OAS3Operation: assembleOAS3Operation(),
		},
	}
}

func assembleOAS3Operation() *openapi3.Operation {
	description := "Redirect"
	return &openapi3.Operation{
		OperationID: meta.ShortName,
		Description: meta.DisplayName,
		Tags:        []string{fmt.Sprintf("auth by %s", meta.DisplayName)},
		Responses: map[string]*openapi3.ResponseRef{
			strconv.Itoa(http.StatusFound): {
				Value: core.AssembleOASRedirectResponse(&description),
			},
		},
	}
}
