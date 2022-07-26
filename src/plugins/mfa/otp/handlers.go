package authenticator

import (
	"aureole/internal/core"
	"aureole/pkg/dgoogauth"
	"encoding/base32"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/skip2/go-qrcode"
	"net/http"
	"strings"
)

func getQR(g *otpAuth) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		// check if user authenticated:
		// yes -> generate new data and persist it, send user qr
		// no -> send error
		// only authenticated users and users, who doesn't yet enable 2fa, can get qr
		in := &getQRReqBody{}
		var idName, idValue string
		if err := c.BodyParser(in); err != nil {
			return err
		}
		if in.Email == "" && in.Phone == "" {
			return errors.New("email or phone is required")
		}
		if in.Email != "" {
			idName = "email"
			idValue = in.Email
		} else {
			idName = "phone"
			idValue = in.Phone
		}
		cred := &core.Credential{Name: idName, Value: idValue}
		ok, err := g.IsEnabled(cred)
		if err != nil {
			return err
		}
		if !ok {
			return errors.New("mfa already enabled")
		}

		mfaData := map[string]interface{}{}
		response := fiber.Map{}

		secret, err := generateSecret(g.pluginAPI)
		if err != nil {
			return core.SendError(c, http.StatusInternalServerError, err.Error())
		}
		mfaData["secret"] = secret

		otp := &dgoogauth.OTPConfig{Secret: strings.TrimSpace(secret)}
		if g.conf.Alg == "hotp" {
			otp.HotpCounter = 1
			mfaData["counter"] = 1
		}
		//if g.conf.ScratchCode.Num != 0 {
		//	scratchCodes, err := generateScratchCodes(g.pluginAPI, g.conf.ScratchCode.Num, g.conf.ScratchCode.Alphabet)
		//	if err != nil {
		//		return core.SendError(c, http.StatusInternalServerError, err.Error())
		//	}
		//	mfaData["scratch_codes"] = scratchCodes
		//	response["scratch_code"] = scratchCodes
		//}

		qr, err := qrcode.Encode(otp.ProvisionURIWithIssuer(cred.Value, g.conf.Iss), qrcode.Low, 256)
		if err != nil {
			return core.SendError(c, http.StatusInternalServerError, err.Error())
		}
		response["qr"] = qr

		manager, ok := g.pluginAPI.GetIDManager()
		if !ok {
			return errors.New("cannot get IDManager")
		}

		err = manager.On2FA(cred, &core.MFAData{
			PluginID:     fmt.Sprintf("%d", meta.PluginID),
			ProviderName: meta.ShortName,
			Payload:      mfaData,
		})
		if err != nil {
			return core.SendError(c, http.StatusInternalServerError, err.Error())
		}
		c.Set(fiber.HeaderContentType, "image/png")
		return c.Send(qr)
	}
}

func getScratchCodes(g *otpAuth) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		// check if user already authenticated
		cred := &core.Credential{Name: "email", Value: "www@example.com"}

		scratchCodes, err := generateScratchCodes(g.pluginAPI, g.conf.ScratchCode.Num, g.conf.ScratchCode.Alphabet)
		if err != nil {
			return core.SendError(c, http.StatusInternalServerError, err.Error())
		}

		manager, ok := g.pluginAPI.GetIDManager()
		if !ok {
			return errors.New("cannot get IDManager")
		}

		err = manager.On2FA(cred, &core.MFAData{
			PluginID:     fmt.Sprintf("%d", meta.PluginID),
			ProviderName: meta.ShortName,
			Payload:      map[string]interface{}{"scratch_codes": scratchCodes},
		})
		if err != nil {
			return core.SendError(c, http.StatusInternalServerError, err.Error())
		}
		return c.JSON(&fiber.Map{"scratch_codes": scratchCodes})
	}
}

func generateSecret(api core.PluginAPI) (string, error) {
	randStr, err := api.GetRandStr(8, "alphanum")
	if err != nil {
		return "", err
	}
	return base32.StdEncoding.EncodeToString([]byte(randStr)), nil
}

func generateScratchCodes(api core.PluginAPI, num int, alphabet string) (scratchCodes []string, err error) {
	scratchCodes = make([]string, num)
	for i := 0; i < num; i++ {
		scratchCodes[i], err = api.GetRandStr(8, alphabet)
		if err != nil {
			return nil, err
		}
	}
	return scratchCodes, err
}
