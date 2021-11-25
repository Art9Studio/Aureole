package internal

import (
	"aureole/internal/identity"
	mfaT "aureole/internal/plugins/2fa/types"
	authnT "aureole/internal/plugins/authn/types"
	authzT "aureole/internal/plugins/authz/types"
	"aureole/internal/router"
	state "aureole/internal/state/interface"
	"github.com/gofiber/fiber/v2"
)

func HandleLogin(authFunc authnT.AuthFunc, project state.ProjectState, app state.AppState) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		cred, userData, err := authFunc(*c)
		if err != nil {
			return router.SendError(c, fiber.StatusUnauthorized, err.Error())
		}
		authnProvider := userData["provider"].(string)

		if secondFactor, err := app.GetSecondFactor(); err == nil {
			enabled, err := secondFactor.IsEnabled(cred, authnProvider)
			if err != nil {
				return router.SendError(c, fiber.StatusUnauthorized, err.Error())
			}

			if enabled {
				serviceStorage, err := project.GetServiceStorage()
				if err != nil {
					return router.SendError(c, fiber.StatusUnauthorized, err.Error())
				}
				err = serviceStorage.Set(app.GetName()+"$auth_pipeline$"+cred.Value, userData, app.GetAuthSessionExp())
				if err != nil {
					return router.SendError(c, fiber.StatusUnauthorized, err.Error())
				}

				mfaData, err := secondFactor.Init2FA(cred, authnProvider, *c)
				if err != nil {
					return router.SendError(c, fiber.StatusUnauthorized, err.Error())
				}
				return c.JSON(mfaData)
			}
		}

		return authorize(c, app, cred, userData)
	}
}

func Handle2FA(mfaFunc mfaT.MFAFunc, project state.ProjectState, app state.AppState) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		cred, mfaData, err := mfaFunc(*c)
		if err != nil && mfaData == nil {
			return router.SendError(c, fiber.StatusUnauthorized, err.Error())
		} else if err != nil && mfaData != nil {
			mfaData["success"] = false
			mfaData["message"] = err.Error()
			return c.Status(fiber.StatusUnauthorized).JSON(mfaData)
		}

		serviceStorage, err := project.GetServiceStorage()
		if err != nil {
			return router.SendError(c, fiber.StatusUnauthorized, err.Error())
		}

		userData := fiber.Map{}
		ok, err := serviceStorage.Get(app.GetName()+"$auth_pipeline$"+cred.Value, &userData)
		if err != nil {
			return router.SendError(c, fiber.StatusUnauthorized, err.Error())
		}
		if !ok {
			return router.SendError(c, fiber.StatusUnauthorized, "auth session has expired, cannot get user data")
		}
		if err := serviceStorage.Delete(app.GetName() + "$auth_pipeline$" + cred.Value); err != nil {
			return router.SendError(c, fiber.StatusUnauthorized, err.Error())
		}

		return authorize(c, app, cred, userData)
	}
}

func authorize(c *fiber.Ctx, app state.AppState, cred *identity.Credential, userData fiber.Map) error {
	authz, err := app.GetAuthorizer()
	if err != nil {
		return err
	}

	if manager, err := app.GetIdentityManager(); err != nil {
		i, err := identity.NewIdentity(userData)
		if err != nil {
			return err
		}

		data, err := manager.OnUserAuthenticated(cred, i, userData["provider"].(string), userData)
		if err != nil {
			return err
		}

		payload, err := authzT.NewPayload(data)
		if err != nil {
			return err
		}
		return authz.Authorize(c, payload)
	}

	payload, err := authzT.NewPayload(userData)
	if err != nil {
		return err
	}
	return authz.Authorize(c, payload)
}
