package phone

import (
	"aureole/internal/core"
	"github.com/gofiber/fiber/v2"
	"net/http"
)

func sendOTP(p *authn) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var phone sendOTPReqBody
		if err := c.BodyParser(&phone); err != nil {
			return core.SendError(c, http.StatusBadRequest, err.Error())
		}
		if phone.Phone == "" {
			return core.SendError(c, http.StatusBadRequest, "phone required")
		}
		u := core.User{Phone: &phone.Phone}

		randStr, err := p.pluginAPI.GetRandStr(p.conf.Otp.Length, p.conf.Otp.Alphabet)
		if err != nil {
			return core.SendError(c, http.StatusInternalServerError, err.Error())
		}

		otp := p.conf.Otp.Prefix + randStr + p.conf.Otp.Postfix
		encOtp, err := p.pluginAPI.Encrypt(p.conf.Otp.Prefix + randStr + p.conf.Otp.Postfix)
		if err != nil {
			return core.SendError(c, http.StatusInternalServerError, err.Error())
		}
		err = p.pluginAPI.SaveToService(phone.Phone, encOtp, p.conf.Otp.Exp)
		if err != nil {
			return core.SendError(c, http.StatusInternalServerError, err.Error())
		}

		token, err := p.pluginAPI.CreateJWT(
			map[string]interface{}{
				"phone":    u.Phone,
				"attempts": 0,
			},
			p.conf.Otp.Exp)
		if err != nil {
			return core.SendError(c, http.StatusInternalServerError, err.Error())
		}

		err = p.sender.Send(*u.Phone, "", p.tmpl, p.tmplExt, map[string]interface{}{"otp": otp})
		if err != nil {
			return core.SendError(c, http.StatusInternalServerError, err.Error())
		}

		return c.JSON(&OTPResponse{Token: token})
	}
}

func resendOTP(p *authn) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var input *resendOTPReqBody
		if err := c.BodyParser(input); err != nil {
			return core.SendError(c, http.StatusBadRequest, err.Error())
		}
		if input.Token == "" {
			return core.SendError(c, http.StatusBadRequest, "token are required")
		}

		t, err := p.pluginAPI.ParseJWTService(input.Token)
		if err != nil {
			return core.SendError(c, http.StatusBadRequest, err.Error())
		}
		phone, ok := t.Get(core.Phone)
		if !ok {
			return core.SendError(c, http.StatusBadRequest, "cannot get phone from token")
		}
		if err := p.pluginAPI.InvalidateJWT(t); err != nil {
			return core.SendError(c, http.StatusInternalServerError, err.Error())
		}

		randStr, err := p.pluginAPI.GetRandStr(p.conf.Otp.Length, p.conf.Otp.Alphabet)
		if err != nil {
			return core.SendError(c, http.StatusInternalServerError, err.Error())
		}

		otp := p.conf.Otp.Prefix + randStr + p.conf.Otp.Postfix
		encOtp, err := p.pluginAPI.Encrypt(p.conf.Otp.Prefix + randStr + p.conf.Otp.Postfix)
		if err != nil {
			return core.SendError(c, http.StatusInternalServerError, err.Error())
		}
		err = p.pluginAPI.SaveToService(phone.(string), encOtp, p.conf.Otp.Exp)
		if err != nil {
			return core.SendError(c, http.StatusInternalServerError, err.Error())
		}

		token, err := p.pluginAPI.CreateJWT(
			map[string]interface{}{
				"phone":    phone,
				"attempts": 0,
			},
			p.conf.Otp.Exp)
		if err != nil {
			return core.SendError(c, http.StatusInternalServerError, err.Error())
		}

		err = p.sender.Send(phone.(string), "", p.tmpl, p.tmplExt, map[string]interface{}{"otp": otp})
		if err != nil {
			return core.SendError(c, http.StatusInternalServerError, err.Error())
		}
		return c.JSON(&OTPResponse{Token: token})
	}
}
