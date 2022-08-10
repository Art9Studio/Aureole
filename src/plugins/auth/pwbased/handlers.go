package pwbased

import (
	"aureole/internal/core"
	"errors"
	"github.com/gofiber/fiber/v2"
	"net/http"
	"net/url"
)

func register(p *pwBased) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var rawCred *RegisterReqBody
		if err := c.BodyParser(rawCred); err != nil {
			return core.SendError(c, http.StatusBadRequest, err.Error())
		}
		if rawCred.Password == "" {
			return core.SendError(c, http.StatusBadRequest, "password required")
		}

		pwHash, err := p.pwHasher.HashPw(rawCred.Password)
		if err != nil {
			return core.SendError(c, http.StatusInternalServerError, err.Error())
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
			return core.SendError(c, http.StatusBadRequest, err.Error())
		}
		manager, ok := p.pluginAPI.GetIDManager()
		if !ok {
			return core.SendError(c, http.StatusInternalServerError, "could not get ID manager")
		}
		user, err := manager.Register(cred, i, nil, meta.ShortName)
		if err != nil {
			return err
		}
		_ = user

		if p.conf.Register.IsVerifyAfter {
			token, err := p.pluginAPI.CreateJWT(map[string]interface{}{"email": rawCred.Email}, p.conf.Verify.Exp)
			if err != nil {
				return core.SendError(c, http.StatusInternalServerError, err.Error())
			}
			link := attachToken(p.verify.confirmLink, token)

			err = p.verify.sender.Send(rawCred.Email, "", p.verify.tmpl, p.verify.tmplExt, map[string]interface{}{"link": link})
			if err != nil {
				return core.SendError(c, http.StatusInternalServerError, err.Error())
			}
		}

		/*if p.conf.register.IsLoginAfter {
			payload, err := authzT.NewIssuerPayload(user)
			if err != nil {
				return router.SendError(c, http.StatusInternalServerError, err.Error())
			}
			return p.authorizer.Authorize(c, payload)
		} else {*/
		return c.SendStatus(http.StatusOK)
		//}
	}
}

func Reset(p *pwBased) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var e *ResetReqBody
		if err := c.BodyParser(e); err != nil {
			return core.SendError(c, http.StatusBadRequest, err.Error())
		}
		if e.Email == "" {
			return core.SendError(c, http.StatusBadRequest, "email required")
		}
		i := &core.Identity{Email: &e.Email}

		token, err := p.pluginAPI.CreateJWT(map[string]interface{}{"email": i.Email}, p.conf.Reset.Exp)
		if err != nil {
			return core.SendError(c, http.StatusInternalServerError, err.Error())
		}

		link := attachToken(p.reset.confirmLink, token)
		err = p.reset.sender.Send(*i.Email, "", p.reset.tmpl, p.reset.tmplExt, map[string]interface{}{"link": link})
		if err != nil {
			return core.SendError(c, http.StatusInternalServerError, err.Error())
		}
		return c.SendStatus(http.StatusOK)
	}
}

func ResetConfirm(p *pwBased) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		query := &ResetConfirmQuery{}
		if err := c.QueryParser(query); err != nil {
			return core.SendError(c, http.StatusBadRequest, "invalid format")
		}
		var input *ResetConfirmReqBody
		if err := c.BodyParser(input); err != nil {
			return core.SendError(c, http.StatusBadRequest, err.Error())
		}

		rawToken := query.Token
		if rawToken == "" {
			return core.SendError(c, http.StatusBadRequest, "token not found")
		}
		if input.Password == "" {
			return core.SendError(c, http.StatusBadRequest, "password required")
		}

		token, err := p.pluginAPI.ParseJWTService(rawToken)
		if err != nil {
			return core.SendError(c, http.StatusBadRequest, err.Error())
		}
		email, ok := token.Get("email")
		if !ok {
			return core.SendError(c, http.StatusBadRequest, "cannot get email from token")
		}
		if err := p.pluginAPI.InvalidateJWT(token); err != nil {
			return core.SendError(c, http.StatusInternalServerError, err.Error())
		}

		pwHash, err := p.pwHasher.HashPw(input.Password)
		if err != nil {
			return core.SendError(c, http.StatusInternalServerError, err.Error())
		}

		manager, ok := p.pluginAPI.GetIDManager()
		if !ok {
			return core.SendError(c, http.StatusInternalServerError, "could not get ID manager")
		}

		_, err = manager.Update(
			&core.Credential{
				Name:  core.Email,
				Value: email.(string),
			},
			&core.Identity{
				Additional: map[string]interface{}{core.Password: pwHash},
			},
			meta.ShortName)
		if err != nil {
			return core.SendError(c, http.StatusInternalServerError, err.Error())
		}

		err = p.reset.sender.SendRaw(email.(string), "Reset your password",
			"Your password has been successfully changed")
		if err != nil {
			return core.SendError(c, http.StatusInternalServerError, err.Error())
		}

		// todo: add expiring any current user session
		redirectUrl := query.URL
		if redirectUrl != "" {
			return c.Redirect(redirectUrl)
		}
		return c.SendStatus(http.StatusOK)
	}
}

func Verify(p *pwBased) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var e *VerifyReqBody
		if err := c.BodyParser(e); err != nil {
			return core.SendError(c, http.StatusBadRequest, err.Error())
		}
		if e.Email == "" {
			return core.SendError(c, http.StatusBadRequest, "email required")
		}
		i := &core.Identity{Email: &e.Email}

		token, err := p.pluginAPI.CreateJWT(map[string]interface{}{"email": i.Email}, p.conf.Verify.Exp)
		if err != nil {
			return core.SendError(c, http.StatusInternalServerError, err.Error())
		}

		link := attachToken(p.verify.confirmLink, token)
		err = p.verify.sender.Send(*i.Email, "", p.verify.tmpl, p.verify.tmplExt, map[string]interface{}{"link": link})
		if err != nil {
			return core.SendError(c, http.StatusInternalServerError, err.Error())
		}
		return c.SendStatus(http.StatusOK)
	}
}

func VerifyConfirm(p *pwBased) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		query := &VerifyConfirmQuery{}
		err := c.QueryParser(query)
		if err != nil {
			return err
		}
		rawToken := query.Token
		if rawToken == "" {
			return core.SendError(c, http.StatusBadRequest, "token not found")
		}

		token, err := p.pluginAPI.ParseJWTService(rawToken)
		if err != nil {
			return core.SendError(c, http.StatusBadRequest, err.Error())
		}
		email, ok := token.Get("email")
		if !ok {
			return core.SendError(c, http.StatusBadRequest, "cannot get email from token")
		}
		if err := p.pluginAPI.InvalidateJWT(token); err != nil {
			return core.SendError(c, http.StatusInternalServerError, err.Error())
		}

		manager, ok := p.pluginAPI.GetIDManager()
		if !ok {
			return core.SendError(c, http.StatusInternalServerError, "could not get ID manager")
		}

		_, err = manager.Update(
			&core.Credential{
				Name:  core.Email,
				Value: email.(string),
			},
			&core.Identity{EmailVerified: true},
			meta.ShortName)
		if err != nil {
			return core.SendError(c, http.StatusInternalServerError, err.Error())
		}

		redirectUrl := query.URL
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
