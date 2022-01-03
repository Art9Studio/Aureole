package jwt

import (
	"aureole/internal/configs"
	"aureole/internal/core"
	"aureole/internal/plugins"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"net/http"
	"os"
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

const pluginID = "4844"

type (
	jwtAuthz struct {
		pluginApi     core.PluginAPI
		app           *core.App
		rawConf       *configs.Authz
		conf          *config
		signKey       plugins.CryptoKey
		verifyKeys    map[string]plugins.CryptoKey
		nativeQueries map[string]string
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

func (j *jwtAuthz) Init(appName string, api core.PluginAPI) (err error) {
	j.pluginApi = api
	j.conf, err = initConfig(&j.rawConf.Config)
	if err != nil {
		return err
	}

	j.app, err = j.pluginApi.GetApp(appName)
	if err != nil {
		return err
	}

	j.signKey, err = j.pluginApi.GetCryptoKey(j.conf.SignKey)
	if err != nil {
		return err
	}

	j.verifyKeys = make(map[string]plugins.CryptoKey)
	for _, keyName := range j.conf.VerifyKeys {
		j.verifyKeys[keyName], err = j.pluginApi.GetCryptoKey(keyName)
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

func (j *jwtAuthz) GetMetaData() plugins.Meta {
	return plugins.Meta{
		Type: adapterName,
		Name: j.rawConf.Name,
		ID:   pluginID,
	}
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

func readNativeQueries(path string) (map[string]string, error) {
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
}

func createRoutes(j *jwtAuthz) {
	routes := []*core.Route{
		{
			Method:  http.MethodPost,
			Path:    refreshUrl,
			Handler: refresh(j),
		},
	}
	j.pluginApi.AddAppRoutes(j.app.GetName(), routes)
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
			err := token.Set(jwt.SubjectKey, fmt.Sprintf("%f", payload.ID))
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
			err := token.Set(jwt.SubjectKey, fmt.Sprintf("%f", payload.ID))
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

func parsePayload(tmplPath string, payload *plugins.Payload) (map[string]interface{}, error) {
	if tmplPath != "" {
		return renderPayload(tmplPath, payload)
	} else {
		return defaultPayload(payload)
	}
}

func renderPayload(tmplPath string, payload *plugins.Payload) (map[string]interface{}, error) {
	tmplFile := tmplPath
	baseName := path.Base(tmplFile)
	bufRawPayload := &bytes.Buffer{}

	extension := path.Ext(tmplFile)
	if extension == ".tmpl" {
		/*tmpl := txtTmpl.Must(txtTmpl.New(baseName).Funcs(txtTmpl.FuncMap{
			"NativeQ": payload.NativeQ,
		}).ParseFiles(tmplFile))*/

		tmpl := txtTmpl.Must(txtTmpl.New(baseName).ParseFiles(tmplFile))
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

func defaultPayload(payload *plugins.Payload) (map[string]interface{}, error) {
	p := make(map[string]interface{})
	if payload.ID != nil {
		p["id"] = payload.ID
	}
	return p, nil
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
