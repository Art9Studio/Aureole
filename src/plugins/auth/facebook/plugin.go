package facebook

import (
	"aureole/internal/configs"
	"aureole/internal/core"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3gen"
	"net/http"
	"path"
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
	GetAuthHandlerReqBody struct {
		state
		code
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
		input := &GetAuthHandlerReqBody{}
		if err := c.QueryParser(input); err != nil {
			return nil, err
		}

		state := input.State
		if state != "state" {
			return nil, errors.New("invalid state")
		}
		code := input.Code
		if code == "" {
			return nil, errors.New("code not found")
		}

		userData, err := getUserData(f, code)
		if err != nil {
			return nil, err
		}

		ok, err := f.pluginAPI.Filter(convertUserData(userData), f.conf.Filter)
		if err != nil {
			return nil, err
		} else if !ok {
			return nil, errors.New("input data doesn't pass filters")
		}
		email := userData["email"].(string)

		return &core.AuthResult{
			Cred: &core.Credential{
				Name:  core.Email,
				Value: email,
			},
			Identity: &core.Identity{
				Email:         &email,
				EmailVerified: true,
				Additional:    map[string]interface{}{"social_provider_data": userData},
			},
			Provider: "social_provider$" + meta.ShortName,
		}, nil
	}
}

func (f *facebook) GetOAS3AuthRequestBody() *openapi3.RequestBody {
	return &openapi3.RequestBody{}
}

func (f *facebook) GetOAS3AuthParameters() *openapi3.Parameters {
	stateSchema, _ := openapi3gen.NewSchemaRefForValue(state{}, nil)
	codeSchema, _ := openapi3gen.NewSchemaRefForValue(code{}, nil)
	return &openapi3.Parameters{
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
	url := f.pluginAPI.GetAppUrl()
	url.Path = path.Clean(url.Path + pathPrefix + redirectUrl)
	f.provider = &oauth2.Config{
		ClientID:     f.conf.ClientId,
		ClientSecret: f.conf.ClientSecret,
		Endpoint:     endpoints.Facebook,
		RedirectURL:  url.String(),
		Scopes:       f.conf.Scopes,
	}
	return nil
}

func (f *facebook) GetCustomAppRoutes() []*core.Route {

	return []*core.Route{
		{
			Method:        http.MethodGet,
			Path:          pathPrefix,
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
		Responses: map[string]*openapi3.ResponseRef{
			strconv.Itoa(http.StatusFound): {
				Value: core.AssembleOASRedirectResponse(&description),
			},
		},
	}
}
