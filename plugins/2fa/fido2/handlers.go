package fido2

import (
	"aureole/internal/core"
	"aureole/internal/plugins"
	"bytes"
	"encoding/binary"
	"github.com/duo-labs/webauthn/protocol"
	"github.com/duo-labs/webauthn/webauthn"
	"github.com/gofiber/fiber/v2"
	"strconv"
)

func startRegistration(m *mfa) func(ctx *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var registerInput credentialInput
		err := c.BodyParser(&registerInput)
		if err != nil {
			return core.SendError(c, fiber.StatusBadRequest, err.Error())
		}

		cred, err := getCredential(&registerInput)
		if err != nil {
			return core.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		user, err := getUser(cred, m)
		if err != nil {
			return core.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		credentialOptions, sessionData, err := m.webauthn.BeginRegistration(user,
			webauthn.WithAuthenticatorSelection(
				protocol.AuthenticatorSelection{
					AuthenticatorAttachment: protocol.AuthenticatorAttachment(m.conf.AuthenticatorType),
					UserVerification:        "discouraged",
				}),
			webauthn.WithConveyancePreference(protocol.ConveyancePreference(m.conf.AttestationType)),
		)
		if err != nil {
			return core.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		err = m.pluginAPI.SaveToService("$registration", sessionData, m.pluginAPI.GetAuthSessionExp())
		if err != nil {
			return core.SendError(c, fiber.StatusInternalServerError, err.Error())
		}
		return c.JSON(credentialOptions)
	}
}

func finishRegistration(m *mfa) func(ctx *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var sessionData webauthn.SessionData
		ok, err := m.pluginAPI.GetFromService("$registration", &sessionData)
		if err != nil {
			return core.SendError(c, fiber.StatusInternalServerError, err.Error())
		}
		if !ok {
			return core.SendError(c, fiber.StatusInternalServerError, "session data has expired")
		}

		id, _ := binary.Uvarint(sessionData.UserID)
		cred := &plugins.Credential{
			Name:  plugins.ID,
			Value: strconv.FormatUint(id, 10),
		}
		user, err := getUser(cred, m)
		if err != nil {
			return core.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		parsedResponse, err := protocol.ParseCredentialCreationResponseBody(bytes.NewReader(c.Body()))
		credential, err := m.webauthn.CreateCredential(user, sessionData, parsedResponse)
		if err != nil {
			return core.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		err = m.idManager.On2FA(cred, &plugins.MFAData{
			PluginID:     pluginID,
			ProviderName: adapterName,
			Payload:      map[string]interface{}{"default": credential},
		})
		if err != nil {
			return core.SendError(c, fiber.StatusInternalServerError, err.Error())
		}
		return c.SendStatus(fiber.StatusCreated)
	}
}

func getUser(cred *plugins.Credential, m *mfa) (u *user, err error) {
	ident := plugins.Identity{}
	ident.ID, err = m.idManager.GetData(cred, "", "id")
	if err != nil {
		return nil, err
	}
	rawUsername, err := m.idManager.GetData(cred, "", "username")
	if err != nil {
		return nil, err
	}
	username := rawUsername.(string)
	ident.Username = &username

	mfaData, err := m.idManager.Get2FAData(cred, pluginID)
	if err != nil {
		return nil, err
	}

	u = &user{
		identity: &ident,
		mfaData:  mfaData,
	}
	return u, nil
}
