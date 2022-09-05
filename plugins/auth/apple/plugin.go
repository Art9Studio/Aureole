package apple

import (
	"context"
	"errors"
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3gen"
	"net/http"
	"strconv"
	"time"

	"aureole/internal/configs"
	"aureole/internal/core"

	"github.com/gofiber/fiber/v2"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/mitchellh/mapstructure"

	_ "embed"
)

//go:embed meta.yaml
var rawMeta []byte

var meta core.Metadata

// init initializes package by register pluginCreator
func init() {
	meta = core.AuthenticatorRepo.Register(rawMeta, Create)
}

type GetAuthHandlerReqBody struct {
	State string `json:"state"`
	Code  string `json:"code"`
}

type apple struct {
	pluginAPI core.PluginAPI
	rawConf   configs.AuthPluginConfig
	conf      *config
	secretKey core.CryptoKey
	publicKey core.CryptoKey
	provider  *providerConfig
}

func (a *apple) GetAuthHTTPMethod() string {
	return http.MethodGet
}

func Create(conf configs.AuthPluginConfig) core.Authenticator {
	return &apple{rawConf: conf}
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

	return nil
}

func (apple) GetMetadata() core.Metadata {
	return meta
}

func (a *apple) GetAuthHandler() core.AuthHandlerFunc {
	return func(c fiber.Ctx) (*core.AuthResult, error) {
		input := &GetAuthHandlerReqBody{}
		if err := c.BodyParser(input); err != nil {
			return nil, err
		}
		if input.State != "state" {
			return nil, errors.New("invalid state")
		}
		if input.Code == "" {
			return nil, errors.New("code not found")
		}

		var (
			pluginId   = fmt.Sprintf("%d", meta.PluginID)
			email      string
			providerId string
		)

		jwtT, err := getJwt(a, input.Code)
		if err != nil {
			return nil, err
		}
		if err = a.pluginAPI.GetFromJWT(jwtT, "email", &email); err != nil {
			return nil, errors.New("cannot get email from token")
		}
		if err = a.pluginAPI.GetFromJWT(jwtT, "sub", &providerId); err != nil {
			return nil, errors.New("cannot get sub from token")
		}

		userData, err := jwtT.AsMap(context.Background())
		if err != nil {
			return nil, err
		}

		ok, err := a.pluginAPI.Filter(convertUserData(userData), a.conf.Filter)
		if err != nil {
			return nil, err
		} else if !ok {
			return nil, errors.New("input data doesn't pass filters")
		}

		return &core.AuthResult{
			User: &core.User{
				Email:         &email,
				EmailVerified: true,
			},
			ImportedUser: &core.ImportedUser{
				ProviderId:   providerId,
				PluginID:     pluginId,
				ProviderName: meta.ShortName,
				Additional:   userData,
			},
			Cred: &core.Credential{
				Name:  core.Email,
				Value: email,
			},
			Provider: meta.ShortName,
		}, nil
	}
}

func (a *apple) GetOAS3AuthRequestBody() *openapi3.RequestBody {
	schema, _ := openapi3gen.NewSchemaRefForValue(GetAuthHandlerReqBody{}, nil)
	return &openapi3.RequestBody{
		Description: "State & Code",
		Required:    true,
		Content: map[string]*openapi3.MediaType{
			fiber.MIMEApplicationJSON: {
				Schema: schema,
			},
		},
	}
}

func (a *apple) GetOAS3AuthParameters() openapi3.Parameters {
	return openapi3.Parameters{}
}

func initConfig(conf *configs.RawConfig) (*config, error) {
	PluginConf := &config{}
	if err := mapstructure.Decode(conf, PluginConf); err != nil {
		return nil, err
	}
	PluginConf.setDefaults()
	return PluginConf, nil
}

func initProvider(a *apple) error {
	url := a.pluginAPI.GetAppUrl()
	redirectUri := a.pluginAPI.GetAppUrl()
	redirectUri.Path = a.pluginAPI.GetAuthRoute(meta.ShortName)
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

func createSecret(p *providerConfig, key core.CryptoKey) error {
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

func (a *apple) GetCustomAppRoutes() []*core.Route {
	return []*core.Route{
		{
			Method:        http.MethodGet,
			Path:          core.GetOAuthPathPostfix(),
			Handler:       getAuthCode(a),
			OAS3Operation: assembleOAS3Operation(),
		},
	}
}

func assembleOAS3Operation() *openapi3.Operation {
	description := "Redirect"
	return &openapi3.Operation{
		OperationID: meta.ShortName,
		Description: meta.DisplayName,
		Responses: map[string]*openapi3.ResponseRef{
			strconv.Itoa(http.StatusFound): {
				Value: core.AssembleOASRedirectResponse(&description),
			},
		},
	}
}
