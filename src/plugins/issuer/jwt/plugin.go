package jwt

import (
	"aureole/internal/configs"
	"aureole/internal/core"
	"fmt"
	"net/http"
	"os"
	"path"
	"regexp"
	"strconv"
	txtTmpl "text/template"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3gen"
	"github.com/gofiber/fiber/v2"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jwt"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/mitchellh/mapstructure"

	"context"
	_ "embed"
	"encoding/json"
	"errors"

	"go.uber.org/zap/buffer"
)

//go:embed meta.yaml
var rawMeta []byte

//go:embed default-payload.json.tmpl
var defaultPayloadTmpl []byte

var meta core.Meta

// init initializes package by register pluginCreator
func init() {
	meta = core.IssuerRepo.Register(rawMeta, Create)
}

type (
	jwtIssuer struct {
		pluginAPI     core.PluginAPI
		rawConf       configs.PluginConfig
		conf          *config
		signKey       core.CryptoKey
		verifyKeys    map[string]core.CryptoKey
		nativeQueries map[string]string
	}

	tokenType string
	tokenResp string
)

const (
	accessToken  tokenType = "access"
	refreshToken tokenType = "refresh"
)

const (
	body   tokenResp = "body"
	cookie tokenResp = "cookie"
	//both   tokenResp = "both"
)

var keyMap = map[tokenType]map[tokenResp]string{
	accessToken: {
		cookie: "access_token",
	},
	refreshToken: {
		body:   "refresh",
		cookie: "refresh_token",
	},
}

// INFO: надо было менять в интерфейсе  Issure GetResponseData что бы он возвращал ошибку?
func Create(conf configs.PluginConfig) core.Issuer {
	return &jwtIssuer{rawConf: conf}
}

func (j *jwtIssuer) Init(api core.PluginAPI) (err error) {
	j.pluginAPI = api
	j.conf, err = initConfig(&j.rawConf.Config)
	if err != nil {
		return err
	}

	var ok bool
	j.signKey, ok = j.pluginAPI.GetCryptoKey(j.conf.SignKey)
	if !ok {
		return fmt.Errorf("cannot get crypto key named %s", j.conf.SignKey)
	}

	j.verifyKeys = make(map[string]core.CryptoKey)
	for _, keyName := range j.conf.VerifyKeys {
		j.verifyKeys[keyName], ok = j.pluginAPI.GetCryptoKey(keyName)
		if !ok {
			return fmt.Errorf("cannot get crypto key named %s", j.conf.SignKey)
		}
	}

	if j.conf.AccessTokenBearer == cookie && j.conf.RefreshTokenBearer == cookie {

	}

	/*if j.conf.NativeQueries != "" {
		if j.nativeQueries, err = readNativeQueries(j.conf.NativeQueries); err != nil {
			return err
		}
	}*/

	return err
}

func (j *jwtIssuer) GetResponseData() (*openapi3.Responses, error) {
	responses := openapi3.NewResponses()

	okSchema, err := openapi3gen.NewSchemaRefForValue(Response{}, nil)
	if err != nil {
		return nil, err
	}

	okResponse := openapi3.NewResponse().
		WithDescription("Successfully authorize and return refresh and access tokens").
		WithContent(openapi3.NewContentWithJSONSchema(okSchema.Value))

	okStatus := strconv.Itoa(http.StatusOK)

	if j.conf.AccessTokenBearer == cookie || j.conf.RefreshTokenBearer == cookie {
		header := &openapi3.Header{}
		header.Name = "Set-Cookie"

		if j.conf.AccessTokenBearer == cookie && j.conf.RefreshTokenBearer == cookie {
			header.Description = "Save access JWT and refresh JWT in cookies"
		} else if j.conf.AccessTokenBearer == cookie {
			header.Description = "Save access JWT in cookies"
		} else {
			header.Description = "Save refresh JWT in cookies"
		}

		setCookieSchema := openapi3.NewSchema()
		setCookieSchema.Type = "string"
		header.Schema = openapi3.NewSchemaRef("", setCookieSchema)

		okResponse.Headers = openapi3.Headers{
			"Set-Cookie": &openapi3.HeaderRef{Value: header},
		}
	}

	bodySchema, err := openapi3gen.NewSchemaRefForValue(Response{}, nil)
	if err != nil {
		return nil, err
	}

	okResponse.Content["application/json"].Schema.Value = bodySchema.Value

	responses[okStatus] = &openapi3.ResponseRef{
		Value: okResponse,
	}
	return &responses, nil
}

