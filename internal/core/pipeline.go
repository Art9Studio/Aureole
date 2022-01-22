package core

import (
	"aureole/internal/plugins"

	"github.com/gofiber/fiber/v2"
)

func handleLogin(authFunc plugins.AuthNLoginFunc, project *Project, app *App) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		authnResult, err := authFunc(*c)
		if err != nil {
			if authnResult.Additional != nil {
				return sendErrorWithBody(c, fiber.StatusUnauthorized, err.Error(), authnResult.Additional)
			}
			return SendError(c, fiber.StatusUnauthorized, err.Error())
		}

		if secondFactor, err := app.GetSecondFactor(); err == nil {
			enabled, err := secondFactor.IsEnabled(authnResult.Cred, authnResult.Provider)
			if err != nil {
				return SendError(c, fiber.StatusUnauthorized, err.Error())
			}

			if enabled {
				serviceStorage, err := project.GetServiceStorage()
				if err != nil {
					return SendError(c, fiber.StatusUnauthorized, err.Error())
				}
				err = serviceStorage.Set(app.GetName()+"$auth_pipeline$"+authnResult.Cred.Value, authnResult, app.GetAuthSessionExp())
				if err != nil {
					return SendError(c, fiber.StatusUnauthorized, err.Error())
				}

				mfaData, err := secondFactor.Init2FA(authnResult.Cred, authnResult.Provider, *c)
				if err != nil {
					return SendError(c, fiber.StatusUnauthorized, err.Error())
				}
				return c.JSON(mfaData)
			}
		}

		return authorize(c, app, authnResult)
	}
}

func handle2FA(mfaFunc plugins.MFAVerifyFunc, project *Project, app *App) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		cred, mfaData, err := mfaFunc(*c)
		if err != nil && mfaData == nil {
			return SendError(c, fiber.StatusUnauthorized, err.Error())
		} else if err != nil && mfaData != nil {
			mfaData["success"] = false
			mfaData["message"] = err.Error()
			return c.Status(fiber.StatusUnauthorized).JSON(mfaData)
		}

		serviceStorage, err := project.GetServiceStorage()
		if err != nil {
			return SendError(c, fiber.StatusUnauthorized, err.Error())
		}

		authnResult := &plugins.AuthNResult{}
		ok, err := serviceStorage.Get(app.GetName()+"$auth_pipeline$"+cred.Value, authnResult)
		if err != nil {
			return SendError(c, fiber.StatusUnauthorized, err.Error())
		}
		if !ok {
			return SendError(c, fiber.StatusUnauthorized, "auth session has expired, cannot get user data")
		}
		if err := serviceStorage.Delete(app.GetName() + "$auth_pipeline$" + cred.Value); err != nil {
			return SendError(c, fiber.StatusUnauthorized, err.Error())
		}

		return authorize(c, app, authnResult)
	}
}

func authorize(c *fiber.Ctx, app *App, authnResult *plugins.AuthNResult) error {
	authz, err := app.GetAuthorizer()
	if err != nil {
		return err
	}

	if manager, err := app.GetIDManager(); err != nil {
		user, err := manager.OnUserAuthenticated(authnResult.Cred, authnResult.Identity, authnResult.Provider)
		if err != nil {
			return err
		}

		payload, err := plugins.NewPayload(user.AsMap())
		if err != nil {
			return err
		}
		return authz.Authorize(c, payload)
	}

	payload, err := plugins.NewPayload(authnResult.Identity.AsMap())
	if err != nil {
		return err
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
