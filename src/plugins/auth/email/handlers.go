package email

import (
	"aureole/internal/core"
	"net/http"
	"net/url"

	"github.com/gofiber/fiber/v2"
)

func sendMagicLink(e *email) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var i SendMagicLinkReqBody
		if err := c.BodyParser(&i); err != nil {
			return core.SendError(c, http.StatusBadRequest, err.Error())
		}

		tokenRaw, err := e.pluginAPI.CreateJWT(map[string]interface{}{"email": i.Email}, e.conf.Exp)
		if err != nil {
			return core.SendError(c, http.StatusInternalServerError, err.Error())
		}
		link := attachToken(e.magicLink, tokenRaw)

		err = e.sender.Send(i.Email, "", e.tmpl, e.tmplExt, map[string]interface{}{"link": link})
		if err != nil {
			return core.SendError(c, http.StatusInternalServerError, err.Error())
		}

		return c.SendStatus(http.StatusOK)
	}
}

func attachToken(u *url.URL, token string) string {
	q := u.Query()
	q.Set("token", token)
	u.RawQuery = q.Encode()

	return u.String()
}