func (j jwtIssuer) GetMetaData() core.Meta {
	return meta
}

func (j *jwtIssuer) GetNativeQueries() map[string]string {
	return j.nativeQueries
}

func (j *jwtIssuer) Authorize(c *fiber.Ctx, payload *core.IssuerPayload) error {
	accessT, err := newToken(accessToken, j.conf, payload)
	if err != nil {
		return core.SendError(c, fiber.StatusInternalServerError, err.Error())
	}
	refreshT, err := newToken(refreshToken, j.conf, payload)
	if err != nil {
		return core.SendError(c, fiber.StatusInternalServerError, err.Error())
	}

	signedAccessT, err := signToken(j.signKey, accessT)
	if err != nil {
		return core.SendError(c, fiber.StatusInternalServerError, err.Error())
	}
	signedRefreshT, err := signToken(j.signKey, refreshT)
	if err != nil {
		return core.SendError(c, fiber.StatusInternalServerError, err.Error())
	}

	bearers := map[tokenType]tokenResp{
		accessToken:  j.conf.AccessTokenBearer,
		refreshToken: j.conf.RefreshTokenBearer,
	}
	tokens := map[tokenType][]byte{
		accessToken:  signedAccessT,
		refreshToken: signedRefreshT,
	}
	if err := attachTokens(c, bearers, tokens); err != nil {
		return core.SendError(c, fiber.StatusInternalServerError, err.Error())
	}

	return nil
}

func initConfig(conf *configs.RawConfig) (*config, error) {
	PluginConf := &config{}
	if err := mapstructure.Decode(conf, PluginConf); err != nil {
		return nil, err
	}
	PluginConf.setDefaults()

	return PluginConf, nil
}

/*func readNativeQueries(path string) (map[string]string, error) {
	q := map[string]string{}

	f, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(f, &q)
	if err != nil {
		return nil, err
	}

	return q, nil
}*/

func buildOperationForRefreshHandler() *openapi3.Operation {
	operation := openapi3.NewOperation()
	operation.Description = "Refresh JWT"

	responses := openapi3.NewResponses()
	{
		okResp := openapi3.NewResponse().
			WithDescription("Successfully refreshed JWT and returns new access JWT")

		okSchema, err := openapi3gen.NewSchemaRefForValue(RefreshResponse{}, nil)
		if err != nil {
			fmt.Printf("cannot build schema for value: %v", err)
		} else {
			okResp = okResp.WithJSONSchema(okSchema.Value)
		}

		okCode := strconv.Itoa(http.StatusOK)

		responses[okCode].Value = okResp
	}
	{
		badReqResp := openapi3.NewResponse().
			WithDescription("Failed to refresh JWT")

		badReqCode := strconv.Itoa(http.StatusBadRequest)
		responses[badReqCode].Value = badReqResp
	}
	operation.Responses = responses
	return operation
}

func (j *jwtIssuer) GetAppRoutes() []*core.Route {
	operation := buildOperationForRefreshHandler()

	return []*core.Route{
		{
			Method:    http.MethodPost,
			Path:      refreshUrl,
			Operation: operation,
			Handler:   refresh(j),
		},
	}
}

