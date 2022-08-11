package facebook

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
	state struct {
		State string `query:"state"`
	}
	code struct {
		Code string `query:"code"`
	}
	GetAuthHandlerQuery struct {
		State string `query:"state"`
		Code  string `query:"code"`
	}
)

type facebook struct {
	pluginAPI core.PluginAPI
	rawConf   configs.AuthPluginConfig
	conf      *config
	provider  *oauth2.Config
}

func (f *facebook) GetAuthHTTPMethod() string {
	return http.MethodGet
}

func Create(conf configs.AuthPluginConfig) core.Authenticator {
	return &facebook{rawConf: conf}
}

func (f *facebook) Init(api core.PluginAPI) (err error) {
	f.pluginAPI = api
	f.conf, err = initConfig(&f.rawConf.Config)
	if err != nil {
		return err
	}
	if err := initProvider(f); err != nil {
		return err
	}

	return nil
}

func (facebook) GetMetadata() core.Metadata {
	return meta
}

func (f *facebook) GetAuthHandler() core.AuthHandlerFunc {
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

		userData, err := getUserData(f, input.Code)
		if err != nil {
			return nil, err
		}

		var (
			pluginId   = fmt.Sprintf("%d", meta.PluginID)
			email      = userData["email"].(string)
			providerId = userData["id"].(string)
		)

		ok, err := f.pluginAPI.Filter(convertUserData(userData), f.conf.Filter)
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
			Provider: "social_provider$" + meta.ShortName,
		}, nil
	}
}

func (f *facebook) GetOAS3AuthRequestBody() *openapi3.RequestBody {
	return nil
}

func (f *facebook) GetOAS3AuthParameters() openapi3.Parameters {
	stateSchema, _ := openapi3gen.NewSchemaRefForValue(state{}, nil)
	codeSchema, _ := openapi3gen.NewSchemaRefForValue(code{}, nil)
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

func initProvider(f *facebook) error {
	redirectUri := f.pluginAPI.GetAppUrl()
	redirectUri.Path = f.pluginAPI.GetAuthRoute(meta.ShortName)
	f.provider = &oauth2.Config{
		ClientID:     f.conf.ClientId,
		ClientSecret: f.conf.ClientSecret,
		Endpoint:     endpoints.Facebook,
		RedirectURL:  redirectUri.String(),
		Scopes:       f.conf.Scopes,
	}
	return nil
}

func (f *facebook) GetCustomAppRoutes() []*core.Route {
	return []*core.Route{
		{
			Method:        http.MethodGet,
			Path:          core.GetOAuthPathPostfix(),
			Handler:       getAuthCode(f),
			OAS3Operation: assembleOAS3Operation(),
		},
	}
}

func assembleOAS3Operation() *openapi3.Operation {
	description := "Redirect"
	return &openapi3.Operation{
		OperationID: meta.ShortName,
		Description: meta.DisplayName,
		Tags:        []string{fmt.Sprintf("auth by %s", meta.ShortName)},
		Responses: map[string]*openapi3.ResponseRef{
			strconv.Itoa(http.StatusFound): {
				Value: core.AssembleOASRedirectResponse(&description),
			},
		},
	}
}
