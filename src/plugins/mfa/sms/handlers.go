package sms

import (
	"aureole/internal/core"
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

func resend(s *sms) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var input *Init2FAReqBody
		if err := c.BodyParser(input); err != nil {
			return core.SendError(c, http.StatusBadRequest, err.Error())
		}
		if input.Token == "" {
			return core.SendError(c, http.StatusBadRequest, "token are required")
		}

		t, err := s.pluginAPI.ParseJWT(input.Token)
		if err != nil {
			return core.SendError(c, http.StatusBadRequest, err.Error())
		}
		phone, ok := t.Get("phone")
		if !ok {
			return core.SendError(c, http.StatusBadRequest, "cannot get phone from token")
		}
		if err := s.pluginAPI.InvalidateJWT(t); err != nil {
			return core.SendError(c, http.StatusInternalServerError, err.Error())
		}

		otp, err := s.pluginAPI.GetRandStr(s.conf.Otp.Length, s.conf.Otp.Alphabet)
		if err != nil {
			return core.SendError(c, http.StatusInternalServerError, err.Error())
		}

		token, err := s.pluginAPI.CreateJWT(
			map[string]interface{}{
				"phone":    phone,
				"attempts": 0,
			},
			s.conf.Otp.Exp)
		if err != nil {
			return core.SendError(c, http.StatusInternalServerError, err.Error())
		}

		encOtp, err := s.pluginAPI.Encrypt(otp)
		if err != nil {
			return core.SendError(c, http.StatusInternalServerError, err.Error())
		}
		_ = encOtp
		err = s.pluginAPI.SaveToService(phone.(string), encOtp, s.conf.Otp.Exp)
		if err != nil {
			return core.SendError(c, http.StatusInternalServerError, err.Error())
		}

		err = s.sender.Send(phone.(string), "", s.tmpl, s.tmplExt, map[string]interface{}{"otp": otp})
		if err != nil {
			return core.SendError(c, http.StatusInternalServerError, err.Error())
		}
		return c.JSON(&fiber.Map{"token": token})
	}
}

func initMFASMS(s *sms) func(*fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		cred, _, err := s.Verify()(*ctx)

		if err != nil {
			//todo(Talgat) handle 500 and 400 errors
			return core.SendError(ctx, http.StatusInternalServerError, err.Error())
		}

		manager, ok := s.pluginAPI.GetIDManager()
		if !ok {
			return core.SendError(ctx, http.StatusInternalServerError, "cannot get IDManager")
		}

		if err = manager.On2FA(cred, &core.MFAData{
			PluginID:     fmt.Sprintf("%d", meta.PluginID),
			ProviderName: meta.ShortName,
		}); err != nil {
			return core.SendError(ctx, http.StatusInternalServerError, err.Error())
		}

		return nil
	}
}
