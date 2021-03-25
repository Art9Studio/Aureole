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
	"path"
	txtTmpl "text/template"
	"time"
)

type jwtAuthz struct {
	appName    string
	rawConf    *configs.Authz
	conf       *config
	signKey    ckeyTypes.CryptoKey
	verifyKeys map[string]ckeyTypes.CryptoKey
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
	Header bearerType = "header"
	Cookie bearerType = "cookie"
	Both   bearerType = "both"
)

var names = map[string]map[string]string{
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

func (j *jwtAuthz) Authorize(fiberCtx *fiber.Ctx, authzCtx *authzTypes.Context) error {
	accessT, err := newToken(AccessToken, j.conf, authzCtx)
	if err != nil {
		return fiberCtx.Status(fiber.StatusInternalServerError).JSON(&fiber.Map{
			"success": false,
			"message": err,
		})
	}
	refreshT, err := newToken(RefreshToken, j.conf, authzCtx)
	if err != nil {
		return fiberCtx.Status(fiber.StatusInternalServerError).JSON(&fiber.Map{
			"success": false,
			"message": err,
		})
	}

	signedAccessT, err := signToken(j.signKey, accessT)
	if err != nil {
		return fiberCtx.Status(fiber.StatusInternalServerError).JSON(&fiber.Map{
			"success": false,
			"message": err,
		})
	}
	signedRefreshT, err := signToken(j.signKey, refreshT)
	if err != nil {
		return fiberCtx.Status(fiber.StatusInternalServerError).JSON(&fiber.Map{
			"success": false,
			"message": err,
		})
	}

	if err := sendToken(fiberCtx, j.conf.AccessBearer, names["access"], signedAccessT); err != nil {
		return fiberCtx.Status(fiber.StatusInternalServerError).JSON(&fiber.Map{
			"success": false,
			"message": err,
		})
	}
	if err := sendToken(fiberCtx, j.conf.RefreshBearer, names["refresh"], signedRefreshT); err != nil {
		return fiberCtx.Status(fiber.StatusInternalServerError).JSON(&fiber.Map{
			"success": false,
			"message": err,
		})
	}

	return nil
}

func newToken(tokenType tokenType, conf *config, authzCtx *authzTypes.Context) (jwt.Token, error) {
	switch tokenType {
	case AccessToken:
		token := jwt.New()
		// todo: think about multiple errors handling
		token.Set(jwt.IssuerKey, conf.Iss)
		token.Set(jwt.AudienceKey, conf.Aud)
		token.Set(jwt.NotBeforeKey, conf.Nbf)
		token.Set(jwt.JwtIDKey, conf.Jti)

		if conf.Sub {
			token.Set(jwt.SubjectKey, string(authzCtx.UserId))
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

		return token, nil
	case RefreshToken:
		token := jwt.New()
		token.Set(jwt.IssuerKey, conf.Iss)

		if conf.Sub {
			token.Set(jwt.SubjectKey, string(authzCtx.UserId))
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

		return token, nil
	default:
		return nil, fmt.Errorf("unexpected token type: %s", tokenType)
	}
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
		tmpl := txtTmpl.Must(txtTmpl.New(baseName).ParseFiles(tmplFile))
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
	payload["user_id"] = authzCtx.UserId

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

func sendToken(fiberCtx *fiber.Ctx, bearer bearerType, names map[string]string, token []byte) error {
	switch bearer {
	case Header:
		return fiberCtx.JSON(&fiber.Map{names["header"]: string(token)})
	case Cookie:
		cookie := &fiber.Cookie{
			Name:  names["cookie"],
			Value: string(token),
		}
		fiberCtx.Cookie(cookie)
	case Both:
		if err := fiberCtx.JSON(&fiber.Map{names["header"]: string(token)}); err != nil {
			return err
		}

		cookie := &fiber.Cookie{
			Name:  names["cookie"],
			Value: string(token),
		}
		fiberCtx.Cookie(cookie)
	default:
		return fmt.Errorf("unexpected bearer name: %s", bearer)
	}

	return nil
}
