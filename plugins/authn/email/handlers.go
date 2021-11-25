package email

import (
	"aureole/internal/identity"
	"aureole/internal/jwt"
	authzT "aureole/internal/plugins/authz/types"
	"github.com/gofiber/fiber/v2"
)

func SendMagicLink(e *email) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var i input
		if err := c.BodyParser(&i); err != nil {
			return sendError(c, fiber.StatusBadRequest, err.Error())
		}

		token, err := jwt.CreateJWT(map[string]interface{}{"email": i.Email}, e.conf.Exp)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}
		link := attachToken(e.magicLink, token)

		err = e.sender.Send(i.Email, "", e.conf.Template, map[string]interface{}{"link": link})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		return c.JSON(&fiber.Map{"status": "success"})
	}
}

func Login(e *email) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		rawToken := c.Query("token")
		if rawToken == "" {
			return sendError(c, fiber.StatusNotFound, "token not found")
		}

		token, err := jwt.ParseJWT(rawToken)
		if err != nil {
			return sendError(c, fiber.StatusBadRequest, err.Error())
		}
		email, ok := token.Get("email")
		if !ok {
			return sendError(c, fiber.StatusBadRequest, "cannot get email from token")
		}
		if err := jwt.InvalidateJWT(token); err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		var i = make(map[string]interface{})
		if e.manager != nil {
			i, err = e.manager.OnUserAuthenticated(
				&identity.Credential{
					Name:  identity.Email,
					Value: email.(string),
				},
				&identity.Identity{
					Email:         email.(string),
					EmailVerified: true,
				},
				AdapterName,
				nil)
			if err != nil {
				return sendError(c, fiber.StatusInternalServerError, err.Error())
			}
		} else {
			i["email"] = email.(string)
		}

		return e.authorizer.Authorize(c, authzT.NewPayload(e.authorizer, nil, i))
	}
}
