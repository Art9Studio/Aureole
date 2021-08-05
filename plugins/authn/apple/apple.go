package apple

import (
	"aureole/internal/collections"
	"aureole/internal/configs"
	"aureole/internal/identity"
	"aureole/internal/plugins/authn"
	authzT "aureole/internal/plugins/authz/types"
	cKeyT "aureole/internal/plugins/cryptokey/types"
	storageT "aureole/internal/plugins/storage/types"
	"aureole/internal/router/interface"
	"context"
	"errors"
	"fmt"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/mitchellh/mapstructure"
	"net/url"
	"path"
	"time"
)

const Provider = "apple"

type apple struct {
	appName    string
	appUrl     *url.URL
	rawConf    *configs.Authn
	conf       *config
	identity   *identity.Identity
	coll       *collections.Collection
	storage    storageT.Storage
	secretKey  cKeyT.CryptoKey
	publicKey  cKeyT.CryptoKey
	provider   *Config
	authorizer authzT.Authorizer
}

func (a *apple) Init(appName string, appUrl *url.URL) (err error) {
	a.appName = appName
	a.appUrl = appUrl

	a.conf, err = initConfig(&a.rawConf.Config)
	if err != nil {
		return err
	}

	pluginApi := authn.Repository.PluginApi
	a.identity, err = pluginApi.Project.GetIdentity(appName)
	if err != nil {
		return fmt.Errorf("identity in app '%s' is not declared", appName)
	}

	a.coll, err = pluginApi.Project.GetCollection(a.conf.Coll)
	if err != nil {
		return fmt.Errorf("collection named '%s' is not declared", a.conf.Coll)
	}

	a.storage, err = pluginApi.Project.GetStorage(a.conf.Storage)
	if err != nil {
		return fmt.Errorf("storage named '%s' is not declared", a.conf.Storage)
	}

	a.secretKey, err = pluginApi.Project.GetCryptoKey(a.conf.SecretKey)
	if err != nil {
		return fmt.Errorf("crypto key named '%s' is not declared", a.conf.SecretKey)
	}

	a.publicKey, err = pluginApi.Project.GetCryptoKey(a.conf.PublicKey)
	if err != nil {
		return fmt.Errorf("crypto key named '%s' is not declared", a.conf.PublicKey)
	}

	a.authorizer, err = pluginApi.Project.GetAuthorizer(a.rawConf.AuthzName, appName)
	if err != nil {
		return fmt.Errorf("authorizer named '%s' is not declared", a.rawConf.AuthzName)
	}

	redirectUrl := a.appUrl
	redirectUrl.Path = path.Clean(redirectUrl.Path + a.rawConf.PathPrefix + a.conf.RedirectUrl)
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

	if err := createSecret(a.provider, a.secretKey); err != nil {
		return err
	}
	createRoutes(a)
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
	routes := []*_interface.Route{
		{
			Method:  "GET",
			Path:    a.rawConf.PathPrefix,
			Handler: GetAuthCode(a),
		},
		{
			Method:  "POST",
			Path:    a.rawConf.PathPrefix + a.conf.RedirectUrl,
			Handler: Login(a),
		},
	}
	authn.Repository.PluginApi.Router.AddAppRoutes(a.appName, routes)
}
