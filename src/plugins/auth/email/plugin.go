package email

import (
	"aureole/internal/configs"
	"aureole/internal/core"
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3gen"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/mitchellh/mapstructure"

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
	email struct {
		pluginAPI     core.PluginAPI
		rawConf       configs.AuthPluginConfig
		conf          *config
		sender        core.Sender
		magicLink     *url.URL
		tmpl, tmplExt string
	}

	SendMagicLinkReqBody struct {
		Email string `json:"email"`
	}

	GetAuthHandlerQuery struct {
		Token string `query:"token"`
	}
)

func (e *email) GetAuthHTTPMethod() string {
	return http.MethodGet
}

func Create(conf configs.AuthPluginConfig) core.Authenticator {
	return &email{rawConf: conf}
}

func (e *email) Init(api core.PluginAPI) (err error) {
	e.pluginAPI = api
	e.conf, err = initConfig(&e.rawConf.Config)
	if err != nil {
		return err
	}

	var ok bool
	e.sender, ok = e.pluginAPI.GetSender(e.conf.Sender)
	if !ok {
		return fmt.Errorf("sender named '%s' is not declared", e.conf.Sender)
	}

	e.magicLink = createMagicLink(e)
	if err != nil {
		return err
	}

	tmpl, err := os.ReadFile(e.conf.TmplPath)
	if err != nil {
		e.tmpl = defaultTmpl
		e.tmplExt = "txt"
	} else {
		e.tmpl = string(tmpl)
		e.tmplExt = path.Ext(e.conf.TmplPath)
	}

	return nil
}

func (email) GetMetadata() core.Metadata {
	return meta
}

func (e *email) GetAuthHandler() core.AuthHandlerFunc {
	return func(c fiber.Ctx) (*core.AuthResult, error) {
		input := &GetAuthHandlerQuery{}
		if err := c.QueryParser(input); err != nil {
			return nil, err
		}

		rawToken := input.Token
		if rawToken == "" {
			return nil, errors.New("token not found")
		}

		var email string
		token, err := e.pluginAPI.ParseJWTService(rawToken)
		if err != nil {
			return nil, errors.New(err.Error())
		}

		if err = e.pluginAPI.GetFromJWT(token, "email", &email); err != nil {
			return nil, errors.New("cannot get email from token")
		}
		if err = e.pluginAPI.InvalidateJWT(token); err != nil {
			return nil, errors.New(err.Error())
		}

		return &core.AuthResult{
			Cred: &core.Credential{
				Name:  core.Email,
				Value: email,
			},
			Identity: &core.Identity{
				Email:         &email,
				EmailVerified: true,
			},
			Provider: meta.ShortName,
		}, nil
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

func createMagicLink(e *email) *url.URL {
	u := e.pluginAPI.GetAppUrl()
	u.Path = path.Clean(u.Path + loginUrl)
	return &u
}

func (e *email) GetOAS3AuthRequestBody() *openapi3.RequestBody {
	return nil
}

func (e *email) GetOAS3AuthParameters() openapi3.Parameters {
	schema, _ := openapi3gen.NewSchemaRefForValue(GetAuthHandlerQuery{}, nil)
	return openapi3.Parameters{
		&openapi3.ParameterRef{
			Value: &openapi3.Parameter{
				Name:     "Token",
				In:       "query",
				Required: true,
				Schema:   schema,
			},
		},
	}
}

func (e *email) GetCustomAppRoutes() []*core.Route {
	return []*core.Route{
		{
			Method:        http.MethodPost,
			Path:          sendUrl,
			Handler:       sendMagicLink(e),
			OAS3Operation: assembleOAS3Operation(),
		},
	}
}

func assembleOAS3Operation() *openapi3.Operation {
	okResponse := "OK"
	badReqResponse := "Bad Request"
	internalErrorResponse := "Internal Server Error"
	inputSchema, _ := openapi3gen.NewSchemaRefForValue(SendMagicLinkReqBody{}, nil)
	operation := &openapi3.Operation{
		OperationID: meta.ShortName,
		Description: meta.DisplayName,
		Tags:        []string{fmt.Sprintf("auth by %s", meta.DisplayName)},
		RequestBody: &openapi3.RequestBodyRef{
			Value: &openapi3.RequestBody{
				Required: true,
				Content: map[string]*openapi3.MediaType{
					fiber.MIMEApplicationJSON: {
						Schema: inputSchema,
					},
				},
			},
		},
		Responses: map[string]*openapi3.ResponseRef{
			strconv.Itoa(http.StatusOK): {
				Value: core.AssembleOAS3OKResponse(&okResponse, nil),
			},
			strconv.Itoa(http.StatusBadRequest): {
				Value: core.AssembleOAS3ErrResponse(&badReqResponse),
			},
			strconv.Itoa(http.StatusInternalServerError): {
				Value: core.AssembleOAS3ErrResponse(&internalErrorResponse),
			},
		},
	}

	return operation
}
