package apple

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/lestrrat-go/jwx/jwt"
)

func getAuthCode(a *apple) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		u := a.provider.authCodeURL("state")
		return c.Redirect(u)
	}
}

func getJwt(a *apple, code string) (jwt.Token, error) {
	t, err := a.provider.exchange(code)
	if err != nil {
		return nil, err
	}
	idToken := t["id_token"]

	keySet := a.publicKey.GetPublicSet()
	if keySet == nil {
		return nil, fmt.Errorf("apple: cannot get public set from %s", a.conf.PublicKey)
	}

	return jwt.ParseString(
		idToken.(string),
		jwt.WithAudience(a.provider.clientId),
		jwt.WithKeySet(keySet))
}

func convertUserData(mapIntr map[string]interface{}) map[string]string {
	mapStr := make(map[string]string, len(mapIntr))
	for key, value := range mapIntr {
		mapStr[key] = fmt.Sprintf("%v", value)
	}
	return mapStr
}
