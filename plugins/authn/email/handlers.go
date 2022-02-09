package email

import (
	"aureole/internal/core"
	"fmt"
	"net/url"

	"github.com/gofiber/fiber/v2"
)

func sendMagicLink(e *authn) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var i input
		if err := c.BodyParser(&i); err != nil {
			return core.SendError(c, fiber.StatusBadRequest, err.Error())
		}

		fmt.Print(c.Request().URI().String())
		token, err := e.pluginAPI.CreateJWT(map[string]interface{}{"email": i.Email, "url": c.Request().URI().String()},
			e.conf.Exp)
		if err != nil {
			return core.SendError(c, fiber.StatusInternalServerError, err.Error())
		}
		link := attachToken(e.magicLink, token)

		err = e.sender.Send(i.Email, "", e.tmpl, e.tmplExt, map[string]interface{}{"link": link})
		if err != nil {
			return core.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		return c.SendStatus(fiber.StatusOK)
	}
}

func attachToken(u *url.URL, token string) string {
	q := u.Query()
	q.Set("token", token)
	u.RawQuery = q.Encode()

	return u.String()
}
