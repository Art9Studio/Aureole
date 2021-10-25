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
	"io/ioutil"
	"path"
	"regexp"
	txtTmpl "text/template"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v3"
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
		"body":   "refresh",
		"cookie": "refresh_token",
	},
}

func (j *jwtAuthz) Init(appName string) (err error) {
	j.appName = appName
	j.rawConf.PathPrefix = "/" + AdapterName

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

	if j.conf.NativeQueries != "" {
		if j.nativeQueries, err = readNativeQueries(j.conf.NativeQueries); err != nil {
			return err
		}
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
	authn.Repository.PluginApi.Router.AddAppRoutes(j.appName, routes)
}

func (j *jwtAuthz) GetNativeQueries() map[string]string {
	return j.nativeQueries
}

func (j *jwtAuthz) Authorize(c *fiber.Ctx, payload *authzTypes.Payload) error {
	accessT, err := newToken(AccessToken, j.conf, payload)
	if err != nil {
		return sendError(c, fiber.StatusInternalServerError, err.Error())
	}
	refreshT, err := newToken(RefreshToken, j.conf, payload)
	if err != nil {
		return sendError(c, fiber.StatusInternalServerError, err.Error())
	}

	signedAccessT, err := signToken(j.signKey, accessT)
	if err != nil {
		return sendError(c, fiber.StatusInternalServerError, err.Error())
	}
	signedRefreshT, err := signToken(j.signKey, refreshT)
	if err != nil {
		return sendError(c, fiber.StatusInternalServerError, err.Error())
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
		return sendError(c, fiber.StatusInternalServerError, err.Error())
	}

	return nil
}

func newToken(tokenType tokenType, conf *config, payload *authzTypes.Payload) (t jwt.Token, err error) {
	switch tokenType {
	case AccessToken:
		token := jwt.New()
		// todo: think about multiple errors handling
		err := token.Set(jwt.IssuerKey, conf.Iss)
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
		err = token.Set(jwt.JwtIDKey, conf.Jti)
		if err != nil {
			return nil, err
		}

		if conf.Sub {
			err := token.Set(jwt.SubjectKey, fmt.Sprintf("%f", payload.Id))
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
	case RefreshToken:
		token := jwt.New()
		err := token.Set(jwt.IssuerKey, conf.Iss)
		if err != nil {
			return nil, err
		}

		if conf.Sub {
			err := token.Set(jwt.SubjectKey, fmt.Sprintf("%f", payload.Id))
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

		payload, err := defaultPayload(payload)
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

func parsePayload(tmplPath string, payload *authzTypes.Payload) (map[string]interface{}, error) {
	if tmplPath != "" {
		return renderPayload(tmplPath, payload)
	} else {
		return defaultPayload(payload)
	}
}

func renderPayload(tmplPath string, payload *authzTypes.Payload) (map[string]interface{}, error) {
	tmplFile := tmplPath
	baseName := path.Base(tmplFile)
	bufRawPayload := &bytes.Buffer{}

	extension := path.Ext(tmplFile)
	if extension == ".tmpl" {
		tmpl := txtTmpl.Must(txtTmpl.New(baseName).Funcs(txtTmpl.FuncMap{
			"NativeQ": payload.NativeQ,
		}).ParseFiles(tmplFile))
		if err := tmpl.Execute(bufRawPayload, payload); err != nil {
			return nil, err
		}

		strRawPayload := regexp.MustCompile(`\s+`).ReplaceAllString(bufRawPayload.String(), "")
		strRawPayload = regexp.MustCompile(`,}`).ReplaceAllString(strRawPayload, "}")

		p := make(map[string]interface{})
		if err := json.Unmarshal([]byte(strRawPayload), &p); err != nil {
			return nil, err
		}

		return p, nil
	} else {
		return nil, fmt.Errorf("jwt: json type expected, '%s' found", extension)
	}
}

func defaultPayload(payload *authzTypes.Payload) (map[string]interface{}, error) {
	p := make(map[string]interface{})
	if payload.Id != nil {
		p["id"] = payload.Id
	}
	return p, nil
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
	jsonBody := make(map[string]interface{})
	if respBody := c.Response().Body(); len(respBody) != 0 {
		if err := json.Unmarshal(respBody, &jsonBody); err != nil {
			return err
		}
	}

	for name, token := range tokens {
		switch bearers[name] {
		case Body:
			jsonBody[keyMap[name]["body"]] = string(token)
		case Header:
			c.Set(keyMap[name]["header"], string(token))
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
