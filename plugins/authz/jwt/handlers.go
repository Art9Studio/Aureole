package jwt

import (
	"aureole/internal/plugins/authz/types"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/lestrrat-go/jwx/jwt"
)

func Refresh(j *jwtAuthz) func(*fiber.Ctx) error {
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
			return c.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		}

		userId, ok := refreshT.Get("user_id")
		if !ok {
			return c.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
				"success": false,
				"message": "can't access user_id from token",
			})
		}

		username, ok := refreshT.Get("username")
		if !ok {
			return c.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
				"success": false,
				"message": "can't access username from token",
			})
		}

		authzCtx := &types.Context{
			Username: username.(string),
			UserId:   int(userId.(float64)),
		}

		accessT, err := newToken(AccessToken, j.conf, authzCtx)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(&fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		}

		signedAccessT, err := signToken(j.signKey, accessT)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(&fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		}

		return attachTokens(c,
			map[string]bearerType{"access": j.conf.AccessBearer},
			keyMap,
			map[string][]byte{"access": signedAccessT})
	}
}

func getRawToken(c *fiber.Ctx, bearer bearerType, names map[string]string) (string, error) {
	switch bearer {
	case Header:
		/*authHeader := c.Get("Authorization")
		splitHeader := strings.Split(authHeader, "Bearer ")
		if len(splitHeader) != 2 {
			return "", c.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
				"success": false,
				"message": "bearer input not in proper format",
			})
		}
		return strings.TrimSpace(splitHeader[1]), nil*/
		var input map[string]string

		if err := c.BodyParser(&input); err != nil {
			return "", err
		}

		return input["refresh"], nil
	case Both, Cookie:
		rawToken := c.Cookies(names["cookie"])
		if rawToken == "" {
			return "", c.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
				"success": false,
				"message": fmt.Sprintf("cookie '%s' doesn't exist", names["cookie"]),
			})
		}
		return rawToken, nil
	default:
		return "", c.Status(fiber.StatusInternalServerError).JSON(&fiber.Map{
			"success": false,
			"message": fmt.Sprintf("unexpected bearer name: %s", bearer),
		})
	}
}
