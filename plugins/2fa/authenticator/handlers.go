package authenticator

import (
	"aureole/internal/identity"
	"aureole/pkg/dgoogauth"
	"github.com/gofiber/fiber/v2"
	"github.com/skip2/go-qrcode"
	"strconv"
	"strings"
)

func GetQR(g *gauth) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		// check if user authenticated:
		// yes -> generate new data and persist it, send user qr
		// no -> send error
		// only authenticated users and users, who doesn't yet enable 2fa, can get qr

		fa2Data := map[string]interface{}{}
		responseJson := fiber.Map{}

		cred := &identity.Credential{Name: identity.Email, Value: "www@example.com"}
		ok, err := g.pluginApi.Is2FactorEnabled(cred, "pwbased")
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}
		if ok {
			return sendError(c, fiber.StatusInternalServerError, "two factor auth is already enabled")
		}

		secret, err := generateSecret()
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}
		fa2Data["secret"] = secret

		otp := &dgoogauth.OTPConfig{Secret: strings.TrimSpace(secret)}
		if g.conf.Alg == "hotp" {
			otp.HotpCounter = 1
			fa2Data["counter"] = 1
		}
		if g.conf.ScratchCode.Num != 0 {
			scratchCodes, err := generateScratchCodes(g.conf.ScratchCode.Num, g.conf.ScratchCode.Alphabet)
			if err != nil {
				return sendError(c, fiber.StatusInternalServerError, err.Error())
			}
			fa2Data["scratch_codes"] = scratchCodes
			responseJson["scratch_code"] = scratchCodes
		}

		qr, err := qrcode.Encode(otp.ProvisionURIWithIssuer(cred.Value, g.conf.Iss), qrcode.Low, 256)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}
		responseJson["qr"] = qr

		if err := g.manager.On2FA(cred, fa2Data); err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}
		return c.JSON(&responseJson)
	}
}

func VerifyOTP(g *gauth) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		cred := &identity.Credential{Name: "email", Value: "www@example.com"}
		provider := "pwbased"

		ok, err := g.pluginApi.Is2FactorEnabled(cred, provider)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}
		if !ok {
			return sendError(c, fiber.StatusInternalServerError, "two factor auth doesn't enabled")
		}

		var input *input
		if err := c.BodyParser(input); err != nil {
			return sendError(c, fiber.StatusBadRequest, err.Error())
		}
		if input.Otp == "" {
			return sendError(c, fiber.StatusBadRequest, "otp is required")
		}

		secret, err := g.manager.GetData(cred, provider, "secret")
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}
		scratchCodes, err := g.manager.GetData(cred, provider, "scratch_codes")
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		var counter int
		if g.conf.Alg == "hotp" {
			rawCounter, err := g.manager.GetData(cred, provider, "counter")
			if err != nil {
				return sendError(c, fiber.StatusInternalServerError, err.Error())
			}
			counter = rawCounter.(int)
		}

		var usedOtp []int
		_, err = g.pluginApi.GetFromService(cred.Value, &usedOtp)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		otp := &dgoogauth.OTPConfig{
			Secret:        secret.(string),
			WindowSize:    g.conf.WindowSize,
			HotpCounter:   counter,
			DisallowReuse: usedOtp,
			ScratchCodes:  scratchCodes.([]string),
		}
		ok, err = otp.Authenticate(strings.TrimSpace(input.Otp))
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}
		if !ok {
			return sendError(c, fiber.StatusUnauthorized, "wrong otp")
		}
		if err := g.manager.On2FA(cred, map[string]interface{}{
			"counter": otp.HotpCounter, "scratch_code": otp.ScratchCodes,
		}); err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		if g.conf.DisallowReuse {
			if usedOtp == nil {
				usedOtp = make([]int, 1)
			}
			intOtp, err := strconv.Atoi(input.Otp)
			if err != nil {
				return sendError(c, fiber.StatusInternalServerError, err.Error())
			}

			usedOtp = append(usedOtp, intOtp)
			if err := g.pluginApi.SaveToService(cred.Value, usedOtp, 1); err != nil {
				return sendError(c, fiber.StatusInternalServerError, err.Error())
			}
		}

		return c.JSON(&fiber.Map{"status": "success"})
	}
}

func GetScratchCodes(g *gauth) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		// check if user already authenticated
		cred := &identity.Credential{Name: "email", Value: "www@example.com"}

		scratchCodes, err := generateScratchCodes(g.conf.ScratchCode.Num, g.conf.ScratchCode.Alphabet)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}
		if err := g.manager.On2FA(cred, map[string]interface{}{"scratch_codes": scratchCodes}); err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		return c.JSON(&fiber.Map{"scratch_codes": scratchCodes})
	}
}
