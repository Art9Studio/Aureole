package sms

import (
	"aureole/internal/core"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v4"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func resend(s *sms) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var input *InitMFAReqBody
		if err := c.BodyParser(input); err != nil {
			return core.SendError(c, http.StatusBadRequest, err.Error())
		}
		if input.Token == "" {
			return core.SendError(c, http.StatusBadRequest, "token are required")
		}

		t, err := s.pluginAPI.ParseJWTService(input.Token)
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
		return core.SendToken(c, token)
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

		manager, ok := s.pluginAPI.GetIDManager()
		if !ok {
			return core.SendError(c, http.StatusInternalServerError, "cannot get id manager")
		}

		cred := &core.Credential{Name: "phone", Value: phone.Phone}

		MFAEnabled, err := manager.IsMFAEnabled(cred)
		if MFAEnabled {
			return core.SendError(c, http.StatusBadRequest, "sms mfa already enabled")
		} else if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return core.SendError(c, http.StatusInternalServerError, err.Error())
		}

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
				"phone":    phone.Phone,
				"attempts": 0,
			},
			s.conf.Otp.Exp)
		if err != nil {
			return core.SendError(c, http.StatusInternalServerError, err.Error())
		}
		fmt.Println(otp)
		//err = s.sender.Send(phone.Phone, "", s.tmpl, s.tmplExt, map[string]interface{}{"otp": otp})
		//if err != nil {
		//	return core.SendError(c, http.StatusInternalServerError, err.Error())
		//}

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

		idRaw := ctx.Locals(core.UserID)
		fmt.Println(idRaw)
		id, ok := idRaw.(string)
		if !ok {
			return core.SendError(ctx, http.StatusInternalServerError, "cannot get user id")
		}

		newCred := &core.Credential{Name: "id", Value: id}
		//todo(Talgat) Update updates EmailVerified as false, if not explicitly indicated in Identity
		_, err = manager.Update(newCred, &core.Identity{Phone: &cred.Value, PhoneVerified: true}, "dummy")
		if err != nil {
			return core.SendError(ctx, http.StatusInternalServerError, err.Error())
		}

		if err = manager.OnMFA(cred, &core.MFAData{
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
			return core.SendError(ctx, http.StatusForbidden, err.Error())
		}

		var id string
		if err = s.pluginAPI.GetFromJWT(token, "sub", &id); err != nil {
			return ctx.SendStatus(http.StatusForbidden)
		}
		ctx.Locals(core.UserID, id)

		return h(ctx)
	}
}
