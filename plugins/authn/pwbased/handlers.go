package pwbased

import (
	"aureole/internal/identity"
	"aureole/internal/jwt"
	authzT "aureole/internal/plugins/authz/types"
	"fmt"
	"github.com/gofiber/fiber/v2"
)

func Login(p *pwBased) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var input *input
		if err := c.BodyParser(input); err != nil {
			return sendError(c, fiber.StatusBadRequest, err.Error())
		}
		if input.Password == "" {
			return sendError(c, fiber.StatusBadRequest, "password required")
		}

		i := &identity.Identity{
			Id:       input.Id,
			Email:    input.Email,
			Phone:    input.Phone,
			Username: input.Username,
		}
		cred, err := getCredential(i)
		if err != nil {
			return sendError(c, fiber.StatusBadRequest, err.Error())
		}

		pw, err := p.manager.GetData(cred, AdapterName, "password")
		if err != nil {
			return sendError(c, fiber.StatusBadRequest, err.Error())
		}

		isMatch, err := p.pwHasher.ComparePw(input.Password, pw.(string))
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		if isMatch {
			userData, err := p.manager.OnUserAuthenticated(cred, i, AdapterName, nil)
			if err != nil {
				return sendError(c, fiber.StatusInternalServerError, err.Error())
			}

			return p.authorizer.Authorize(c, authzT.NewPayload(p.authorizer, nil, userData))
		} else {
			return sendError(c, fiber.StatusUnauthorized, fmt.Sprintf("wrong password or %s", cred.Name))
		}
	}
}

func Register(p *pwBased) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var input *input
		if err := c.BodyParser(input); err != nil {
			return sendError(c, fiber.StatusBadRequest, err.Error())
		}
		if input.Password == "" {
			return sendError(c, fiber.StatusBadRequest, "password required")
		}

		i := &identity.Identity{
			Id:       input.Id,
			Username: input.Username,
			Phone:    input.Phone,
			Email:    input.Email,
		}
		pwHash, err := p.pwHasher.HashPw(input.Password)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		cred, err := getCredential(i)
		if err != nil {
			return sendError(c, fiber.StatusBadRequest, err.Error())
		}
		user, err := p.manager.OnRegister(cred, i, AdapterName, map[string]interface{}{"password": pwHash})
		if err != nil {
			return err
		}

		if p.conf.Register.IsVerifyAfter {
			token, err := jwt.CreateJWT(map[string]interface{}{"email": input.Email}, p.conf.Verif.Exp)
			if err != nil {
				return sendError(c, fiber.StatusInternalServerError, err.Error())
			}
			link := attachToken(p.verif.confirmLink, token)

			err = p.verif.sender.Send(input.Email, "", p.conf.Verif.Template, map[string]interface{}{"link": link})
			if err != nil {
				return sendError(c, fiber.StatusInternalServerError, err.Error())
			}
		}

		if p.conf.Register.IsLoginAfter {
			return p.authorizer.Authorize(c, authzT.NewPayload(p.authorizer, nil, user))
		} else {
			return c.JSON(&fiber.Map{"status": "success"})
		}
	}
}

func Reset(p *pwBased) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var input *input
		if err := c.BodyParser(input); err != nil {
			return sendError(c, fiber.StatusBadRequest, err.Error())
		}
		if input.Email == "" {
			return sendError(c, fiber.StatusBadRequest, "email required")
		}
		i := &identity.Identity{Email: input.Email}

		token, err := jwt.CreateJWT(map[string]interface{}{"email": i.Email}, p.conf.Reset.Exp)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		link := attachToken(p.reset.confirmLink, token)
		err = p.verif.sender.Send(i.Email, "", p.conf.Verif.Template, map[string]interface{}{"link": link})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}
		return c.JSON(&fiber.Map{"status": "success"})
	}
}

func ResetConfirm(p *pwBased) func(*fiber.Ctx) error {
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

		var input *input
		if err := c.BodyParser(input); err != nil {
			return sendError(c, fiber.StatusBadRequest, err.Error())
		}
		if input.Password == "" {
			return sendError(c, fiber.StatusBadRequest, "password required")
		}

		pwHash, err := p.pwHasher.HashPw(input.Password)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		_, err = p.manager.Update(
			&identity.Credential{
				Name:  "email",
				Value: email},
			AdapterName,
			map[string]interface{}{"password": pwHash})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		err = p.reset.sender.SendRaw(email.(string), "Reset your password",
			"Your password has been successfully changed")
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
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
			return sendError(c, fiber.StatusBadRequest, err.Error())
		}
		if input.Email == "" {
			return sendError(c, fiber.StatusBadRequest, "email required")
		}
		i := &identity.Identity{Email: input.Email}

		token, err := jwt.CreateJWT(map[string]interface{}{"email": i.Email}, p.conf.Verif.Exp)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		link := attachToken(p.verif.confirmLink, token)
		err = p.verif.sender.Send(i.Email, "", p.conf.Verif.Template, map[string]interface{}{"link": link})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}
		return c.JSON(&fiber.Map{"status": "success"})
	}
}

func VerifyConfirm(p *pwBased) func(*fiber.Ctx) error {
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

		_, err = p.manager.Update(
			&identity.Credential{
				Name:  "email",
				Value: email,
			},
			AdapterName,
			map[string]interface{}{"email_verified": true})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		redirectUrl := c.Query("redirect_url")
		if redirectUrl != "" {
			return c.Redirect(redirectUrl)
		}
		return c.JSON(&fiber.Map{"status": "success"})
	}
}