func newToken(tokenType tokenType, conf *config, payload *core.IssuerPayload) (t jwt.Token, err error) {
	switch tokenType {
	case accessToken:
		token := jwt.New()
		// todo: think about multiple errors handling
		jti, err := gonanoid.New(16)
		if err != nil {
			return nil, err
		}
		err = token.Set(jwt.JwtIDKey, jti)
		if err != nil {
			return nil, err
		}
		err = token.Set(jwt.IssuerKey, conf.Iss)
		if err != nil {
			return nil, err
		}
		err = token.Set(jwt.AudienceKey, conf.Aud)
		if err != nil {
			return nil, err
		}
		err = token.Set(jwt.NotBeforeKey, conf.Nbf)
		if err != nil {
			return nil, err
		}

		if conf.Sub && payload.ID != nil {
			err := token.Set(jwt.SubjectKey, fmt.Sprintf("%v", payload.ID))
			if err != nil {
				return nil, err
			}
		}

		currTime := time.Now()
		nbf := currTime.Add(time.Duration(conf.Nbf) * time.Second).Unix()
		err = token.Set(jwt.NotBeforeKey, nbf)
		if err != nil {
			return nil, err
		}

		if conf.Iat {
			err := token.Set(jwt.IssuedAtKey, currTime.Unix())
			if err != nil {
				return nil, err
			}
		}

		exp := currTime.Add(time.Duration(conf.AccessExp) * time.Second).Unix()
		err = token.Set(jwt.ExpirationKey, exp)
		if err != nil {
			return nil, err
		}

		p, err := parsePayload(conf.TmplPath, payload)
		if err != nil {
			return nil, err
		}
		for k, v := range p {
			err := token.Set(k, v)
			if err != nil {
				return nil, err
			}
		}

		t = token
	case refreshToken:
		token := jwt.New()
		err := token.Set(jwt.IssuerKey, conf.Iss)
		if err != nil {
			return nil, err
		}

		if conf.Sub && payload.ID != nil {
			err := token.Set(jwt.SubjectKey, fmt.Sprintf("%v", payload.ID))
			if err != nil {
				return nil, err
			}
		}

		currTime := time.Now()
		refreshExp := currTime.Add(time.Duration(conf.RefreshExp) * time.Second).Unix()
		err = token.Set(jwt.ExpirationKey, refreshExp)
		if err != nil {
			return nil, err
		}

		payload, err := renderPayload(string(defaultPayloadTmpl), payload)
		if err != nil {
			return nil, err
		}
		for k, v := range payload {
			err := token.Set(k, v)
			if err != nil {
				return nil, err
			}
		}

		t = token
	}

	return t, err
}

func parsePayload(tmplPath string, payload *core.IssuerPayload) (map[string]interface{}, error) {
	if tmplPath != "" {
		extension := path.Ext(tmplPath)
		if extension == ".tmpl" {
			tmplBytes, err := os.ReadFile(tmplPath)
			if err != nil {
				return nil, err
			}
			return renderPayload(string(tmplBytes), payload)
		}
		return nil, fmt.Errorf("jwt: json type expected, '%s' found", extension)
	} else {
		return renderPayload(string(defaultPayloadTmpl), payload)
	}
}

func renderPayload(tmplStr string, payload *core.IssuerPayload) (map[string]interface{}, error) {
	tmpl, err := txtTmpl.New("payload").Parse(tmplStr)
	if err != nil {
		return nil, err
	}

	bufRawPayload := &buffer.Buffer{}
	err = tmpl.Execute(bufRawPayload, payload)
	if err != nil {
		return nil, err
	}

	strRawPayload := regexp.MustCompile(`\s+`).ReplaceAllString(bufRawPayload.String(), "")
	strRawPayload = regexp.MustCompile(`,}`).ReplaceAllString(strRawPayload, "}")

	p := make(map[string]interface{})
	err = json.Unmarshal([]byte(strRawPayload), &p)
	return p, err
}

// todo: move sing functionality to core
func signToken(signKey core.CryptoKey, token jwt.Token) ([]byte, error) {
	keySet := signKey.GetPrivateSet()

	for it := keySet.Iterate(context.Background()); it.Next(context.Background()); {
		pair := it.Pair()
		key := pair.Value.(jwk.Key)

		if key.KeyUsage() == "sig" {
			var signAlg jwa.SignatureAlgorithm
			if err := signAlg.Accept(key.Algorithm()); err != nil {
				return []byte{}, err
			}
			return jwt.Sign(token, signAlg, key)
		}
	}

	return []byte{}, errors.New("key set don't contain sig key")
}

func attachTokens(c *fiber.Ctx, bearers map[tokenType]tokenResp, tokens map[tokenType][]byte) error {
	jsonBody := make(map[string]interface{})
	if respBody := c.Response().Body(); len(respBody) != 0 {
		if err := json.Unmarshal(respBody, &jsonBody); err != nil {
			return err
		}
	}

	for name, token := range tokens {
		switch bearers[name] {
		case body:
			jsonBody[keyMap[name]["body"]] = string(token)
		case cookie:
			cookie := &fiber.Cookie{
				Name:  keyMap[name]["cookie"],
				Value: string(token),
			}
			c.Cookie(cookie)
		default:
			return fmt.Errorf("unexpected bearer name: %s", bearers[name])
		}
	}
	return c.JSON(jsonBody)
}
