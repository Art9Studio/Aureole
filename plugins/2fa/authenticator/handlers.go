package authenticator

import (
	"aureole/internal/core"
	"aureole/internal/identity"
	"aureole/pkg/dgoogauth"
	"encoding/base32"
	"github.com/gofiber/fiber/v2"
	"github.com/skip2/go-qrcode"
	"strings"
)

func getQR(g *gauth) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		// check if user authenticated:
		// yes -> generate new data and persist it, send user qr
		// no -> send error
		// only authenticated users and users, who doesn't yet enable 2fa, can get qr

		fa2Data := map[string]interface{}{}
		response := fiber.Map{}

		secret, err := generateSecret()
		if err != nil {
			return core.SendError(c, fiber.StatusInternalServerError, err.Error())
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
				return core.SendError(c, fiber.StatusInternalServerError, err.Error())
			}
			fa2Data["scratch_codes"] = scratchCodes
			response["scratch_code"] = scratchCodes
		}

		cred := &identity.Credential{}
		qr, err := qrcode.Encode(otp.ProvisionURIWithIssuer(cred.Value, g.conf.Iss), qrcode.Low, 256)
		if err != nil {
			return core.SendError(c, fiber.StatusInternalServerError, err.Error())
		}
		response["qr"] = qr

		if err := g.manager.On2FA(cred, fa2Data); err != nil {
			return core.SendError(c, fiber.StatusInternalServerError, err.Error())
		}
		return c.JSON(fiber.Map{"qr": qr})
	}
}

func getScratchCodes(g *gauth) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		// check if user already authenticated
		cred := &identity.Credential{Name: "email", Value: "www@example.com"}

		scratchCodes, err := generateScratchCodes(g.conf.ScratchCode.Num, g.conf.ScratchCode.Alphabet)
		if err != nil {
			return core.SendError(c, fiber.StatusInternalServerError, err.Error())
		}
		if err := g.manager.On2FA(cred, map[string]interface{}{"scratch_codes": scratchCodes}); err != nil {
			return core.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		return c.JSON(&fiber.Map{"scratch_codes": scratchCodes})
	}
}

func generateSecret() (string, error) {
	randStr, err := core.GetRandStr(8, "alphanum")
	if err != nil {
		return "", err
	}
	return base32.StdEncoding.EncodeToString([]byte(randStr)), nil
}

func generateScratchCodes(num int, alphabet string) (scratchCodes []string, err error) {
	scratchCodes = make([]string, num)
	for i := 0; i < num; i++ {
		scratchCodes[i], err = core.GetRandStr(8, alphabet)
		if err != nil {
			return nil, err
		}
	}
	return scratchCodes, err
}