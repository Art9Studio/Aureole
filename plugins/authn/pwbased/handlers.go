package pwbased

import (
	"aureole/internal/core"
	"aureole/internal/identity"
	"errors"
	"github.com/gofiber/fiber/v2"
	"net/url"
)

func register(p *pwBased) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var input *input
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

		i := &identity.Identity{
			ID:         input.Id,
			Username:   input.Username,
			Phone:      input.Phone,
			Email:      input.Email,
			Additional: map[string]interface{}{identity.Password: pwHash},
		}
		cred, err := getCredential(i)
		if err != nil {
			return core.SendError(c, fiber.StatusBadRequest, err.Error())
		}
		user, err := p.manager.Register(cred, i, adapterName)
		if err != nil {
			return err
		}
		_ = user

		if p.conf.Register.IsVerifyAfter {
			token, err := core.CreateJWT(map[string]interface{}{"email": input.Email}, p.conf.Verif.Exp)
			if err != nil {
				return core.SendError(c, fiber.StatusInternalServerError, err.Error())
			}
			link := attachToken(p.verifyConfirmLink, token)

			err = p.verifySender.Send(input.Email, "", p.conf.Verif.Template, map[string]interface{}{"link": link})
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
		return c.JSON(fiber.Map{"success": true})
		//}
	}
}

func Reset(p *pwBased) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var input *input
		if err := c.BodyParser(input); err != nil {
			return core.SendError(c, fiber.StatusBadRequest, err.Error())
		}
		if input.Email == "" {
			return core.SendError(c, fiber.StatusBadRequest, "email required")
		}
		i := &identity.Identity{Email: input.Email}

		token, err := core.CreateJWT(map[string]interface{}{"email": i.Email}, p.conf.Reset.Exp)
		if err != nil {
			return core.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		link := attachToken(p.resetConfirmLink, token)
		err = p.resetSender.Send(i.Email, "", p.conf.Reset.Template, map[string]interface{}{"link": link})
		if err != nil {
			return core.SendError(c, fiber.StatusInternalServerError, err.Error())
		}
		return c.JSON(fiber.Map{"success": true})
	}
}

func ResetConfirm(p *pwBased) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		rawToken := c.Query("token")
		if rawToken == "" {
			return core.SendError(c, fiber.StatusNotFound, "token not found")
		}

		token, err := core.ParseJWT(rawToken)
		if err != nil {
			return core.SendError(c, fiber.StatusBadRequest, err.Error())
		}
		email, ok := token.Get("email")
		if !ok {
			return core.SendError(c, fiber.StatusBadRequest, "cannot get email from token")
		}
		if err := core.InvalidateJWT(token); err != nil {
			return core.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		var input *input
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
			&identity.Credential{
				Name:  identity.Email,
				Value: email.(string),
			},
			&identity.Identity{
				Additional: map[string]interface{}{identity.Password: pwHash},
			},
			adapterName)
		if err != nil {
			return core.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		err = p.resetSender.SendRaw(email.(string), "Reset your password",
			"Your password has been successfully changed")
		if err != nil {
			return core.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		// todo: add expiring any current user session
		redirectUrl := c.Query("redirect_url")
		if redirectUrl != "" {
			return c.Redirect(redirectUrl)
		}
		return c.JSON(fiber.Map{"success": true})
	}
}

func Verify(p *pwBased) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var input *input
		if err := c.BodyParser(input); err != nil {
			return core.SendError(c, fiber.StatusBadRequest, err.Error())
		}
		if input.Email == "" {
			return core.SendError(c, fiber.StatusBadRequest, "email required")
		}
		i := &identity.Identity{Email: input.Email}

		token, err := core.CreateJWT(map[string]interface{}{"email": i.Email}, p.conf.Verif.Exp)
		if err != nil {
			return core.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		link := attachToken(p.verifyConfirmLink, token)
		err = p.verifySender.Send(i.Email, "", p.conf.Verif.Template, map[string]interface{}{"link": link})
		if err != nil {
			return core.SendError(c, fiber.StatusInternalServerError, err.Error())
		}
		return c.JSON(fiber.Map{"success": true})
	}
}

func VerifyConfirm(p *pwBased) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		rawToken := c.Query("token")
		if rawToken == "" {
			return core.SendError(c, fiber.StatusNotFound, "token not found")
		}

		token, err := core.ParseJWT(rawToken)
		if err != nil {
			return core.SendError(c, fiber.StatusBadRequest, err.Error())
		}
		email, ok := token.Get("email")
		if !ok {
			return core.SendError(c, fiber.StatusBadRequest, "cannot get email from token")
		}
		if err := core.InvalidateJWT(token); err != nil {
			return core.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		_, err = p.manager.Update(
			&identity.Credential{
				Name:  identity.Email,
				Value: email.(string),
			},
			&identity.Identity{EmailVerified: true},
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

func getCredential(i *identity.Identity) (*identity.Credential, error) {
	if i.Username != "nil" {
		return &identity.Credential{
			Name:  "username",
			Value: i.Username,
		}, nil
	}

	if i.Email != "nil" {
		return &identity.Credential{
			Name:  "email",
			Value: i.Email,
		}, nil
	}

	if i.Phone != "nil" {
		return &identity.Credential{
			Name:  "phone",
			Value: i.Phone,
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
