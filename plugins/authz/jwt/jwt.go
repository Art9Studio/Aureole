package jwt

import (
	"aureole/internal/configs"
	"aureole/internal/core"
	"aureole/internal/plugins"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-openapi/spec"
	"go.uber.org/zap/buffer"
	"net/http"
	"os"
	"path"
	"regexp"
	txtTmpl "text/template"
	"time"

	gonanoid "github.com/matoous/go-nanoid/v2"

	"github.com/gofiber/fiber/v2"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/mitchellh/mapstructure"
)

const pluginID = "4844"

type (
	jwtAuthz struct {
		pluginAPI     core.PluginAPI
		rawConf       *configs.Authz
		conf          *config
		signKey       plugins.CryptoKey
		verifyKeys    map[string]plugins.CryptoKey
		nativeQueries map[string]string
		swagger       struct {
			Paths       *spec.Paths
			Definitions spec.Definitions
		}
		responses struct {
			Responses   *spec.Responses
			Definitions spec.Definitions
		}
	}

	tokenType  string
	bearerType string
)

const (
	accessToken  tokenType = "access"
	refreshToken tokenType = "refresh"
)

const (
	body   bearerType = "body"
	header bearerType = "header"
	cookie bearerType = "cookie"
	both   bearerType = "both"
)

const defaultPayloadTmpl = `{
  {{ if .ID }}
    "id": {{ .ID }},
  {{ end }}

  {{ if .Username }}
    "username": "{{ .Username }}",
  {{ end }}

  {{ if .Phone }}
    "phone": "{{ .Phone }}",
  {{ end }}

  {{ if .Email }}
    "email": "{{ .Email }}",
  {{ end }}
}`

var keyMap = map[string]map[string]string{
	"access": {
		"header": "access",
		"cookie": "access_token",
	},
	"refresh": {
		"body":   "refresh",
		"cookie": "refresh_token",
	},
}

var (
	//go:embed swagger.json
	swaggerJson []byte
	//go:embed responses.json
	responsesJson []byte
)

func (j *jwtAuthz) Init(api core.PluginAPI) (err error) {
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

	j.verifyKeys = make(map[string]plugins.CryptoKey)
	for _, keyName := range j.conf.VerifyKeys {
		j.verifyKeys[keyName], ok = j.pluginAPI.GetCryptoKey(keyName)
		if !ok {
			return fmt.Errorf("cannot get crypto key named %s", j.conf.SignKey)
		}
	}

	/*if j.conf.NativeQueries != "" {
		if j.nativeQueries, err = readNativeQueries(j.conf.NativeQueries); err != nil {
			return err
		}
	}*/

	err = assembleResponses(j)
	if err != nil {
		return err
	}
	err = assembleSwagger(j)
	if err != nil {
		return err
	}

	createRoutes(j)
	return err
}

func assembleSwagger(j *jwtAuthz) error {
	err := json.Unmarshal(swaggerJson, &j.swagger)
	if err != nil {
		return fmt.Errorf("jwt authz: cannot marshal swagger docs: %v", err)
	}

	handler := j.swagger.Paths.Paths["/jwt/refresh"].Post
	if j.conf.RefreshBearer == body {
		handler.Parameters = handler.Parameters[:1]
	} else if j.conf.RefreshBearer == cookie {
		handler.Consumes = nil
		handler.Produces = nil
		handler.Parameters = handler.Parameters[1:]
	}

	var resp spec.Responses
	bytes, err := j.responses.Responses.MarshalJSON()
	if err != nil {
		return err
	}
	err = resp.UnmarshalJSON(bytes)
	if err != nil {
		return err
	}

	errResp := handler.Responses.StatusCodeResponses[400]
	handler.Responses = &resp
	handler.Responses.StatusCodeResponses[400] = errResp
	okResp := handler.Responses.StatusCodeResponses[200]
	okResp.Description = "Successfully refresh JWT and returns new refresh JWT and old access JWT"
	handler.Responses.StatusCodeResponses[200] = okResp

	return nil
}

