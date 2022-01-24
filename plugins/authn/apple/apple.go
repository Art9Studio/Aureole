package apple

import (
	"aureole/internal/configs"
	"aureole/internal/core"
	"aureole/internal/plugins"
	"context"
	"errors"
	"fmt"
	"net/http"
	"path"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/mitchellh/mapstructure"
)

const pluginID = "5771"

type apple struct {
	pluginAPI core.PluginAPI
	rawConf   *configs.Authn
	conf      *config
	secretKey plugins.CryptoKey
	publicKey plugins.CryptoKey
	provider  *providerConfig
}

func (a *apple) Init(api core.PluginAPI) (err error) {
	a.pluginAPI = api
	a.conf, err = initConfig(&a.rawConf.Config)
	if err != nil {
		return err
	}

	var ok bool
	a.secretKey, ok = a.pluginAPI.GetCryptoKey(a.conf.SecretKey)
	if !ok {
		return fmt.Errorf("crypto key named '%s' is not declared", a.conf.SecretKey)
	}

	a.publicKey, ok = a.pluginAPI.GetCryptoKey(a.conf.PublicKey)
	if !ok {
		return fmt.Errorf("crypto key named '%s' is not declared", a.conf.PublicKey)
	}

	if err := initProvider(a); err != nil {
		return err
	}
	createRoutes(a)
	return nil
}

func (*apple) GetMetaData() plugins.Meta {
	return plugins.Meta{
		Type: adapterName,
		ID:   pluginID,
	}
}

func (a *apple) GetLoginHandler() (string, func() plugins.AuthNLoginFunc) {
	return http.MethodPost, a.login
}

func (a *apple) login() plugins.AuthNLoginFunc {
	return func(c fiber.Ctx) (*plugins.AuthNResult, error) {
		input := struct {
			State string
			Code  string
		}{}
		if err := c.BodyParser(&input); err != nil {
			return nil, err
		}
		if input.State != "state" {
			return nil, errors.New("invalid state")
		}
		if input.Code == "" {
			return nil, errors.New("code not found")
		}

		var email string
		jwtT, err := getJwt(a, input.Code)
		if err != nil {
			return nil, err
		}
		err = a.pluginAPI.GetFromJWT(jwtT, "email", &email)
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

		ok, err := a.pluginAPI.Filter(convertUserData(userData), a.rawConf.Filter)
		if err != nil {
			return nil, err
		} else if !ok {
			return nil, errors.New("input data doesn't pass filters")
		}

		return &plugins.AuthNResult{
			Cred: &plugins.Credential{
				Name:  plugins.Email,
				Value: email,
			},
			Identity: &plugins.Identity{
				Email:         &email,
				EmailVerified: true,
				Additional:    map[string]interface{}{"social_provider_data": userData},
			},
			Provider: "social_provider$" + adapterName,
		}, nil
	}
}

func initConfig(rawConf *configs.RawConfig) (*config, error) {
	adapterConf := &config{}
	if err := mapstructure.Decode(rawConf, adapterConf); err != nil {
		return nil, err
	}
	adapterConf.setDefaults()
	return adapterConf, nil
}

func initProvider(a *apple) error {
	url := a.pluginAPI.GetAppUrl()
	url.Path = path.Clean(url.Path + pathPrefix + redirectUrl)
	a.provider = &providerConfig{
		clientId: a.conf.ClientId,
		teamId:   a.conf.TeamId,
		keyId:    a.conf.KeyId,
		endpoint: endpoint{
			authUrl:  authUrl,
			tokenUrl: tokenUrl,
		},
		redirectUrl: url.String(),
		scopes:      a.conf.Scopes,
	}
	return createSecret(a.provider, a.secretKey)
}

func createSecret(p *providerConfig, key plugins.CryptoKey) error {
	t := jwt.New()
	claims := []struct {
		Name string
		Val  interface{}
	}{
		{Name: jwt.IssuerKey, Val: p.teamId},
		{Name: jwt.AudienceKey, Val: "https://appleid.apple.com"},
		{Name: jwt.SubjectKey, Val: p.clientId},
		{Name: jwt.IssuedAtKey, Val: time.Now().Unix()},
		{Name: jwt.ExpirationKey, Val: time.Now().Add(time.Hour * 24 * 180).Unix()},
		{Name: jwk.KeyIDKey, Val: p.keyId},
	}

	for _, claim := range claims {
		if err := t.Set(claim.Name, claim.Val); err != nil {
			return err
		}
	}

	signedT, err := signToken(key, t)
	if err != nil {
		return err
	}

	p.clientSecret = string(signedT)
	return nil
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

func createRoutes(a *apple) {
	routes := []*core.Route{
		{
			Method:  http.MethodGet,
			Path:    pathPrefix,
			Handler: getAuthCode(a),
		},
	}
	a.pluginAPI.AddAppRoutes(routes)
}
