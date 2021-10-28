package pwbased

import (
	"aureole/internal/identity"
	ckeyTypes "aureole/internal/plugins/cryptokey/types"
	storageT "aureole/internal/plugins/storage/types"
	"context"
	"errors"
	"fmt"
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

func getCredField(i *identity.Identity, iData *storageT.IdentityData) (string, interface{}, error) {
	var credName string
	credVals := map[string]interface{}{}

	if iData.Username != nil && isCredential(i.Username) {
		credVals["username"] = iData.Username
		credName = "username"
	}

	if iData.Email != nil && isCredential(i.Email) {
		credVals["email"] = iData.Email
		credName = "email"
	}

	if iData.Phone != nil && isCredential(i.Phone) {
		credVals["phone"] = iData.Phone
		credName = "phone"
	}

	if l := len(credVals); l != 1 {
		return "", nil, fmt.Errorf("expects 1 credential, %d got", l)
	}

	return credName, credVals[credName], nil
}

func isCredential(trait identity.Trait) bool {
	return trait.IsCredential && trait.IsUnique
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
