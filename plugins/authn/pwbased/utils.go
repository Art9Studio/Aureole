package pwbased

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
)

func sendError(c *fiber.Ctx, statusCode int, message string) error {
	return c.Status(statusCode).JSON(&fiber.Map{
		"success": false,
		"message": message,
	})
}

func getCredential(i *identity.Identity) (*identity.Credential, error) {
	if i.Username != "nil" {
		return &identity.Credential{
			Name:  "username",
			Value: i.Username,
		}, nil
	}

	if i.Email != "nil" {
		return &identity.Credential{
			Name:  "email",
			Value: i.Email,
		}, nil
	}

	if i.Phone != "nil" {
		return &identity.Credential{
			Name:  "phone",
			Value: i.Phone,
		}, nil
	}

	return nil, errors.New("credential not found")
}

func createToken(p *pwBased, claims map[string]interface{}) (string, error) {
	token := jwt.New()

	if err := token.Set(jwt.IssuerKey, "Aureole Internal"); err != nil {
		return "", err
	}
	if err := token.Set(jwt.AudienceKey, "Aureole Internal"); err != nil {
		return "", err
	}

	for claimName, claim := range claims {
		if err := token.Set(claimName, claim); err != nil {
			return "", err
		}
	}

	signedToken, err := signToken(p.serviceKey, token)
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
