package email

import (
	"aureole/internal/core"
	"github.com/gofiber/fiber/v2"
	"net/url"
)

func sendMagicLink(e *email) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var i input
		if err := c.BodyParser(&i); err != nil {
			return core.SendError(c, fiber.StatusBadRequest, err.Error())
		}

		token, err := core.CreateJWT(map[string]interface{}{"email": i.Email}, e.conf.Exp)
		if err != nil {
			return core.SendError(c, fiber.StatusInternalServerError, err.Error())
		}
		link := attachToken(e.magicLink, token)

		err = e.sender.Send(i.Email, "", e.conf.Template, map[string]interface{}{"link": link})
		if err != nil {
			return core.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		return c.JSON(&fiber.Map{"success": true})
	}
}

func attachToken(u *url.URL, token string) string {
	q := u.Query()
	q.Set("token", token)
	u.RawQuery = q.Encode()

	return u.String()
}
