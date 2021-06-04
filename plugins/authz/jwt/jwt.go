package jwt

import (
	"aureole/internal/configs"
	"aureole/internal/plugins/authn"
	"aureole/internal/plugins/authz"
	authzTypes "aureole/internal/plugins/authz/types"
	ckeyTypes "aureole/internal/plugins/cryptokey/types"
	_interface "aureole/internal/router/interface"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"path"
	txtTmpl "text/template"
	"time"
)

type jwtAuthz struct {
	appName       string
	rawConf       *configs.Authz
	conf          *config
	signKey       ckeyTypes.CryptoKey
	verifyKeys    map[string]ckeyTypes.CryptoKey
	nativeQueries map[string]string
}

type (
	tokenType  string
	bearerType string
)

const (
	AccessToken  tokenType = "access"
	RefreshToken tokenType = "refresh"
)

const (
	Body   bearerType = "body"
	Header bearerType = "header"
	Cookie bearerType = "cookie"
	Both   bearerType = "both"
)

var keyMap = map[string]map[string]string{
	"access": {
		"header": "access",
		"cookie": "access_token",
	},
	"refresh": {
		"header": "refresh",
		"cookie": "refresh_token",
	},
}

func (j *jwtAuthz) Init(appName string) (err error) {
	j.appName = appName

	j.conf, err = initConfig(&j.rawConf.Config)
	if err != nil {
		return err
	}

	pluginsApi := authz.Repository.PluginApi
	j.signKey, err = pluginsApi.Project.GetCryptoKey(j.conf.SignKey)
	if err != nil {
		return err
	}

	j.verifyKeys = make(map[string]ckeyTypes.CryptoKey)
	for _, keyName := range j.conf.VerifyKeys {
		j.verifyKeys[keyName], err = pluginsApi.Project.GetCryptoKey(keyName)
		if err != nil {
			return err
		}
	}

	if j.nativeQueries, err = readNativeQueries(j.conf.NativeQueries); err != nil {
		return err
	}

	createRoutes(j)
	return err
}

func initConfig(rawConf *configs.RawConfig) (*config, error) {
	adapterConf := &config{}
	if err := mapstructure.Decode(rawConf, adapterConf); err != nil {
		return nil, err
	}
	adapterConf.setDefaults()

	return adapterConf, nil
}

func readNativeQueries(path string) (map[string]string, error) {
	q := map[string]string{}

	f, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(f, &q)
	if err != nil {
		return nil, err
	}

	return q, nil
}

func createRoutes(j *jwtAuthz) {
	routes := []*_interface.Route{
		{
			Method:  "POST",
			Path:    j.rawConf.PathPrefix + j.conf.RefreshUrl,
			Handler: Refresh(j),
		},
	}
	authn.Repository.PluginApi.Router.Add(j.appName, routes)
}

func (j *jwtAuthz) GetNativeQueries() map[string]string {
	return j.nativeQueries
}

func (j *jwtAuthz) Authorize(fiberCtx *fiber.Ctx, authzCtx *authzTypes.Context) error {
	accessT, err := newToken(AccessToken, j.conf, authzCtx)
	if err != nil {
		return sendError(fiberCtx, fiber.StatusInternalServerError, err.Error())
	}
	refreshT, err := newToken(RefreshToken, j.conf, authzCtx)
	if err != nil {
		return sendError(fiberCtx, fiber.StatusInternalServerError, err.Error())
	}

	signedAccessT, err := signToken(j.signKey, accessT)
	if err != nil {
		return sendError(fiberCtx, fiber.StatusInternalServerError, err.Error())
	}
	signedRefreshT, err := signToken(j.signKey, refreshT)
	if err != nil {
		return sendError(fiberCtx, fiber.StatusInternalServerError, err.Error())
	}

	bearers := map[string]bearerType{
		"access":  j.conf.AccessBearer,
		"refresh": j.conf.RefreshBearer,
	}
	tokens := map[string][]byte{
		"access":  signedAccessT,
		"refresh": signedRefreshT,
	}
	if err := attachTokens(fiberCtx, bearers, keyMap, tokens); err != nil {
		return sendError(fiberCtx, fiber.StatusInternalServerError, err.Error())
	}

	return nil
}

