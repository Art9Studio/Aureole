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
		in := &getQRReqBody{}
		var credName, credValue string
		if err := c.BodyParser(in); err != nil {
			return err
		}
		if in.Email == "" && in.Phone == "" {
			return errors.New("email or phone is required")
		}
		if in.Email != "" {
			credName = core.Email
			credValue = in.Email
		} else {
			credName = core.Phone
			credValue = in.Phone
		}

		mfaData := core.Secrets{}
		response := make(map[string]interface{})
		userId := g.pluginAPI.GetUserID(c)
		if userId == "" {
			return core.SendError(c, http.StatusForbidden, "auth id not found")
		}

		secret, err := generateSecret(g.pluginAPI)
		if err != nil {
			return core.SendError(c, http.StatusInternalServerError, err.Error())
		}
		mfaData["secret"] = &secret

		otp := &dgoogauth.OTPConfig{Secret: strings.TrimSpace(secret)}
		if g.conf.Alg == hotp {
			otp.HotpCounter = 1
			cnt := fmt.Sprintf("%d", 1)
			mfaData[counter] = &cnt
		}

		cred := &core.Credential{Name: credName, Value: credValue}
		qr, err := qrcode.Encode(otp.ProvisionURI(cred.Value), qrcode.Low, 256)
		if err != nil {
			return core.SendError(c, http.StatusInternalServerError, err.Error())
		}
		response[qrCode] = qr

		manager, ok := g.pluginAPI.GetIDManager()
		if !ok {
			return errors.New("cannot get IDManager")
		}

		if _, err = manager.RegisterOrUpdate(
			&core.AuthResult{
				Cred: &core.Credential{
					Name:  core.ID,
					Value: userId,
				},
				User: &core.User{
					ID:          userId,
					EnabledMFAs: []string{fmt.Sprintf("%d", meta.PluginID)},
				},
			},
		); err != nil {
			return core.SendError(c, http.StatusInternalServerError, err.Error())
		}

		if err = manager.SetSecrets(cred, fmt.Sprintf("%d", meta.PluginID), &mfaData); err != nil {
			return core.SendError(c, http.StatusInternalServerError, err.Error())
		}
		c.Set(fiber.HeaderContentType, core.MIMEImagePNG)
		return c.Send(qr)
	}
}

func generateSecret(api core.PluginAPI) (string, error) {
	randStr, err := api.GetRandStr(8, "alphanum")
	if err != nil {
		return "", err
	}
	return base32.StdEncoding.EncodeToString([]byte(randStr)), nil
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
		if err = g.pluginAPI.GetFromJWT(token, core.Sub, &id); err != nil {
			return core.SendError(ctx, http.StatusForbidden, err.Error())
		}
		ctx.Locals(core.UserID, id)

		return next(ctx)
	}
}
