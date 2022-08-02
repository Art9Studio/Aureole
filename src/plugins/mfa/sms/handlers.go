package sms

import (
	"aureole/internal/core"
	"fmt"
	"net/http"
	"strings"

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

type sendOTPReqBody struct {
	Phone string `json:"phone"`
}

type OTPResponse struct {
	Token string `json:"token"`
}

func sendOTP(s *sms) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var phone sendOTPReqBody
		if err := c.BodyParser(&phone); err != nil {
			return core.SendError(c, http.StatusBadRequest, err.Error())
		}
		if phone.Phone == "" {
			return core.SendError(c, http.StatusBadRequest, "phone required")
		}
		i := core.Identity{Phone: &phone.Phone}

		randStr, err := s.pluginAPI.GetRandStr(s.conf.Otp.Length, s.conf.Otp.Alphabet)
		if err != nil {
			return core.SendError(c, http.StatusInternalServerError, err.Error())
		}

		otp := s.conf.Otp.Prefix + randStr + s.conf.Otp.Postfix
		encOtp, err := s.pluginAPI.Encrypt(s.conf.Otp.Prefix + randStr + s.conf.Otp.Postfix)
		if err != nil {
			return core.SendError(c, http.StatusInternalServerError, err.Error())
		}
		err = s.pluginAPI.SaveToService(phone.Phone, encOtp, s.conf.Otp.Exp)
		if err != nil {
			return core.SendError(c, http.StatusInternalServerError, err.Error())
		}

		token, err := s.pluginAPI.CreateJWT(
			map[string]interface{}{
				"phone":    i.Phone,
				"attempts": 0,
			},
			s.conf.Otp.Exp)
		if err != nil {
			return core.SendError(c, http.StatusInternalServerError, err.Error())
		}
		fmt.Println(otp)
		err = s.sender.Send(*i.Phone, "", s.tmpl, s.tmplExt, map[string]interface{}{"otp": otp})
		if err != nil {
			return core.SendError(c, http.StatusInternalServerError, err.Error())
		}

		return c.JSON(&OTPResponse{Token: token})
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

		return ctx.SendStatus(http.StatusOK)
	}
}

func authMiddleware(s *sms, h fiber.Handler) func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		bearer := ctx.Get(fiber.HeaderAuthorization)
		tokenSplit := strings.Split(bearer, "Bearer ")

		var rawToken string
		if len(tokenSplit) == 2 && tokenSplit[1] != "" {
			rawToken = tokenSplit[1]
		} else {
			return ctx.SendStatus(http.StatusForbidden)
		}

		token, err := s.pluginAPI.ParseJWT(rawToken)
		if err != nil {
			return ctx.SendStatus(http.StatusForbidden)
		}

		var id string
		if err = s.pluginAPI.GetFromJWT(token, "ID", &id); err != nil {
			return ctx.SendStatus(http.StatusForbidden)
		}
		ctx.Locals(core.UserID, id)

		return h(ctx)
	}
}
