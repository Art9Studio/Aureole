package apple

import (
	"aureole/internal/configs"
	"aureole/internal/identity"
	"aureole/internal/plugins"
	authnT "aureole/internal/plugins/authn/types"
	authzT "aureole/internal/plugins/authz/types"
	"aureole/internal/plugins/core"
	cKeyT "aureole/internal/plugins/cryptokey/types"
	"aureole/internal/router"
	app "aureole/internal/state/interface"
	"context"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/mitchellh/mapstructure"
	"path"
	"time"
)

const PluginID = "5771"

type apple struct {
	pluginApi  core.PluginAPI
	app        app.AppState
	rawConf    *configs.Authn
	conf       *config
	manager    identity.ManagerI
	secretKey  cKeyT.CryptoKey
	publicKey  cKeyT.CryptoKey
	provider   *Config
	authorizer authzT.Authorizer
}

func (a *apple) Init(appName string, api core.PluginAPI) (err error) {
	a.pluginApi = api
	a.conf, err = initConfig(&a.rawConf.Config)
	if err != nil {
		return err
	}

	a.app, err = a.pluginApi.GetApp(appName)
	if err != nil {
		return fmt.Errorf("app named '%s' is not declared", appName)
	}

	a.manager, err = a.app.GetIdentityManager()
	if err != nil {
		fmt.Printf("manager for app '%s' is not declared", appName)
	}

	a.secretKey, err = a.pluginApi.GetCryptoKey(a.conf.SecretKey)
	if err != nil {
		return fmt.Errorf("crypto key named '%s' is not declared", a.conf.SecretKey)
	}

	a.publicKey, err = a.pluginApi.GetCryptoKey(a.conf.PublicKey)
	if err != nil {
		return fmt.Errorf("crypto key named '%s' is not declared", a.conf.PublicKey)
	}

	a.authorizer, err = a.app.GetAuthorizer()
	if err != nil {
		return fmt.Errorf("authorizer named for app '%s' is not declared", appName)
	}

	if err := initProvider(a); err != nil {
		return err
	}
	createRoutes(a)
	return nil
}

func (*apple) GetMetaData() plugins.Meta {
	return plugins.Meta{
		Type: AdapterName,
		ID:   PluginID,
	}
}

func (a *apple) Login() authnT.AuthFunc {
	return func(c fiber.Ctx) (*identity.Credential, fiber.Map, error) {
		input := struct {
			State string
			Code  string
		}{}
		if err := c.BodyParser(&input); err != nil {
			return nil, nil, err
		}
		if input.State != "state" {
			return nil, nil, errors.New("invalid state")
		}
		if input.Code == "" {
			return nil, nil, errors.New("code not found")
		}

		jwtT, err := getJwt(a, input.Code)
		if err != nil {
			return nil, nil, err
		}

		email, ok := jwtT.Get("email")
		if !ok {
			return nil, nil, errors.New("can't get 'email' from token")
		}
		socialId, ok := jwtT.Get("sub")
		if !ok {
			return nil, nil, errors.New("can't get 'social_id' from token")
		}
		userData, err := jwtT.AsMap(context.Background())
		if err != nil {
			return nil, nil, err
		}

		if ok, err := a.app.Filter(convertUserData(userData), a.rawConf.Filter); err != nil {
			return nil, nil, err
		} else if !ok {
			return nil, nil, errors.New("input data doesn't pass filters")
		}

		return &identity.Credential{
				Name:  identity.Email,
				Value: email.(string),
			},
			fiber.Map{
				identity.Email:         email,
				identity.AuthnProvider: AdapterName,
				identity.SocialID:      socialId,
				identity.UserData:      userData,
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
	redirectUrl, err := a.app.GetUrl()
	if err != nil {
		return err
	}

	redirectUrl.Path = path.Clean(redirectUrl.Path + a.conf.RedirectUri)
	a.provider = &Config{
		ClientId: a.conf.ClientId,
		TeamId:   a.conf.TeamId,
		KeyId:    a.conf.KeyId,
		Endpoint: Endpoint{
			AuthUrl:  AuthUrl,
			TokenUrl: TokenUrl,
		},
		RedirectUrl: redirectUrl.String(),
		Scopes:      a.conf.Scopes,
	}

	return createSecret(a.provider, a.secretKey)
}

func createSecret(p *Config, key cKeyT.CryptoKey) error {
	t := jwt.New()
	claims := []struct {
		Name string
		Val  interface{}
	}{
		{Name: jwt.IssuerKey, Val: p.TeamId},
		{Name: jwt.AudienceKey, Val: "https://appleid.apple.com"},
		{Name: jwt.SubjectKey, Val: p.ClientId},
		{Name: jwt.IssuedAtKey, Val: time.Now().Unix()},
		{Name: jwt.ExpirationKey, Val: time.Now().Add(time.Hour * 24 * 180).Unix()},
		{Name: jwk.KeyIDKey, Val: p.KeyId},
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

	p.ClientSecret = string(signedT)
	return nil
}

func signToken(signKey cKeyT.CryptoKey, token jwt.Token) ([]byte, error) {
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
	routes := []*router.Route{
		{
			Method:  router.MethodGET,
			Path:    a.conf.PathPrefix,
			Handler: GetAuthCode(a),
		},
	}
	router.GetRouter().AddAppRoutes(a.app.GetName(), routes)
}
