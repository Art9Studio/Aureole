package core

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type AuthUnauthorizedResult struct {
	Error string                 `json:"error"`
	Data  map[string]interface{} `json:"data"`
}

func ErrorBody(err error, body map[string]interface{}) AuthUnauthorizedResult {
	return AuthUnauthorizedResult{Error: err.Error(), Data: body}
}

func pipelineAuthWrapper(authFunc AuthHandlerFunc, app *app) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		authnResult, err := authFunc(*c)
		if err != nil {
			if authnResult != nil && len(authnResult.ErrorData) != 0 {
				return c.Status(http.StatusUnauthorized).JSON(ErrorBody(err, authnResult.ErrorData))
			}
			return c.Status(http.StatusUnauthorized).JSON(ErrorBody(err, nil))
		}

		enabled2FA, err := getEnabledMFA(app, authnResult)
		if err != nil {
			return c.Status(http.StatusUnauthorized).JSON(ErrorBody(err, nil))
		}

		if len(enabled2FA) != 0 {
			serviceStorage, ok := app.getServiceStorage()
			if !ok {
				return c.Status(http.StatusUnauthorized).JSON(ErrorBody(errors.New("cannot get internal storage"), nil))
			}
			err := serviceStorage.Set(app.name+"$auth_pipeline$"+authnResult.Cred.Value, authnResult, app.authSessionExp)
			if err != nil {
				return c.Status(http.StatusUnauthorized).JSON(ErrorBody(err, nil))
			}

			token, err := createJWT(app, map[string]interface{}{
				"credential": authnResult.Cred,
				"provider":   authnResult.Provider,
			}, app.authSessionExp)
			if err != nil {
				return c.Status(http.StatusUnauthorized).JSON(ErrorBody(err, nil))
			}
			// todo: document this
			return c.Status(http.StatusAccepted).JSON(fiber.Map{"token": token, "2fa": enabled2FA})
		}

		// todo: I don't like this name
		authRes, err := authenticate(app, authnResult)
		if err != nil {
			return c.Status(http.StatusUnauthorized).JSON(ErrorBody(err, nil))
		}
		return authorize(c, app, authRes.User)
	}
}

func mfaInitHandler(init2FA MFAInitFunc, _ *app) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		mfaData, err := init2FA(*c)
		if err != nil {
			return c.Status(http.StatusUnauthorized).JSON(ErrorBody(err, nil))
		}
		return c.JSON(mfaData)
	}
}

func mfaVerificationHandler(verify2FA MFAVerifyFunc, app *app) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		cred, mfaData, err := verify2FA(*c)
		if err != nil {
			if mfaData != nil {
				return c.Status(http.StatusUnauthorized).JSON(ErrorBody(err, mfaData))
			}
			return c.Status(http.StatusUnauthorized).JSON(ErrorBody(err, nil))
		}

		serviceStorage, ok := app.getServiceStorage()
		if !ok {
			return c.Status(http.StatusUnauthorized).JSON(ErrorBody(errors.New("cannot get internal storage"), nil))
		}

		authnResult := &AuthResult{}
		ok, err = serviceStorage.Get(app.name+"$auth_pipeline$"+cred.Value, authnResult)
		if err != nil {
			return c.Status(http.StatusUnauthorized).JSON(ErrorBody(err, nil))
		}
		if !ok {
			return c.Status(http.StatusUnauthorized).JSON(ErrorBody(errors.New("auth session has expired, cannot get user Data"), nil))
		}

		err = serviceStorage.Delete(app.name + "$auth_pipeline$" + cred.Value)
		if err != nil {
			return c.Status(http.StatusUnauthorized).JSON(ErrorBody(err, nil))
		}
		id := c.Locals(UserID)
		if id != "" {
			authnResult.Identity.ID = id
		}
		authRes, err := authenticate(app, authnResult)
		if err != nil {
			return c.Status(http.StatusUnauthorized).JSON(ErrorBody(err, nil))
		}
		//todo(Talgat) add User instead of nil
		return authorize(c, app, authRes.User)
	}
}

func getEnabledMFA(app *app, authnResult *AuthResult) (fiber.Map, error) {
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

func authenticate(app *app, authnResult *AuthResult) (*AuthResult, error) {
	manager, ok := app.getIDManager()
	if ok {
		return manager.Register(authnResult)
	}
	return authnResult, nil
}

func authorize(c *fiber.Ctx, app *app, user *User) error {
	authz, ok := app.getIssuer()
	if !ok {
		return c.Status(http.StatusUnauthorized).JSON(ErrorBody(errors.New(fmt.Sprintf("app %s: cannot get issuer", app.name)), nil))
	}

	payload, err := NewIssuerPayload(user.AsMap())
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(ErrorBody(err, nil))
	}
	return authz.Authorize(c, payload)
}
