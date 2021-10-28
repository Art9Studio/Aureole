package apple

import (
	"aureole/internal/configs"
	"aureole/internal/identity"
	"aureole/internal/plugins/authn"
	authzT "aureole/internal/plugins/authz/types"
	cKeyT "aureole/internal/plugins/cryptokey/types"
	storageT "aureole/internal/plugins/storage/types"
	router "aureole/internal/router/interface"
	app "aureole/internal/state/interface"
	"context"
	"errors"
	"fmt"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/mitchellh/mapstructure"
	"path"
	"time"
)

const Provider = "apple"

type apple struct {
	app     app.AppState
	rawConf *configs.Authn
	conf    *config
	//coll       *collections.Collection
	identity   *identity.Identity
	storage    storageT.Storage
	secretKey  cKeyT.CryptoKey
	publicKey  cKeyT.CryptoKey
	provider   *Config
	authorizer authzT.Authorizer
}

func (a *apple) Init(app app.AppState) (err error) {
	a.app = app
	a.rawConf.PathPrefix = "/oauth2/" + AdapterName

	a.conf, err = initConfig(&a.rawConf.Config)
	if err != nil {
		return err
	}

	pluginApi := authn.Repository.PluginApi
	a.identity, err = app.GetIdentity()
	if err != nil {
		return fmt.Errorf("identity for app '%s' is not declared", app.GetName())
	}

	/*a.coll, err = pluginApi.Project.GetCollection(a.conf.Coll)
	if err != nil {
		return fmt.Errorf("collection named '%s' is not declared", a.conf.Coll)
	}

	a.storage, err = pluginApi.Project.GetStorage(a.conf.Storage)
	if err != nil {
		return fmt.Errorf("storage named '%s' is not declared", a.conf.Storage)
	}*/

	a.secretKey, err = pluginApi.Project.GetCryptoKey(a.conf.SecretKey)
	if err != nil {
		return fmt.Errorf("crypto key named '%s' is not declared", a.conf.SecretKey)
	}

	a.publicKey, err = pluginApi.Project.GetCryptoKey(a.conf.PublicKey)
	if err != nil {
		return fmt.Errorf("crypto key named '%s' is not declared", a.conf.PublicKey)
	}

	a.authorizer, err = a.app.GetAuthorizer(a.rawConf.AuthzName)
	if err != nil {
		return fmt.Errorf("authorizer named '%s' is not declared", a.rawConf.AuthzName)
	}

	if err := initProvider(a); err != nil {
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

func initProvider(a *apple) error {
	redirectUrl, err := a.app.GetUrl()
	if err != nil {
		return err
	}

	redirectUrl.Path = path.Clean(redirectUrl.Path + a.rawConf.PathPrefix + a.conf.RedirectUri)
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
			Method:  "GET",
			Path:    a.rawConf.PathPrefix,
			Handler: GetAuthCode(a),
		},
		{
			Method:  "POST",
			Path:    a.rawConf.PathPrefix + a.conf.RedirectUri,
			Handler: Login(a),
		},
	}
	authn.Repository.PluginApi.Router.AddAppRoutes(a.app.GetName(), routes)
}
