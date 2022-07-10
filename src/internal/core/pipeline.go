package core

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func pipelineAuthWrapper(authFunc AuthHandlerFunc, app *app) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		authnResult, err := authFunc(*c)
		if err != nil {
			if authnResult != nil && len(authnResult.ErrorData) != 0 {
				return sendErrorWithBody(c, fiber.StatusUnauthorized, err.Error(), authnResult.ErrorData)
			}
			return SendError(c, fiber.StatusUnauthorized, err.Error())
		}

		enabled2FA, err := getEnabled2FA(app, authnResult)
		if err != nil {
			return SendError(c, fiber.StatusUnauthorized, err.Error())
		}

		if len(enabled2FA) != 0 {
			serviceStorage, ok := app.getServiceStorage()
			if !ok {
				return SendError(c, fiber.StatusUnauthorized, "cannot get internal storage")
			}
			err := serviceStorage.Set(app.name+"$auth_pipeline$"+authnResult.Cred.Value, authnResult, app.authSessionExp)
			if err != nil {
				return SendError(c, fiber.StatusUnauthorized, err.Error())
			}

			token, err := createJWT(app, map[string]interface{}{
				"credential": authnResult.Cred,
				"provider":   authnResult.Provider,
			}, app.authSessionExp)
			if err != nil {
				return SendError(c, fiber.StatusUnauthorized, err.Error())
			}
			return c.Status(fiber.StatusAccepted).JSON(fiber.Map{"token": token, "2fa": enabled2FA})
		}

		identity, err := authenticate(app, authnResult)
		if err != nil {
			return SendError(c, fiber.StatusUnauthorized, err.Error())
		}
		return authorize(c, app, identity)
	}
}

func mfaInitHandler(init2FA MFAInitFunc, _ *app) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		mfaData, err := init2FA(*c)
		if err != nil {
			return SendError(c, fiber.StatusUnauthorized, err.Error())
		}
		return c.JSON(mfaData)
	}
}

func mfaVerificationHandler(verify2FA MFAVerifyFunc, app *app) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		cred, mfaData, err := verify2FA(*c)
		if err != nil {
			if mfaData != nil {
				return sendErrorWithBody(c, fiber.StatusUnauthorized, err.Error(), mfaData)
			}
			return SendError(c, fiber.StatusUnauthorized, err.Error())
		}

		serviceStorage, ok := app.getServiceStorage()
		if !ok {
			return SendError(c, fiber.StatusUnauthorized, "cannot get internal storage")
		}

		authnResult := &AuthResult{}
		ok, err = serviceStorage.Get(app.name+"$auth_pipeline$"+cred.Value, authnResult)
		if err != nil {
			return SendError(c, fiber.StatusUnauthorized, err.Error())
		}
		if !ok {
			return SendError(c, fiber.StatusUnauthorized, "auth session has expired, cannot get user data")
		}

		err = serviceStorage.Delete(app.name + "$auth_pipeline$" + cred.Value)
		if err != nil {
			return SendError(c, fiber.StatusUnauthorized, err.Error())
		}

		identity, err := authenticate(app, authnResult)
		if err != nil {
			return SendError(c, fiber.StatusUnauthorized, err.Error())
		}
		return authorize(c, app, identity)
	}
}

func getEnabled2FA(app *app, authnResult *AuthResult) (fiber.Map, error) {
	secondFactors, ok := app.getSecondFactors()
	if ok {
		var enabled2FA []MFA
		for _, secondFactor := range secondFactors {
			enabled, err := secondFactor.IsEnabled(authnResult.Cred)
			if err != nil {
				return nil, err
			}
			if enabled {
				enabled2FA = append(enabled2FA, secondFactor)
			}
		}

		if len(enabled2FA) != 0 {
			enabledFactorsMap := fiber.Map{}
			for _, enabledFactor := range enabled2FA {
				path := app.url.String() + "/2fa/" +
					strings.ReplaceAll(enabledFactor.GetMetadata().ShortName, "_", "-")
				enabledFactorsMap[enabledFactor.GetMetadata().ShortName] = path
			}
			return enabledFactorsMap, nil
		}
	}
	return nil, nil
}

func authenticate(app *app, authnResult *AuthResult) (*Identity, error) {
	manager, ok := app.getIDManager()
	if ok {
		return manager.OnUserAuthenticated(authnResult.Cred, authnResult.Identity, authnResult.Provider)
	}
	return authnResult.Identity, nil
}

func authorize(c *fiber.Ctx, app *app, identity *Identity) error {
	authz, ok := app.getIssuer()
	if !ok {
		return SendError(c, fiber.StatusUnauthorized, fmt.Sprintf("app %s: cannot get issuer", app.name))
	}

	payload, err := NewIssuerPayload(identity.AsMap())
	if err != nil {
		return SendError(c, fiber.StatusUnauthorized, err.Error())
	}
	return authz.Authorize(c, payload)
}

func sendErrorWithBody(c *fiber.Ctx, statusCode int, message string, body fiber.Map) error {
	response := fiber.Map{"error": message}
	for k, v := range body {
		response[k] = v
	}
	return c.Status(statusCode).JSON(response)
}
