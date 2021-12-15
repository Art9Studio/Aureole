package phone

import (
	"aureole/internal/core"
	"aureole/internal/plugins"

	"github.com/gofiber/fiber/v2"
)

func sendOTP(p *phone) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var input *input
		if err := c.BodyParser(input); err != nil {
			return core.SendError(c, fiber.StatusBadRequest, err.Error())
		}
		if input.Phone == "" {
			return core.SendError(c, fiber.StatusBadRequest, "phone required")
		}
		i := plugins.Identity{Phone: &input.Phone}

		randStr, err := core.GetRandStr(p.conf.Otp.Length, p.conf.Otp.Alphabet)
		if err != nil {
			return core.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		otp := p.conf.Otp.Prefix + randStr + p.conf.Otp.Postfix
		encOtp, err := core.Encrypt(p.conf.Otp.Prefix + randStr + p.conf.Otp.Postfix)
		if err != nil {
			return core.SendError(c, fiber.StatusInternalServerError, err.Error())
		}
		err = p.pluginApi.SaveToService(input.Phone, encOtp, p.conf.Otp.Exp)
		if err != nil {
			return core.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		token, err := core.CreateJWT(
			map[string]interface{}{
				"phone":    i.Phone,
				"attempts": 0,
			},
			p.conf.Otp.Exp)
		if err != nil {
			return core.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		err = p.sender.Send(*i.Phone, "", p.conf.Template, map[string]interface{}{"otp": otp})
		if err != nil {
			return core.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		return c.JSON(fiber.Map{"token": token})
	}
}

func resendOTP(p *phone) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var input *input
		if err := c.BodyParser(input); err != nil {
			return core.SendError(c, fiber.StatusBadRequest, err.Error())
		}
		if input.Token == "" {
			return core.SendError(c, fiber.StatusBadRequest, "token are required")
		}

		t, err := core.ParseJWT(input.Token)
		if err != nil {
			return core.SendError(c, fiber.StatusBadRequest, err.Error())
		}
		phone, ok := t.Get("phone")
		if !ok {
			return core.SendError(c, fiber.StatusBadRequest, "cannot get phone from token")
		}
		if err := core.InvalidateJWT(t); err != nil {
			return core.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		randStr, err := core.GetRandStr(p.conf.Otp.Length, p.conf.Otp.Alphabet)
		if err != nil {
			return core.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		otp := p.conf.Otp.Prefix + randStr + p.conf.Otp.Postfix
		encOtp, err := core.Encrypt(p.conf.Otp.Prefix + randStr + p.conf.Otp.Postfix)
		if err != nil {
			return core.SendError(c, fiber.StatusInternalServerError, err.Error())
		}
		err = p.pluginApi.SaveToService(phone.(string), encOtp, p.conf.Otp.Exp)
		if err != nil {
			return core.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		token, err := core.CreateJWT(
			map[string]interface{}{
				"phone":    phone,
				"attempts": 0,
			},
			p.conf.Otp.Exp)
		if err != nil {
			return core.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		err = p.sender.Send(phone.(string), "", p.conf.Template, map[string]interface{}{"otp": otp})
		if err != nil {
			return core.SendError(c, fiber.StatusInternalServerError, err.Error())
		}
		return c.JSON(&fiber.Map{"token": token})
	}
}
