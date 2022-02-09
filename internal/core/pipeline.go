package core

import (
	"aureole/internal/plugins"
	"fmt"
	"github.com/gofiber/fiber/v2"
)

func loginHandler(authFunc plugins.AuthNLoginFunc, app *app) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		authnResult, err := authFunc(*c)
		if err != nil {
			if authnResult != nil && len(authnResult.ErrorData) != 0 {
				return sendErrorWithBody(c, fiber.StatusUnauthorized, err.Error(), authnResult.ErrorData)
			}
			return SendError(c, fiber.StatusUnauthorized, err.Error())
		}

		enabled2FA, err := getUserEnabled2FA(app, authnResult)
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
			return c.Status(fiber.StatusAccepted).JSON(fiber.Map{"token": token, "2fa": enabled2FA})
		}

		identity, err := authenticate(app, authnResult)
		if err != nil {
			return SendError(c, fiber.StatusUnauthorized, err.Error())
		}
		return authorize(c, app, identity)
	}
}

func mfaInitHandler(init2FA plugins.MFAInitFunc, _ *app) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		mfaData, err := init2FA(*c)
		if err != nil {
			return SendError(c, fiber.StatusUnauthorized, err.Error())
		}
		return c.JSON(mfaData)
	}
}

func mfaVerificationHandler(verify2FA plugins.MFAVerifyFunc, app *app) func(*fiber.Ctx) error {
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

func getUserEnabled2FA(app *app, authnResult *plugins.AuthNResult) ([]string, error) {
	secondFactors, ok := app.getSecondFactors()
	if ok {
		var enabled2FA []string
		for _, secondFactor := range secondFactors {
			enabled, err := secondFactor.IsEnabled(authnResult.Cred)
			if err != nil {
				return nil, err
			}
			if enabled {
				enabled2FA = append(enabled2FA, secondFactor.GetMetaData().Type)
			}

		}
		return enabled2FA, nil
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
	response := fiber.Map{"error": message}
	for k, v := range body {
		response[k] = v
	}
	return c.Status(statusCode).JSON(response)
}
