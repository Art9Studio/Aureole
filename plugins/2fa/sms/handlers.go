package sms

import (
	"aureole/internal/encrypt"
	"aureole/internal/jwt"
	"github.com/gofiber/fiber/v2"
)

func SendOtp(s *sms) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		phone := ""

		otp, err := getRandomString(s.conf.Otp.Length, s.conf.Otp.Alphabet)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		token, err := jwt.CreateJWT(
			map[string]interface{}{
				"phone":    phone,
				"attempts": 0,
			},
			s.conf.Otp.Exp)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		encOtp, err := encrypt.Encrypt(otp)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}
		_ = encOtp
		err = s.pluginApi.SaveToService(phone, encOtp, s.conf.Otp.Exp)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		err = s.sender.Send(phone, "", s.conf.Template, map[string]interface{}{"otp": otp})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		return c.JSON(&fiber.Map{"token": token})
	}
}

func Verify(s *sms) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var input *input
		if err := c.BodyParser(input); err != nil {
			return sendError(c, fiber.StatusBadRequest, err.Error())
		}
		if input.Token == "" || input.Otp == "" {
			return sendError(c, fiber.StatusBadRequest, "token and otp are required")
		}

		t, err := jwt.ParseJWT(input.Token)
		if err != nil {
			return sendError(c, fiber.StatusBadRequest, err.Error())
		}
		phone, ok := t.Get("phone")
		if !ok {
			return sendError(c, fiber.StatusBadRequest, "cannot get otp from token")
		}
		attempts, ok := t.Get("attempts")
		if !ok {
			return sendError(c, fiber.StatusBadRequest, "cannot get attempts from token")
		}
		if err := jwt.InvalidateJWT(t); err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		if int(attempts.(float64)) >= s.conf.MaxAttempts {
			return sendError(c, fiber.StatusUnauthorized, "too much attempts")
		}

		var (
			encOtp  []byte
			decrOtp string
		)
		ok, err = s.pluginApi.GetFromService(phone.(string), &encOtp)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}
		if !ok {
			return sendError(c, fiber.StatusUnauthorized, "otp has expired")
		}
		err = encrypt.Decrypt(encOtp, &decrOtp)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		if decrOtp == input.Otp {
			return c.JSON(&fiber.Map{"status": "success"})
		} else {
			token, err := jwt.CreateJWT(
				map[string]interface{}{
					"phone":    phone,
					"attempts": int(attempts.(float64)) + 1,
				},
				s.conf.Otp.Exp)
			if err != nil {
				return sendError(c, fiber.StatusInternalServerError, err.Error())
			}
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"token": token})
		}
	}
}

func Resend(s *sms) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var input *input
		if err := c.BodyParser(input); err != nil {
			return sendError(c, fiber.StatusBadRequest, err.Error())
		}
		if input.Token == "" {
			return sendError(c, fiber.StatusBadRequest, "token are required")
		}

		t, err := jwt.ParseJWT(input.Token)
		if err != nil {
			return sendError(c, fiber.StatusBadRequest, err.Error())
		}
		phone, ok := t.Get("phone")
		if !ok {
			return sendError(c, fiber.StatusBadRequest, "cannot get phone from token")
		}
		if err := jwt.InvalidateJWT(t); err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		otp, err := getRandomString(s.conf.Otp.Length, s.conf.Otp.Alphabet)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		token, err := jwt.CreateJWT(
			map[string]interface{}{
				"phone":    phone,
				"attempts": 0,
			},
			s.conf.Otp.Exp)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		encOtp, err := encrypt.Encrypt(otp)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}
		_ = encOtp
		err = s.pluginApi.SaveToService(phone.(string), encOtp, s.conf.Otp.Exp)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		err = s.sender.Send(phone.(string), "", s.conf.Template, map[string]interface{}{"otp": otp})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}
		return c.JSON(&fiber.Map{"token": token})
	}
}