func assembleResponses(j *jwtAuthz) error {
	err := json.Unmarshal(responsesJson, &j.responses)
	if err != nil {
		return fmt.Errorf("jwt authz: cannot marshal jwt responses docs: %v", err)
	}

	okResponse := j.responses.Responses.ResponsesProps.StatusCodeResponses[200]

	if j.conf.AccessBearer == both {
		okResponse.Headers["Set-Cookie"] = okResponse.Headers["Access-Set-Cookie"]
	} else if j.conf.AccessBearer == cookie {
		okResponse.Headers["Set-Cookie"] = okResponse.Headers["Access-Set-Cookie"]
		delete(okResponse.Headers, "access")
	}

	if j.conf.RefreshBearer == both {
		_, ok := okResponse.Headers["Set-Cookie"]
		if ok {
			okResponse.Headers["Set-Cookie"] = okResponse.Headers["Both-Set-Cookie"]
		} else {
			okResponse.Headers["Set-Cookie"] = okResponse.Headers["Refresh-Set-Cookie"]
		}
	} else if j.conf.RefreshBearer == cookie {
		okResponse.Schema = nil
		_, ok := okResponse.Headers["Set-Cookie"]
		if ok {
			okResponse.Headers["Set-Cookie"] = okResponse.Headers["Both-Set-Cookie"]
		} else {
			okResponse.Headers["Set-Cookie"] = okResponse.Headers["Refresh-Set-Cookie"]
		}
	}

	delete(okResponse.Headers, "Access-Set-Cookie")
	delete(okResponse.Headers, "Refresh-Set-Cookie")
	delete(okResponse.Headers, "Both-Set-Cookie")

	j.responses.Responses.ResponsesProps.StatusCodeResponses[200] = okResponse
	return nil
}

func (j *jwtAuthz) GetMetaData() plugins.Meta {
	return plugins.Meta{
		Type: adapterName,
		ID:   pluginID,
	}
}

func (j *jwtAuthz) GetHandlersSpec() (*spec.Paths, spec.Definitions) {
	return j.swagger.Paths, j.swagger.Definitions
}

func (j *jwtAuthz) GetResponseData() (*spec.Responses, spec.Definitions) {
	return j.responses.Responses, j.responses.Definitions
}

func (j *jwtAuthz) GetNativeQueries() map[string]string {
	return j.nativeQueries
}

func (j *jwtAuthz) Authorize(c *fiber.Ctx, payload *plugins.Payload) error {
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

	bearers := map[string]bearerType{
		"access":  j.conf.AccessBearer,
		"refresh": j.conf.RefreshBearer,
	}
	tokens := map[string][]byte{
		"access":  signedAccessT,
		"refresh": signedRefreshT,
	}
	if err := attachTokens(c, bearers, keyMap, tokens); err != nil {
		return core.SendError(c, fiber.StatusInternalServerError, err.Error())
	}

	return nil
}

func initConfig(rawConf *configs.RawConfig) (*config, error) {
	adapterConf := &config{}
	if err := mapstructure.Decode(rawConf, adapterConf); err != nil {
		return nil, err
	}
	adapterConf.setDefaults()

	return adapterConf, nil
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

func createRoutes(j *jwtAuthz) {
	routes := []*core.Route{
		{
			Method:  http.MethodPost,
			Path:    refreshUrl,
			Handler: refresh(j),
		},
	}
	j.pluginAPI.AddAppRoutes(routes)
}

func newToken(tokenType tokenType, conf *config, payload *plugins.Payload) (t jwt.Token, err error) {
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

		payload, err := renderPayload(defaultPayloadTmpl, payload)
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

func parsePayload(tmplPath string, payload *plugins.Payload) (map[string]interface{}, error) {
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
		return renderPayload(defaultPayloadTmpl, payload)
	}
}

func renderPayload(tmplStr string, payload *plugins.Payload) (map[string]interface{}, error) {
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

func signToken(signKey plugins.CryptoKey, token jwt.Token) ([]byte, error) {
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

func attachTokens(c *fiber.Ctx, bearers map[string]bearerType, keyMap map[string]map[string]string, tokens map[string][]byte) error {
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
		case header:
			c.Set(keyMap[name]["header"], string(token))
		case cookie:
			cookie := &fiber.Cookie{
				Name:  keyMap[name]["cookie"],
				Value: string(token),
			}
			c.Cookie(cookie)
		case both:
			jsonBody[keyMap[name]["header"]] = string(token)

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
