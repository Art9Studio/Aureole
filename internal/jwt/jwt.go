package jwt

import (
	cryptoKeyT "aureole/internal/plugins/cryptokey/types"
	state "aureole/internal/state/interface"
	"context"
	"errors"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jwt"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"time"
)

var project state.ProjectState

func Init(p state.ProjectState) {
	project = p
}

func CreateJWT(payload map[string]interface{}, exp int) (string, error) {
	token, err := newToken(payload, exp)
	if err != nil {
		return "", err
	}

	keySet, err := project.GetServiceSignKey()
	if err != nil {
		return "", err
	}
	signedToken, err := signToken(keySet, token)
	if err != nil {
		return "", err
	}
	return string(signedToken), nil
}

func ParseJWT(rawToken string) (jwt.Token, error) {
	keySet, err := project.GetServiceSignKey()
	if err != nil {
		return nil, err
	}
	token, err := jwt.ParseString(
		rawToken,
		jwt.WithIssuer("Aureole"),
		jwt.WithAudience("Aureole"),
		jwt.WithClaimValue("type", "service"),
		jwt.WithValidate(true),
		jwt.WithKeySet(keySet.GetPublicSet()),
	)
	if err != nil {
		return nil, err
	}

	storage, err := project.GetServiceStorage()
	if err != nil {
		return nil, err
	}
	ok, err := storage.Exists(token.JwtID())
	if err != nil {
		return nil, err
	} else if ok {
		return nil, errors.New("jwt: invalid token")
	} else {
		return token, nil
	}
}

func InvalidateJWT(token jwt.Token) error {
	if time.Now().Before(token.Expiration()) {
		storage, err := project.GetServiceStorage()
		if err != nil {
			return err
		}

		err = storage.Set(token.JwtID(), token, int(time.Until(token.Expiration()).Seconds()))
		if err != nil {
			return err
		}
	}
	return nil
}

func newToken(payload map[string]interface{}, exp int) (jwt.Token, error) {
	predefinedKeys := []string{
		jwt.JwtIDKey,
		jwt.IssuerKey,
		jwt.AudienceKey,
		jwt.ExpirationKey,
		jwt.NotBeforeKey,
		jwt.IssuedAtKey,
		"type",
	}
	for _, k := range predefinedKeys {
		delete(payload, k)
	}

	token := jwt.New()

	jti, err := gonanoid.New(16)
	if err != nil {
		return nil, err
	}
	err = token.Set(jwt.JwtIDKey, jti)
	if err != nil {
		return nil, err
	}
	err = token.Set(jwt.IssuerKey, "Aureole")
	if err != nil {
		return nil, err
	}
	err = token.Set(jwt.AudienceKey, "Aureole")
	if err != nil {
		return nil, err
	}
	err = token.Set(jwt.ExpirationKey, time.Now().Add(time.Duration(exp)*time.Second).Unix())
	if err != nil {
		return nil, err
	}
	err = token.Set(jwt.NotBeforeKey, time.Now().Unix())
	if err != nil {
		return nil, err
	}
	err = token.Set(jwt.IssuedAtKey, time.Now().Unix())
	if err != nil {
		return nil, err
	}
	err = token.Set("type", "service")
	if err != nil {
		return nil, err
	}

	for k, v := range payload {
		err := token.Set(k, v)
		if err != nil {
			return nil, err
		}
	}
	return token, err
}

func signToken(signKey cryptoKeyT.CryptoKey, token jwt.Token) ([]byte, error) {
	keySet := signKey.GetPrivateSet()

	for it := keySet.Iterate(context.Background()); it.Next(context.Background()); {
		pair := it.Pair()
		key := pair.Value.(jwk.Key)

		if key.KeyUsage() == "sig" {
			var signAlg jwa.SignatureAlgorithm
			if err := signAlg.Accept(key.Algorithm()); err != nil {
				return nil, err
			}
			return jwt.Sign(token, signAlg, key)
		}
	}

	return nil, errors.New("key set don't contain sig key")
}
