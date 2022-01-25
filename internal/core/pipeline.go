package core

import (
	"aureole/internal/plugins"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func handleLogin(authFunc plugins.AuthNLoginFunc, app *app) func(*fiber.Ctx) error {
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
				return SendError(c, fiber.StatusUnauthorized, "cannot get service storage")
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
			return c.JSON(fiber.Map{"token": token, "2fa": enabled2FA})
		}

		identity, err := authenticate(app, authnResult)
		if err != nil {
			return SendError(c, fiber.StatusUnauthorized, err.Error())
		}
		return authorize(c, app, identity)
	}
}

func handle2FAInit(mfaFunc plugins.MFAInitFunc, _ *app) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		mfaData, err := mfaFunc(*c)
		if err != nil {
			return SendError(c, fiber.StatusUnauthorized, err.Error())
		}
		return c.JSON(mfaData)
	}
}

func handle2FAVerify(mfaFunc plugins.MFAVerifyFunc, app *app) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		cred, mfaData, err := mfaFunc(*c)
		if err != nil && mfaData == nil {
			return SendError(c, fiber.StatusUnauthorized, err.Error())
		} else if err != nil && mfaData != nil {
			mfaData["success"] = false
			mfaData["message"] = err.Error()
			return c.Status(fiber.StatusUnauthorized).JSON(mfaData)
		}

		serviceStorage, ok := app.getServiceStorage()
		if !ok {
			return SendError(c, fiber.StatusUnauthorized, "cannot get service storage")
		}

		authnResult := &plugins.AuthNResult{}
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

func getEnabled2FA(app *app, authnResult *plugins.AuthNResult) (fiber.Map, error) {
	secondFactors, ok := app.getSecondFactors()
	if ok {
		var enabled2FA []plugins.SecondFactor
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
					strings.ReplaceAll(enabledFactor.GetMetaData().Type, "_", "-")
				enabledFactorsMap[enabledFactor.GetMetaData().Type] = path
			}
			return enabledFactorsMap, nil
		}
	}
	return nil, nil
}

func authenticate(app *app, authnResult *plugins.AuthNResult) (*plugins.Identity, error) {
	manager, ok := app.getIDManager()
	if ok {
		return manager.OnUserAuthenticated(authnResult.Cred, authnResult.Identity, authnResult.Provider)
	}
	return authnResult.Identity, nil
}

func authorize(c *fiber.Ctx, app *app, identity *plugins.Identity) error {
	authz, ok := app.getAuthorizer()
	if !ok {
		return SendError(c, fiber.StatusUnauthorized, fmt.Sprintf("app %s: cannot get authorizer", app.name))
	}

	payload, err := plugins.NewPayload(identity.AsMap())
	if err != nil {
		return SendError(c, fiber.StatusUnauthorized, err.Error())
	}
	return authz.Authorize(c, payload)
}

func sendErrorWithBody(c *fiber.Ctx, statusCode int, message string, body fiber.Map) error {
	responseJSON := fiber.Map{
		"success": false,
		"message": message,
	}
	for k, v := range body {
		responseJSON[k] = v
	}
	return c.Status(statusCode).JSON(responseJSON)
}
