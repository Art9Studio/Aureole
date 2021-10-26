package apple

import (
	"aureole/internal/identity"
	authzT "aureole/internal/plugins/authz/types"
	"context"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/lestrrat-go/jwx/jwt"
)

func GetAuthCode(a *apple) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		u := a.provider.AuthCodeURL("state")
		return c.Redirect(u)
	}
}

func Login(a *apple) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		input := struct {
			State string
			Code  string
		}{}
		if err := c.BodyParser(&input); err != nil {
			return err
		}
		if input.State != "state" {
			return sendError(c, fiber.StatusBadRequest, "invalid state")
		}
		if input.Code == "" {
			return sendError(c, fiber.StatusBadRequest, "code not found")
		}

		jwtT, err := getJwt(a, input.Code)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		email, ok := jwtT.Get("email")
		if !ok {
			return sendError(c, fiber.StatusInternalServerError, "can't get 'email' from token")
		}
		socialId, ok := jwtT.Get("sub")
		if !ok {
			return sendError(c, fiber.StatusInternalServerError, "can't get 'social_id' from token")
		}
		userData, err := jwtT.AsMap(context.Background())
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		if ok, err := a.app.Filter(convertUserData(userData), a.rawConf.Filter); err != nil {
			return sendError(c, fiber.StatusBadRequest, err.Error())
		} else if !ok {
			return sendError(c, fiber.StatusBadRequest, "apple: input data doesn't pass filters")
		}

		var i map[string]interface{}
		if a.manager != nil {
			i, err = a.manager.OnUserAuthenticated(
				&identity.Credential{
					Name:  "email",
					Value: email.(string),
				},
				&identity.Identity{
					Email: email.(string),
				},
				AdapterName,
				map[string]interface{}{
					"social_id": socialId,
					"user_data": userData,
				})
			if err != nil {
				return sendError(c, fiber.StatusInternalServerError, err.Error())
			}
		} else {
			i = map[string]interface{}{
				"email":     email,
				"provider":  AdapterName,
				"social_id": socialId,
				"user_data": userData,
			}
		}

		return a.authorizer.Authorize(c, authzT.NewPayload(a.authorizer, nil, i))
	}
}

func getJwt(a *apple, code string) (jwt.Token, error) {
	t, err := a.provider.Exchange(code)
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
		jwt.WithAudience(a.provider.ClientId),
		jwt.WithKeySet(keySet))
}