func newToken(tokenType tokenType, conf *config, authzCtx *authzTypes.Context) (t jwt.Token, err error) {
	switch tokenType {
	case AccessToken:
		token := jwt.New()
		// todo: think about multiple errors handling
		token.Set(jwt.IssuerKey, conf.Iss)
		token.Set(jwt.AudienceKey, conf.Aud)
		token.Set(jwt.NotBeforeKey, conf.Nbf)
		token.Set(jwt.JwtIDKey, conf.Jti)

		if conf.Sub {
			token.Set(jwt.SubjectKey, authzCtx.Id)
		}

		currTime := time.Now()
		nbf := currTime.Add(time.Duration(conf.Nbf) * time.Second).Unix()
		token.Set(jwt.NotBeforeKey, nbf)

		if conf.Iat {
			token.Set(jwt.IssuedAtKey, currTime.Unix())
		}

		exp := currTime.Add(time.Duration(conf.AccessExp) * time.Second).Unix()
		token.Set(jwt.ExpirationKey, exp)

		payload, err := getPayload(conf.Payload, authzCtx)
		if err != nil {
			return nil, err
		}
		for k, v := range payload {
			token.Set(k, v)
		}

		t = token
	case RefreshToken:
		token := jwt.New()
		token.Set(jwt.IssuerKey, conf.Iss)

		if conf.Sub {
			token.Set(jwt.SubjectKey, authzCtx.Id)
		}

		currTime := time.Now()
		refreshExp := currTime.Add(time.Duration(conf.RefreshExp) * time.Second).Unix()
		token.Set(jwt.ExpirationKey, refreshExp)

		payload, err := defaultPayload(authzCtx)
		if err != nil {
			return nil, err
		}
		for k, v := range payload {
			token.Set(k, v)
		}

		t = token
	}

	return t, err
}

func getPayload(filePath string, authzCtx *authzTypes.Context) (map[string]interface{}, error) {
	if filePath != "" {
		return parsePayload(filePath, authzCtx)
	} else {
		return defaultPayload(authzCtx)
	}
}

func parsePayload(filePath string, authzCtx *authzTypes.Context) (map[string]interface{}, error) {
	tmplFile := filePath
	baseName := path.Base(tmplFile)
	extension := path.Ext(tmplFile)
	rawPayload := &bytes.Buffer{}

	if extension == ".json" {
		tmpl := txtTmpl.Must(txtTmpl.New(baseName).Funcs(txtTmpl.FuncMap{
			"NativeQ": authzCtx.NativeQ,
		}).ParseFiles(tmplFile))
		if err := tmpl.Execute(rawPayload, authzCtx); err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("jwt: json type expected, '%s' found", extension)
	}

	payload := make(map[string]interface{})
	if err := json.Unmarshal(rawPayload.Bytes(), &payload); err != nil {
		return nil, err
	}

	return payload, nil
}

func defaultPayload(authzCtx *authzTypes.Context) (map[string]interface{}, error) {
	payload := make(map[string]interface{})

	if authzCtx.Id != nil {
		payload["id"] = authzCtx.Id
	}

	return payload, nil
}

func signToken(signKey ckeyTypes.CryptoKey, token jwt.Token) ([]byte, error) {
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
	jsonBody := map[string]string{}

	for name, token := range tokens {
		switch bearers[name] {
		case Body, Header:
			jsonBody[keyMap[name]["header"]] = string(token)
		case Cookie:
			cookie := &fiber.Cookie{
				Name:  keyMap[name]["cookie"],
				Value: string(token),
			}
			c.Cookie(cookie)
		case Both:
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

func sendError(c *fiber.Ctx, statusCode int, message string) error {
	return c.Status(statusCode).JSON(&fiber.Map{
		"success": false,
		"message": message,
	})
}
