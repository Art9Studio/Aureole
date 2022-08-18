package pwbased

import (
	"aureole/internal/core"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"net/http"
	"net/url"
)

func register(p *pwBased) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		rawCred := &RegisterReqBody{}
		if err := c.BodyParser(rawCred); err != nil {
			return core.SendError(c, http.StatusBadRequest, err.Error())
		}
		if rawCred.Password == "" || rawCred.Email == "" {
			return core.SendError(c, http.StatusBadRequest, "password and email required")
		}

		pwHash, err := p.pwHasher.HashPw(rawCred.Password)
		if err != nil {
			return core.SendError(c, http.StatusInternalServerError, err.Error())
		}

		var (
			u = &core.User{
				Username: &rawCred.Username,
				Phone:    &rawCred.Phone,
				Email:    &rawCred.Email,
			}
			secret = &core.Secrets{password: pwHash}
			cred   = &core.Credential{Name: "email", Value: rawCred.Email}
		)

		manager, ok := p.pluginAPI.GetIDManager()
		if !ok {
			return core.SendError(c, http.StatusInternalServerError, "could not get ID manager")
		}

		user, err := manager.RegisterOrUpdate(&core.AuthResult{
			Cred: cred, User: u,
			Secrets:    secret,
			ProviderId: fmt.Sprintf("%d", meta.PluginID),
		})
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
		e := &ResetReqBody{}
		if err := c.BodyParser(e); err != nil {
			return core.SendError(c, http.StatusBadRequest, err.Error())
		}
		if e.Email == "" {
			return core.SendError(c, http.StatusBadRequest, "email required")
		}

		token, err := p.pluginAPI.CreateJWT(map[string]interface{}{"email": e.Email}, p.conf.Reset.Exp)
		if err != nil {
			return core.SendError(c, http.StatusInternalServerError, err.Error())
		}

		link := attachToken(p.reset.confirmLink, token)
		err = p.reset.sender.Send(e.Email, "", p.reset.tmpl, p.reset.tmplExt, map[string]interface{}{"link": link})
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
		input := &ResetConfirmReqBody{}
		if err := c.BodyParser(input); err != nil {
			return core.SendError(c, http.StatusBadRequest, err.Error())
		}

		rawToken := query.Token
		if rawToken == "" {
			return core.SendError(c, http.StatusBadRequest, "token not found")
		}
		if input.Password == "" {
			return core.SendError(c, http.StatusBadRequest, "password and email required")
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

		emailStr, _ := email.(string)
		if _, err = manager.RegisterOrUpdate(
			&core.AuthResult{
				User: &core.User{
					Email: &emailStr,
				},
				Cred: &core.Credential{
					Name:  core.Email,
					Value: email.(string),
				},
				Secrets: &core.Secrets{core.Password: pwHash},
			}); err != nil {
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
		e := &VerifyReqBody{}
		if err := c.BodyParser(e); err != nil {
			return core.SendError(c, http.StatusBadRequest, err.Error())
		}
		if e.Email == "" {
			return core.SendError(c, http.StatusBadRequest, "email required")
		}

		token, err := p.pluginAPI.CreateJWT(map[string]interface{}{"email": e.Email}, p.conf.Verify.Exp)
		if err != nil {
			return core.SendError(c, http.StatusInternalServerError, err.Error())
		}

		link := attachToken(p.verify.confirmLink, token)
		err = p.verify.sender.Send(e.Email, "", p.verify.tmpl, p.verify.tmplExt, map[string]interface{}{"link": link})
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

		emailStr, _ := email.(string)
		_, err = manager.RegisterOrUpdate(
			&core.AuthResult{
				Cred: &core.Credential{
					Name:  core.Email,
					Value: email.(string),
				},
				User: &core.User{
					Email:         &emailStr,
					EmailVerified: true},
			},
		)
		if err != nil {
			return core.SendError(c, http.StatusInternalServerError, err.Error())
		}

		redirectUrl := query.URL
		if redirectUrl != "" {
			return c.Redirect(redirectUrl)
		}
		return c.JSON(VerifyConfirmRes{Success: true})
	}
}

func attachToken(u *url.URL, token string) string {
	q := u.Query()
	q.Set("token", token)
	u.RawQuery = q.Encode()

	return u.String()
}
