package jwt

import (
	"aureole/internal/plugins/authz/types"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/lestrrat-go/jwx/jwt"
	"strings"
)

func Refresh(j *jwtAuthz) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		rawRefreshT, err := getRawToken(c, j.conf.RefreshBearer, names["refresh"])
		if err != nil {
			return err
		}

		keySet := j.signKey.GetPublicSet()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(&fiber.Map{
				"success": false,
				"message": err,
			})
		}

		refreshT, err := jwt.ParseString(
			rawRefreshT,
			jwt.WithIssuer(j.conf.Iss),
			jwt.WithValidate(true),
			jwt.WithKeySet(keySet),
		)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
				"success": false,
				"message": err,
			})
		}

		if userId, ok := refreshT.Get("user_id"); ok {
			accessT, err := newToken(AccessToken, j.conf, &types.Context{UserId: int(userId.(float64))})
			if err != nil {
				return err
			}

			signedAccessT, err := signToken(j.signKey, accessT)
			if err != nil {
				return err
			}

			return sendToken(c, j.conf.AccessBearer, names["access"], signedAccessT)
		} else {
			return c.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
				"success": false,
				"message": "can't access user_id from token",
			})
		}
	}
}

func getRawToken(c *fiber.Ctx, bearer bearerType, names map[string]string) (string, error) {
	switch bearer {
	case Header:
		authHeader := c.Get("Authorization")
		splitHeader := strings.Split(authHeader, "Bearer ")
		if len(splitHeader) != 2 {
			return "", c.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
				"success": false,
				"message": "bearer token not in proper format",
			})
		}
		return strings.TrimSpace(splitHeader[1]), nil
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
