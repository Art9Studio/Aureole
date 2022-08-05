package google

import (
	"aureole/internal/configs"
	"aureole/internal/core"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3gen"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/mitchellh/mapstructure"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/endpoints"
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

type google struct {
	pluginAPI core.PluginAPI
	rawConf   configs.AuthPluginConfig
	conf      *config
	provider  *oauth2.Config
}

func (g *google) GetAuthHTTPMethod() string {
	return http.MethodGet
}

func Create(conf configs.AuthPluginConfig) core.Authenticator {
	return &google{rawConf: conf}
}

func (g *google) Init(api core.PluginAPI) (err error) {
	g.pluginAPI = api
	g.conf, err = initConfig(&g.rawConf.Config)
	if err != nil {
		return err
	}
	if err := initProvider(g); err != nil {
		return err
	}

	return nil
}

func (google) GetMetadata() core.Metadata {
	return meta
}

func (g *google) GetAuthHandler() core.AuthHandlerFunc {
	return func(c fiber.Ctx) (*core.AuthResult, error) {
		input := &GetAuthHandlerQuery{}
		if err := c.QueryParser(input); err != nil {
			return nil, err
		}
		// todo: save state and compare later #2
		state := input.State
		if state != "state" {
			return nil, errors.New("invalid state")
		}
		code := input.Code
		if code == "" {
			return nil, errors.New("code not found")
		}

		var email string
		jwtT, err := getJwt(g, code)
		if err != nil {
			return nil, errors.New("error while exchange")
		}
		err = g.pluginAPI.GetFromJWT(jwtT, "email", &email)
		if err != nil {
			return nil, errors.New("cannot get email from token")
		}
		/*socialId, ok := jwtT.Get("sub")
		if !ok {
			return nil, errors.New("can't get 'social_id' from token")
		}*/
		userData, err := jwtT.AsMap(context.Background())
		if err != nil {
			return nil, err
		}

		ok, err := g.pluginAPI.Filter(convertUserData(userData), g.conf.Filter)
		if err != nil {
			return nil, err
		} else if !ok {
			return nil, errors.New("input data doesn't pass filters")
		}

		return &core.AuthResult{
			Cred: &core.Credential{
				Name:  core.Email,
				Value: email,
			},
			Identity: &core.Identity{
				Email:         &email,
				EmailVerified: true,
				Additional: map[string]interface{}{
					"social_provider_data": map[string]interface{}{
						"plugin_name": meta.ShortName, "payload": userData,
					},
				},
			},
			Provider: "social_provider$" + meta.ShortName,
		}, nil
	}
}

func (g *google) GetOAS3AuthRequestBody() *openapi3.RequestBody {
	return nil
}

func (g *google) GetOAS3AuthParameters() openapi3.Parameters {
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

func initProvider(g *google) error {
	redirectUri := g.pluginAPI.GetAppUrl()
	redirectUri.Path = g.pluginAPI.GetAuthRoute(meta.ShortName)
	g.provider = &oauth2.Config{
		ClientID:     g.conf.ClientId,
		ClientSecret: g.conf.ClientSecret,
		Endpoint:     endpoints.Google,
		RedirectURL:  redirectUri.String(),
		Scopes:       g.conf.Scopes,
	}
	return nil
}

func (g *google) GetCustomAppRoutes() []*core.Route {
	return []*core.Route{
		{
			Method:        http.MethodGet,
			Path:          core.GetOAuthPathPostfix(),
			Handler:       getAuthCode(g),
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
