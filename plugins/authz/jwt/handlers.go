package jwt

import (
	"aureole/internal/core"
	"aureole/internal/plugins"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/lestrrat-go/jwx/jwt"
)

func refresh(j *jwtAuthz) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		rawRefreshT, err := getRawToken(c, j.conf.RefreshBearer, keyMap["refresh"])
		if err != nil {
			return err
		}

		keySet := j.signKey.GetPublicSet()
		refreshT, err := jwt.ParseString(
			rawRefreshT,
			jwt.WithIssuer(j.conf.Iss),
			jwt.WithValidate(true),
			jwt.WithKeySet(keySet),
		)

		if err != nil {
			return core.SendError(c, fiber.StatusBadRequest, err.Error())
		}

		id, ok := refreshT.Get("id")
		if !ok {
			return core.SendError(c, fiber.StatusBadRequest, "can't access user_id from token")
		}

		// todo: add identity support
		// username, ok := refreshT.GetData("username")
		// if !ok {
		// 	return router.SendError(c, fiber.StatusBadRequest, "can't access username from token")
		// }

		payload := &plugins.Payload{
			// Username: username.(string),
			ID: int(id.(float64)),
		}

		accessT, err := newToken(accessToken, j.conf, payload)
		if err != nil {
			return core.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		signedAccessT, err := signToken(j.signKey, accessT)
		if err != nil {
			return core.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		return attachTokens(c,
			map[string]bearerType{"access": j.conf.AccessBearer},
			keyMap,
			map[string][]byte{"access": signedAccessT})
	}
}

func getRawToken(c *fiber.Ctx, bearer bearerType, names map[string]string) (token string, err error) {
	switch bearer {
	case cookie:
		rawToken := c.Cookies(names["cookie"])
		if rawToken == "" {
			return "", core.SendError(c, fiber.StatusBadRequest, fmt.Sprintf("cookie '%s' doesn't exist", names["cookie"]))
		}
		token = rawToken
	case both, body:
		var input map[string]string
		if err := c.BodyParser(&input); err != nil {
			return "", err
		}
		token = input["refresh"]
	}
	return token, err
}
