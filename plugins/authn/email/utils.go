package email

import (
	"aureole/internal/identity"
	ckeyTypes "aureole/internal/plugins/cryptokey/types"
	"context"
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jwt"
	"net/url"
	"time"
)

func sendError(c *fiber.Ctx, statusCode int, message string) error {
	return c.Status(statusCode).JSON(&fiber.Map{
		"success": false,
		"message": message,
	})
}

func isCredential(trait *identity.Trait) bool {
	return trait.IsCredential && trait.IsUnique
}

func createToken(e *email, claims map[string]interface{}) (string, error) {
	token := jwt.New()

	if err := token.Set(jwt.IssuerKey, "Aureole Internal"); err != nil {
		return "", err
	}
	if err := token.Set(jwt.AudienceKey, "Aureole Internal"); err != nil {
		return "", err
	}
	if err := token.Set(jwt.ExpirationKey, time.Now().Add(time.Duration(e.conf.Exp)*time.Second).Unix()); err != nil {
		return "", err
	}

	for claimName, claim := range claims {
		if err := token.Set(claimName, claim); err != nil {
			return "", err
		}
	}

	signedToken, err := signToken(e.serviceKey, token)
	if err != nil {
		return "", err
	}

	return string(signedToken), err
}

func signToken(key ckeyTypes.CryptoKey, token jwt.Token) ([]byte, error) {
	keySet := key.GetPrivateSet()

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

func attachToken(u *url.URL, token string) string {
	q := u.Query()
	q.Set("token", token)
	u.RawQuery = q.Encode()

	return u.String()
}
