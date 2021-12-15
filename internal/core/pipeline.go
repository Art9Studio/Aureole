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

		secondFactors, ok := app.getSecondFactors()
		if !ok {
			return authorize(c, app, authnResult)
		}

		var enabledFactors []plugins.SecondFactor
		for _, secondFactor := range secondFactors {
			enabled, err := secondFactor.IsEnabled(authnResult.Cred)
			if err != nil {
				return SendError(c, fiber.StatusUnauthorized, err.Error())
			}
			if enabled {
				enabledFactors = append(enabledFactors, secondFactor)
			}
		}

		if len(enabledFactors) != 0 {
			serviceStorage, ok := app.getServiceStorage()
			if !ok {
				return SendError(c, fiber.StatusUnauthorized, "cannot get service storage")
			}
			err = serviceStorage.Set(app.getName()+"$auth_pipeline$"+authnResult.Cred.Value, authnResult, app.getAuthSessionExp())
			if err != nil {
				return SendError(c, fiber.StatusUnauthorized, err.Error())
			}

			var enabledFactorsJson fiber.Map
			for _, enabledFactor := range enabledFactors {
				path := app.url.String() + "/2fa/" +
					strings.ReplaceAll(enabledFactor.GetMetaData().Type, "_", "-")
				enabledFactorsJson[enabledFactor.GetMetaData().Type] = path
			}

			token, err := createJWT(app, map[string]interface{}{
				"credential": authnResult.Cred,
				"provider":   authnResult.Provider,
			}, app.getAuthSessionExp())
			if err != nil {
				return SendError(c, fiber.StatusUnauthorized, err.Error())
			}
			return c.JSON(fiber.Map{"token": token, "2fa": enabledFactorsJson})
		}
		return authorize(c, app, authnResult)
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
		ok, err = serviceStorage.Get(app.getName()+"$auth_pipeline$"+cred.Value, authnResult)
		if err != nil {
			return SendError(c, fiber.StatusUnauthorized, err.Error())
		}
		if !ok {
			return SendError(c, fiber.StatusUnauthorized, "auth session has expired, cannot get user data")
		}
		if err := serviceStorage.Delete(app.getName() + "$auth_pipeline$" + cred.Value); err != nil {
			return SendError(c, fiber.StatusUnauthorized, err.Error())
		}

		return authorize(c, app, authnResult)
	}
}

func authorize(c *fiber.Ctx, app *app, authnResult *plugins.AuthNResult) error {
	authz, ok := app.getAuthorizer()
	if !ok {
		return SendError(c, fiber.StatusUnauthorized, fmt.Sprintf("app %s: cannot get authorizer", app.name))
	}

	manager, ok := app.getIDManager()
	if ok {
		user, err := manager.OnUserAuthenticated(authnResult.Cred, authnResult.Identity, authnResult.Provider)
		if err != nil {
			return SendError(c, fiber.StatusUnauthorized, err.Error())
		}

		payload, err := plugins.NewPayload(user.AsMap())
		if err != nil {
			return SendError(c, fiber.StatusUnauthorized, err.Error())
		}
		return authz.Authorize(c, payload)
	}

	payload, err := plugins.NewPayload(authnResult.Identity.AsMap())
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
