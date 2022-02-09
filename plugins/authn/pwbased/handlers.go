package pwbased

import (
	"aureole/internal/core"
	"aureole/internal/plugins"
	"errors"
	"net/url"

	"github.com/gofiber/fiber/v2"
)

func register(p *authn) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var input credentialInput
		if err := c.BodyParser(&input); err != nil {
			return core.SendError(c, fiber.StatusBadRequest, err.Error())
		}
		if input.Password == "" {
			return core.SendError(c, fiber.StatusBadRequest, "password required")
		}

		pwHash, err := p.pwHasher.HashPw(input.Password)
		if err != nil {
			return core.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		i, err := plugins.NewIdentity(input.AsMap())
		if err != nil {
			return core.SendError(c, fiber.StatusInternalServerError, err.Error())
		}
		i.Additional = map[string]interface{}{plugins.Password: pwHash}

		cred, err := getCredential(&input)
		if err != nil {
			return core.SendError(c, fiber.StatusBadRequest, err.Error())
		}
		user, err := p.manager.Register(cred, i, adapterName)
		if err != nil {
			return err
		}
		_ = user

		if p.conf.Register.IsVerifyAfter {
			token, err := p.pluginAPI.CreateJWT(map[string]interface{}{"email": input.Email}, p.conf.Verify.Exp)
			if err != nil {
				return core.SendError(c, fiber.StatusInternalServerError, err.Error())
			}
			link := attachToken(p.verify.confirmLink, token)

			err = p.verify.sender.Send(input.Email, "", p.verify.tmpl, p.verify.tmplExt, map[string]interface{}{"link": link})
			if err != nil {
				return core.SendError(c, fiber.StatusInternalServerError, err.Error())
			}
		}

		/*if p.conf.register.IsLoginAfter {
			payload, err := authzT.NewPayload(user)
			if err != nil {
				return router.SendError(c, fiber.StatusInternalServerError, err.Error())
			}
			return p.authorizer.Authorize(c, payload)
		} else {*/
		return c.SendStatus(fiber.StatusOK)
		//}
	}
}

func Reset(p *authn) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var e *email
		if err := c.BodyParser(e); err != nil {
			return core.SendError(c, fiber.StatusBadRequest, err.Error())
		}
		if e.Email == "" {
			return core.SendError(c, fiber.StatusBadRequest, "email required")
		}
		i := &plugins.Identity{Email: &e.Email}

		token, err := p.pluginAPI.CreateJWT(map[string]interface{}{"email": i.Email}, p.conf.Reset.Exp)
		if err != nil {
			return core.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		link := attachToken(p.reset.confirmLink, token)
		err = p.reset.sender.Send(*i.Email, "", p.reset.tmpl, p.reset.tmplExt, map[string]interface{}{"link": link})
		if err != nil {
			return core.SendError(c, fiber.StatusInternalServerError, err.Error())
		}
		return c.SendStatus(fiber.StatusOK)
	}
}

func ResetConfirm(p *authn) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		rawToken := c.Query("token")
		if rawToken == "" {
			return core.SendError(c, fiber.StatusBadRequest, "token not found")
		}

		token, err := p.pluginAPI.ParseJWT(rawToken)
		if err != nil {
			return core.SendError(c, fiber.StatusBadRequest, err.Error())
		}
		email, ok := token.Get("email")
		if !ok {
			return core.SendError(c, fiber.StatusBadRequest, "cannot get email from token")
		}
		if err := p.pluginAPI.InvalidateJWT(token); err != nil {
			return core.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		var input *credentialInput
		if err := c.BodyParser(input); err != nil {
			return core.SendError(c, fiber.StatusBadRequest, err.Error())
		}
		if input.Password == "" {
			return core.SendError(c, fiber.StatusBadRequest, "password required")
		}

		pwHash, err := p.pwHasher.HashPw(input.Password)
		if err != nil {
			return core.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		_, err = p.manager.Update(
			&plugins.Credential{
				Name:  plugins.Email,
				Value: email.(string),
			},
			&plugins.Identity{
				Additional: map[string]interface{}{plugins.Password: pwHash},
			},
			adapterName)
		if err != nil {
			return core.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		err = p.reset.sender.SendRaw(email.(string), "Reset your password",
			"Your password has been successfully changed")
		if err != nil {
			return core.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		// todo: add expiring any current user session
		redirectUrl := c.Query("redirect_url")
		if redirectUrl != "" {
			return c.Redirect(redirectUrl)
		}
		return c.SendStatus(fiber.StatusOK)
	}
}

func Verify(p *authn) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var e *email
		if err := c.BodyParser(e); err != nil {
			return core.SendError(c, fiber.StatusBadRequest, err.Error())
		}
		if e.Email == "" {
			return core.SendError(c, fiber.StatusBadRequest, "email required")
		}
		i := &plugins.Identity{Email: &e.Email}

		token, err := p.pluginAPI.CreateJWT(map[string]interface{}{"email": i.Email}, p.conf.Verify.Exp)
		if err != nil {
			return core.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		link := attachToken(p.verify.confirmLink, token)
		err = p.verify.sender.Send(*i.Email, "", p.verify.tmpl, p.verify.tmplExt, map[string]interface{}{"link": link})
		if err != nil {
			return core.SendError(c, fiber.StatusInternalServerError, err.Error())
		}
		return c.SendStatus(fiber.StatusOK)
	}
}

func VerifyConfirm(p *authn) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		rawToken := c.Query("token")
		if rawToken == "" {
			return core.SendError(c, fiber.StatusBadRequest, "token not found")
		}

		token, err := p.pluginAPI.ParseJWT(rawToken)
		if err != nil {
			return core.SendError(c, fiber.StatusBadRequest, err.Error())
		}
		email, ok := token.Get("email")
		if !ok {
			return core.SendError(c, fiber.StatusBadRequest, "cannot get email from token")
		}
		if err := p.pluginAPI.InvalidateJWT(token); err != nil {
			return core.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		_, err = p.manager.Update(
			&plugins.Credential{
				Name:  plugins.Email,
				Value: email.(string),
			},
			&plugins.Identity{EmailVerified: true},
			adapterName)
		if err != nil {
			return core.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		redirectUrl := c.Query("redirect_url")
		if redirectUrl != "" {
			return c.Redirect(redirectUrl)
		}
		return c.JSON(fiber.Map{"success": true})
	}
}

func getCredential(c *credentialInput) (*plugins.Credential, error) {
	if c.Username != "" {
		return &plugins.Credential{
			Name:  "username",
			Value: c.Username,
		}, nil
	}

	if c.Email != "" {
		return &plugins.Credential{
			Name:  "email",
			Value: c.Email,
		}, nil
	}

	if c.Phone != "" {
		return &plugins.Credential{
			Name:  "phone",
			Value: c.Phone,
		}, nil
	}

	return nil, errors.New("credential not found")
}

func attachToken(u *url.URL, token string) string {
	q := u.Query()
	q.Set("token", token)
	u.RawQuery = q.Encode()

	return u.String()
}
