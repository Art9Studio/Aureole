package authenticator

import (
	"aureole/internal/core"
	"aureole/pkg/dgoogauth"
	"encoding/base32"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v4"
	"github.com/skip2/go-qrcode"
	"net/http"
	"strings"
)

func getQR(g *otpAuth) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		in := &getQRReqBody{}
		var credName, credValue string
		if err := c.BodyParser(in); err != nil {
			return err
		}
		if in.Email == "" && in.Phone == "" {
			return errors.New("email or phone is required")
		}
		if in.Email != "" {
			credName = "email"
			credValue = in.Email
		} else {
			credName = "phone"
			credValue = in.Phone
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
		if g.conf.ScratchCode.Num != 0 {
			scratchCodes, err := generateScratchCodes(g.pluginAPI, g.conf.ScratchCode.Num, g.conf.ScratchCode.Alphabet)
			if err != nil {
				return core.SendError(c, http.StatusInternalServerError, err.Error())
			}
			mfaData["scratch_codes"] = scratchCodes
			response["scratch_code"] = scratchCodes
		}

		cred := &core.Credential{Name: credName, Value: credValue}
		qr, err := qrcode.Encode(otp.ProvisionURI(cred.Value), qrcode.Low, 256)
		if err != nil {
			return core.SendError(c, http.StatusInternalServerError, err.Error())
		}
		response["qr"] = qr

		manager, ok := g.pluginAPI.GetIDManager()
		if !ok {
			return errors.New("cannot get IDManager")
		}

		err = manager.OnMFA(cred, &core.MFAData{
			PluginID:     fmt.Sprintf("%d", meta.PluginID),
			ProviderName: meta.ShortName,
			Payload:      mfaData,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return core.SendError(c, http.StatusBadRequest, err.Error())
			}
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

		err = manager.OnMFA(cred, &core.MFAData{
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

func authMiddleware(g *otpAuth, next fiber.Handler) func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		bearer := ctx.Get(fiber.HeaderAuthorization)
		tokenSplit := strings.Split(bearer, "Bearer ")

		var rawToken string
		if len(tokenSplit) == 2 && tokenSplit[1] != "" {
			rawToken = tokenSplit[1]
		} else {
			return ctx.SendStatus(http.StatusForbidden)
		}

		token, err := g.pluginAPI.ParseJWT(rawToken)
		if err != nil {
			return core.SendError(ctx, http.StatusForbidden, err.Error())
		}

		var id string
		if err = g.pluginAPI.GetFromJWT(token, "sub", &id); err != nil {
			return ctx.SendStatus(http.StatusForbidden)
		}
		ctx.Locals(core.UserID, id)

		return next(ctx)
	}
}
