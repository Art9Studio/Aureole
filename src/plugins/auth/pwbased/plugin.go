package pwbased

import (
	"aureole/internal/configs"
	"aureole/internal/core"
	"aureole/plugins/auth/pwbased/pwhasher"
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

const password = "password"

type (
	pwBased struct {
		pluginAPI core.PluginAPI
		rawConf   configs.AuthPluginConfig
		conf      *config
		pwHasher  pwhasher.PWHasher
		reset     struct {
			sender      core.Sender
			tmpl        string
			tmplExt     string
			confirmLink *url.URL
		}
		verify struct {
			sender      core.Sender
			tmpl        string
			tmplExt     string
			confirmLink *url.URL
		}
	}

	RegisterReqBody struct {
		Id       interface{} `json:"id"`
		Email    string      `json:"email"`
		Phone    string      `json:"phone"`
		Username string      `json:"username"`
		Password string      `json:"password"`
	}

	ResetConfirmReqBody struct {
		RegisterReqBody
	}

	VerifyReqBody struct {
		Email string `json:"email"`
	}

	ResetReqBody struct {
		VerifyReqBody
	}

	ResetConfirmQuery struct {
		Token string `query:"token"`
		URL   string `query:"redirect_url"`
	}

	VerifyConfirmQuery struct {
		Token string `query:"token"`
		URL   string `query:"redirect_url"`
	}

	OAS3Operation struct {
		openapi3.Operation
	}

	linkType string
)

type (
	VerifyConfirmRes struct {
		Success bool `json:"success"`
	}
)

const (
	ResetLink  linkType = "reset"
	VerifyLink linkType = "verify"
)

func (p *pwBased) GetAuthHTTPMethod() string {
	return http.MethodPost
}

func Create(conf configs.AuthPluginConfig) core.Authenticator {
	return &pwBased{rawConf: conf}
}

func (p *pwBased) Init(api core.PluginAPI) error {
	var err error
	p.pluginAPI = api
	p.conf, err = initConfig(&p.rawConf.Config)
	if err != nil {
		return err
	}

	_, ok := p.pluginAPI.GetIDManager()
	if !ok {
		return fmt.Errorf("manager for app '%s' is not declared", p.pluginAPI.GetAppName())
	}

	p.pwHasher, err = pwhasher.NewPWHasher(p.conf.MainHasher)
	if err != nil {
		return fmt.Errorf("cannot intit pwbased authn: %v", err)
	}

	if resetEnabled(p) {
		p.reset.sender, ok = p.pluginAPI.GetSender(p.conf.Reset.Sender)
		if !ok {
			return fmt.Errorf("sender named '%s' is not declared", p.conf.Reset.Sender)
		}

		resetTmpl, err := os.ReadFile(p.conf.Reset.TmplPath)
		if err != nil {
			p.reset.tmpl = defaultResetTmpl
			p.reset.tmplExt = "txt"
		} else {
			p.reset.tmpl = string(resetTmpl)
			p.reset.tmplExt = path.Ext(p.conf.Reset.TmplPath)
		}

		p.reset.confirmLink = createConfirmLink(ResetLink, p)
	}

	if verifyEnabled(p) {
		p.verify.sender, ok = p.pluginAPI.GetSender(p.conf.Verify.Sender)
		if !ok {
			return fmt.Errorf("sender named '%s' is not declared", p.conf.Verify.Sender)
		}

		verifyTmpl, err := os.ReadFile(p.conf.Verify.TmplPath)
		if err != nil {
			p.verify.tmpl = defaultVerifyTmpl
			p.verify.tmplExt = "txt"
		} else {
			p.verify.tmpl = string(verifyTmpl)
			p.verify.tmplExt = path.Ext(p.conf.Verify.TmplPath)
		}

		p.verify.confirmLink = createConfirmLink(VerifyLink, p)
	}

	return nil
}

func (pwBased) GetMetadata() core.Metadata {
	return meta
}

func (p *pwBased) GetAuthHandler() core.AuthHandlerFunc {
	return func(c fiber.Ctx) (*core.AuthResult, error) {
		input := &RegisterReqBody{}
		if err := c.BodyParser(input); err != nil {
			return nil, err
		}
		if input.Password == "" || input.Email == "" {
			return nil, errors.New("password and email are required")
		}

		var (
			user = &core.User{
				Email:    &input.Email,
				Phone:    &input.Phone,
				Username: &input.Username,
			}
			cred = &core.Credential{
				Name:  core.Email,
				Value: input.Email,
			}
		)

		manager, ok := p.pluginAPI.GetIDManager()
		if !ok {
			return nil, fmt.Errorf("id manager for app '%s' is required but not declared", p.pluginAPI.GetAppName())
		}

		pw, err := manager.GetSecret(cred, fmt.Sprintf("%d", meta.PluginID), password)
		if err != nil {
			return nil, err
		}

		isMatch, err := p.pwHasher.ComparePw(input.Password, *pw)
		if err != nil {
			return nil, err
		}

		if isMatch {
			return &core.AuthResult{
				User:     user,
				Cred:     cred,
				Provider: meta.ShortName,
			}, nil
		} else {
			return nil, errors.New("wrong password")
		}
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

func resetEnabled(p *pwBased) bool {
	return p.conf.Reset.Sender != "" && p.conf.Reset.TmplPath != ""
}

func verifyEnabled(p *pwBased) bool {
	return p.conf.Verify.Sender != "" && p.conf.Verify.TmplPath != ""
}

func createConfirmLink(linkType linkType, p *pwBased) *url.URL {
	u := p.pluginAPI.GetAppUrl()
	switch linkType {
	case ResetLink:
		u.Path = path.Clean(u.Path + resetConfirmUrl)
	case VerifyLink:
		u.Path = path.Clean(u.Path + verifyConfirmUrl)
	}
	return &u
}

func (p *pwBased) GetOAS3AuthRequestBody() *openapi3.RequestBody {
	credentialSchema, _ := openapi3gen.NewSchemaRefForValue(RegisterReqBody{}, nil)
	return &openapi3.RequestBody{
		Description: "Credential Info",
		Required:    true,
		Content: map[string]*openapi3.MediaType{
			fiber.MIMEApplicationJSON: {
				Schema: credentialSchema,
			},
		},
	}
}

func (p *pwBased) GetOAS3AuthParameters() openapi3.Parameters {
	return openapi3.Parameters{}
}

func (p *pwBased) GetCustomAppRoutes() []*core.Route {
	credentialSchema, _ := openapi3gen.NewSchemaRefForValue(RegisterReqBody{}, nil)
	emailSchema, _ := openapi3gen.NewSchemaRefForValue(VerifyReqBody{}, nil)
	resetConfirmQuerySchema, _ := openapi3gen.NewSchemaRefForValue(ResetConfirmQuery{}, nil)
	verifyConfirmQuerySchema, _ := openapi3gen.NewSchemaRefForValue(VerifyConfirmQuery{}, nil)

	routes := []*core.Route{
		{
			Method:        http.MethodPost,
			Path:          pathPrefix + registerUrl,
			Handler:       register(p),
			OAS3Operation: assembleOAS3Operation(credentialSchema),
		},
	}

	if resetEnabled(p) {
		resetRoutes := []*core.Route{
			{
				Method:        http.MethodPost,
				Path:          pathPrefix + resetUrl,
				Handler:       Reset(p),
				OAS3Operation: assembleOAS3Operation(emailSchema),
			},
			{
				Method:  http.MethodPost,
				Path:    pathPrefix + resetConfirmUrl,
				Handler: ResetConfirm(p),
				OAS3Operation: Redirect(
					Params(
						assembleOAS3Operation(credentialSchema),
						resetConfirmQuerySchema,
					),
				),
			},
		}
		routes = append(routes, resetRoutes...)
	}

	if verifyEnabled(p) {
		verifyRoutes := []*core.Route{
			{
				Method:        http.MethodPost,
				Path:          pathPrefix + verifyUrl,
				Handler:       Verify(p),
				OAS3Operation: assembleOAS3Operation(emailSchema),
			},
			{
				Method:  http.MethodPost,
				Path:    pathPrefix + verifyConfirmUrl,
				Handler: VerifyConfirm(p),
				OAS3Operation: Redirect(
					Params(
						NoRequestBody(assembleOAS3Operation(nil)),
						verifyConfirmQuerySchema,
					),
				),
			},
		}
		routes = append(routes, verifyRoutes...)
	}
	return routes
}

func assembleOAS3Operation(reqSchema *openapi3.SchemaRef) *openapi3.Operation {
	okResponse := "OK"
	badReqResponse := "BadRequest"
	internalErrResponse := "Internal Server Error"

	operation := &openapi3.Operation{
		OperationID: meta.ShortName,
		Description: meta.DisplayName,
		Tags:        []string{fmt.Sprintf("auth by %s", meta.DisplayName)},
		RequestBody: &openapi3.RequestBodyRef{
			Value: &openapi3.RequestBody{
				Required: true,
				Content: map[string]*openapi3.MediaType{
					fiber.MIMEApplicationJSON: {
						Schema: reqSchema,
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
				Value: core.AssembleOAS3ErrResponse(&internalErrResponse),
			},
		},
	}

	return operation
}

// Redirect adds 302 Status found response to openapi3.Operation.Responses
func Redirect(op *openapi3.Operation) *openapi3.Operation {
	redirectDesc := "Redirect"
	op.Responses[strconv.Itoa(http.StatusFound)] = &openapi3.ResponseRef{
		Value: core.AssembleOASRedirectResponse(&redirectDesc),
	}
	return op
}

// Params adds query parameters to openapi3.Operation.Parameters
func Params(op *openapi3.Operation, schema *openapi3.SchemaRef) *openapi3.Operation {
	op.Parameters = []*openapi3.ParameterRef{
		{
			Value: &openapi3.Parameter{
				Name:        "token",
				In:          "query",
				Description: "ResetConfirmQuery",
				Schema:      schema,
			},
		},
	}
	return op
}

// NoRequestBody removes RequestBody from openapi3.Operation
func NoRequestBody(op *openapi3.Operation) *openapi3.Operation {
	op.RequestBody = nil
	return op
}
