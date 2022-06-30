package pwbased

import (
	"aureole/internal/core"
	"errors"
	"github.com/gofiber/fiber/v2"
	"net/url"
)

func register(p *pwBased) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var rawCred *credential
		if err := c.BodyParser(rawCred); err != nil {
			return core.SendError(c, fiber.StatusBadRequest, err.Error())
		}
		if rawCred.Password == "" {
			return core.SendError(c, fiber.StatusBadRequest, "password required")
		}

		pwHash, err := p.pwHasher.HashPw(rawCred.Password)
		if err != nil {
			return core.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		i := &core.Identity{
			ID:         rawCred.Id,
			Username:   &rawCred.Username,
			Phone:      &rawCred.Phone,
			Email:      &rawCred.Email,
			Additional: map[string]interface{}{core.Password: pwHash},
		}
		cred, err := getCredential(i)
		if err != nil {
			return core.SendError(c, fiber.StatusBadRequest, err.Error())
		}
		manager, ok := p.pluginAPI.GetIDManager()
		if !ok {
			return core.SendError(c, fiber.StatusInternalServerError, "could not get ID manager")
		}
		user, err := manager.Register(cred, i, meta.Name)
		if err != nil {
			return err
		}
		_ = user

		if p.conf.Register.IsVerifyAfter {
			token, err := p.pluginAPI.CreateJWT(map[string]interface{}{"email": rawCred.Email}, p.conf.Verify.Exp)
			if err != nil {
				return core.SendError(c, fiber.StatusInternalServerError, err.Error())
			}
			link := attachToken(p.verify.confirmLink, token)

			err = p.verify.sender.Send(rawCred.Email, "", p.verify.tmpl, p.verify.tmplExt, map[string]interface{}{"link": link})
			if err != nil {
				return core.SendError(c, fiber.StatusInternalServerError, err.Error())
			}
		}

		/*if p.conf.register.IsLoginAfter {
			payload, err := authzT.NewIssuerPayload(user)
			if err != nil {
				return router.SendError(c, fiber.StatusInternalServerError, err.Error())
			}
			return p.authorizer.Authorize(c, payload)
		} else {*/
		return c.SendStatus(fiber.StatusOK)
		//}
	}
}

func Reset(p *pwBased) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var e *email
		if err := c.BodyParser(e); err != nil {
			return core.SendError(c, fiber.StatusBadRequest, err.Error())
		}
		if e.Email == "" {
			return core.SendError(c, fiber.StatusBadRequest, "email required")
		}
		i := &core.Identity{Email: &e.Email}

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

func ResetConfirm(p *pwBased) func(*fiber.Ctx) error {
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

		var input *credential
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

		manager, ok := p.pluginAPI.GetIDManager()
		if !ok {
			return core.SendError(c, fiber.StatusInternalServerError, "could not get ID manager")
		}

		_, err = manager.Update(
			&core.Credential{
				Name:  core.Email,
				Value: email.(string),
			},
			&core.Identity{
				Additional: map[string]interface{}{core.Password: pwHash},
			},
			meta.Name)
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

func Verify(p *pwBased) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var e *email
		if err := c.BodyParser(e); err != nil {
			return core.SendError(c, fiber.StatusBadRequest, err.Error())
		}
		if e.Email == "" {
			return core.SendError(c, fiber.StatusBadRequest, "email required")
		}
		i := &core.Identity{Email: &e.Email}

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

func VerifyConfirm(p *pwBased) func(*fiber.Ctx) error {
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

		manager, ok := p.pluginAPI.GetIDManager()
		if !ok {
			return core.SendError(c, fiber.StatusInternalServerError, "could not get ID manager")
		}

		_, err = manager.Update(
			&core.Credential{
				Name:  core.Email,
				Value: email.(string),
			},
			&core.Identity{EmailVerified: true},
			meta.Name)
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

func getCredential(i *core.Identity) (*core.Credential, error) {
	if *i.Username != "nil" {
		return &core.Credential{
			Name:  "username",
			Value: *i.Username,
		}, nil
	}

	if *i.Email != "nil" {
		return &core.Credential{
			Name:  "email",
			Value: *i.Email,
		}, nil
	}

	if *i.Phone != "nil" {
		return &core.Credential{
			Name:  "phone",
			Value: *i.Phone,
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