package pwbased

import (
	"aureole/internal/identity"
	"aureole/internal/jwt"
	authzT "aureole/internal/plugins/authz/types"
	"aureole/internal/router"
	"errors"
	"github.com/gofiber/fiber/v2"
	"net/url"
)

func Register(p *pwBased) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var input *input
		if err := c.BodyParser(input); err != nil {
			return router.SendError(c, fiber.StatusBadRequest, err.Error())
		}
		if input.Password == "" {
			return router.SendError(c, fiber.StatusBadRequest, "password required")
		}

		i := &identity.Identity{
			Id:       input.Id,
			Username: input.Username,
			Phone:    input.Phone,
			Email:    input.Email,
		}
		pwHash, err := p.pwHasher.HashPw(input.Password)
		if err != nil {
			return router.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		cred, err := getCredential(i)
		if err != nil {
			return router.SendError(c, fiber.StatusBadRequest, err.Error())
		}
		user, err := p.manager.OnRegister(cred, i, AdapterName, map[string]interface{}{identity.Password: pwHash})
		if err != nil {
			return err
		}

		if p.conf.Register.IsVerifyAfter {
			token, err := jwt.CreateJWT(map[string]interface{}{"email": input.Email}, p.conf.Verif.Exp)
			if err != nil {
				return router.SendError(c, fiber.StatusInternalServerError, err.Error())
			}
			link := attachToken(p.verif.confirmLink, token)

			err = p.verif.sender.Send(input.Email, "", p.conf.Verif.Template, map[string]interface{}{"link": link})
			if err != nil {
				return router.SendError(c, fiber.StatusInternalServerError, err.Error())
			}
		}

		if p.conf.Register.IsLoginAfter {
			payload, err := authzT.NewPayload(user)
			if err != nil {
				return router.SendError(c, fiber.StatusInternalServerError, err.Error())
			}
			return p.authorizer.Authorize(c, payload)
		} else {
			return c.JSON(&fiber.Map{"status": "success"})
		}
	}
}

func Reset(p *pwBased) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var input *input
		if err := c.BodyParser(input); err != nil {
			return router.SendError(c, fiber.StatusBadRequest, err.Error())
		}
		if input.Email == "" {
			return router.SendError(c, fiber.StatusBadRequest, "email required")
		}
		i := &identity.Identity{Email: input.Email}

		token, err := jwt.CreateJWT(map[string]interface{}{"email": i.Email}, p.conf.Reset.Exp)
		if err != nil {
			return router.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		link := attachToken(p.reset.confirmLink, token)
		err = p.verif.sender.Send(i.Email, "", p.conf.Verif.Template, map[string]interface{}{"link": link})
		if err != nil {
			return router.SendError(c, fiber.StatusInternalServerError, err.Error())
		}
		return c.JSON(&fiber.Map{"status": "success"})
	}
}

func ResetConfirm(p *pwBased) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		rawToken := c.Query("token")
		if rawToken == "" {
			return router.SendError(c, fiber.StatusNotFound, "token not found")
		}

		token, err := jwt.ParseJWT(rawToken)
		if err != nil {
			return router.SendError(c, fiber.StatusBadRequest, err.Error())
		}
		email, ok := token.Get("email")
		if !ok {
			return router.SendError(c, fiber.StatusBadRequest, "cannot get email from token")
		}
		if err := jwt.InvalidateJWT(token); err != nil {
			return router.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		var input *input
		if err := c.BodyParser(input); err != nil {
			return router.SendError(c, fiber.StatusBadRequest, err.Error())
		}
		if input.Password == "" {
			return router.SendError(c, fiber.StatusBadRequest, "password required")
		}

		pwHash, err := p.pwHasher.HashPw(input.Password)
		if err != nil {
			return router.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		_, err = p.manager.Update(
			&identity.Credential{
				Name:  identity.Email,
				Value: email.(string),
			},
			AdapterName,
			map[string]interface{}{identity.Password: pwHash})
		if err != nil {
			return router.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		err = p.reset.sender.SendRaw(email.(string), "Reset your password",
			"Your password has been successfully changed")
		if err != nil {
			return router.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		// todo: add expiring any current user session
		redirectUrl := c.Query("redirect_url")
		if redirectUrl != "" {
			return c.Redirect(redirectUrl)
		}
		return c.JSON(&fiber.Map{"status": "success"})
	}
}

func Verify(p *pwBased) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var input *input
		if err := c.BodyParser(input); err != nil {
			return router.SendError(c, fiber.StatusBadRequest, err.Error())
		}
		if input.Email == "" {
			return router.SendError(c, fiber.StatusBadRequest, "email required")
		}
		i := &identity.Identity{Email: input.Email}

		token, err := jwt.CreateJWT(map[string]interface{}{"email": i.Email}, p.conf.Verif.Exp)
		if err != nil {
			return router.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		link := attachToken(p.verif.confirmLink, token)
		err = p.verif.sender.Send(i.Email, "", p.conf.Verif.Template, map[string]interface{}{"link": link})
		if err != nil {
			return router.SendError(c, fiber.StatusInternalServerError, err.Error())
		}
		return c.JSON(&fiber.Map{"status": "success"})
	}
}

func VerifyConfirm(p *pwBased) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		rawToken := c.Query("token")
		if rawToken == "" {
			return router.SendError(c, fiber.StatusNotFound, "token not found")
		}

		token, err := jwt.ParseJWT(rawToken)
		if err != nil {
			return router.SendError(c, fiber.StatusBadRequest, err.Error())
		}
		email, ok := token.Get("email")
		if !ok {
			return router.SendError(c, fiber.StatusBadRequest, "cannot get email from token")
		}
		if err := jwt.InvalidateJWT(token); err != nil {
			return router.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		_, err = p.manager.Update(
			&identity.Credential{
				Name:  identity.Email,
				Value: email.(string),
			},
			AdapterName,
			map[string]interface{}{identity.EmailVerified: true})
		if err != nil {
			return router.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		redirectUrl := c.Query("redirect_url")
		if redirectUrl != "" {
			return c.Redirect(redirectUrl)
		}
		return c.JSON(&fiber.Map{"status": "success"})
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
