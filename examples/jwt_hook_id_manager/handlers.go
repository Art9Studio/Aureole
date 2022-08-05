package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/lestrrat-go/jwx/jwt"
)

type Input struct {
	Token string
}

func handleRequest(m *IDManager) func(ctx *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		i := Input{}
		err := c.BodyParser(&i)
		if err != nil {
			return sendError(c, http.StatusBadRequest, fmt.Sprintf("body parse failed: %s", err.Error()))
		}
		if i.Token == "" {
			return sendError(c, http.StatusBadRequest, "token is required")
		}

		token, err := jwt.ParseString(
			i.Token,
			jwt.WithIssuer("Aureole"),
			jwt.WithValidate(true),
			jwt.WithKeySet(m.publicSet),
		)
		if err != nil {
			return sendError(c, http.StatusBadRequest, "cannot parse given token: "+err.Error())
		}

		event, ok := token.Get("event")
		if !ok {
			return sendError(c, http.StatusBadRequest, "cannot get 'event' field from token: "+err.Error())
		}

		switch event {
		case "Register":
			return m.register(c, token)
		case "OnUserAuthenticated":
			return m.onUserAuthenticated(c, token)
		case "OnMFA":
			return m.onMFA(c, token)
		case "GetData":
			return m.getData(c, token)
		case "GetMFAData":
			return m.getMFAData(c, token)
		case "Update":
			return m.update(c, token)
		case "CheckFeaturesAvailable":
			return m.checkFeaturesAvailable(c, token)
		default:
			return sendError(c, http.StatusBadRequest, fmt.Sprintf("event '%s' is not supported", event))
		}
	}
}

func (m *IDManager) register(c *fiber.Ctx, token jwt.Token) error {
	var (
		cred          *Credential
		ident         *Identity
		authnProvider string
	)

	err := getFromJWT(token, "credential", cred)
	if err != nil {
		return sendError(c, http.StatusBadRequest, "cannot get credential from token: "+err.Error())
	}
	err = getFromJWT(token, "identity", ident)
	if err != nil {
		return sendError(c, http.StatusBadRequest, "cannot get ident from token: "+err.Error())
	}
	err = getFromJWT(token, "authn_provider", &authnProvider)
	if err != nil {
		return sendError(c, http.StatusBadRequest, "cannot get authn_provider from token: "+err.Error())
	}

	fmt.Printf("'Register' event request with parameters: \n"+
		"Credential: %v\n"+
		"Identity: %v\n"+
		"AuthN provider: %s\n", cred, ident, authnProvider)

	conn, err := m.pool.Acquire(context.Background())
	if err != nil {
		return sendError(c, http.StatusInternalServerError, "cannot acquire connection: "+err.Error())
	}
	defer conn.Release()

	return c.JSON(fiber.Map{"token": token})
}

func (m *IDManager) onUserAuthenticated(c *fiber.Ctx, token jwt.Token) error {
	var (
		cred          *Credential
		ident         *Identity
		authnProvider string
	)

	err := getFromJWT(token, "credential", &cred)
	if err != nil {
		return sendError(c, http.StatusBadRequest, "cannot get credential from token: "+err.Error())
	}
	err = getFromJWT(token, "identity", &ident)
	if err != nil {
		return sendError(c, http.StatusBadRequest, "cannot get identity from token: "+err.Error())
	}
	err = getFromJWT(token, "authn_provider", &authnProvider)
	if err != nil {
		return sendError(c, http.StatusBadRequest, "cannot get authn_provider from token: "+err.Error())
	}

	fmt.Printf("'OnUserAuthenticated' event request with parameters: \n"+
		"Credential: %v\n"+
		"Identity: %v\n"+
		"AuthN provider: %s\n", cred, ident, authnProvider)

	conn, err := m.pool.Acquire(context.Background())
	if err != nil {
		return sendError(c, http.StatusInternalServerError, "cannot acquire connection: "+err.Error())
	}
	defer conn.Release()

	return c.JSON(fiber.Map{"token": token})
}

func (m *IDManager) onMFA(c *fiber.Ctx, token jwt.Token) error {
	var (
		cred        *Credential
		mfaProvider string
		mfaData     fiber.Map
	)

	err := getFromJWT(token, "credential", &cred)
	if err != nil {
		return sendError(c, http.StatusBadRequest, "cannot get credential from token: "+err.Error())
	}
	err = getFromJWT(token, "2fa_provider", &mfaProvider)
	if err != nil {
		return sendError(c, http.StatusBadRequest, "cannot get 2fa_provider from token: "+err.Error())
	}
	err = getFromJWT(token, "2fa_data", &mfaData)
	if err != nil {
		return sendError(c, http.StatusBadRequest, "cannot get 2fa_data from token: "+err.Error())
	}

	fmt.Printf("'OnMFA' event request with parameters: \n"+
		"Credential: %v\n"+
		"2FA provider: %v\n"+
		"2FA data: %v\n", cred, mfaProvider, mfaData)

	conn, err := m.pool.Acquire(context.Background())
	if err != nil {
		return sendError(c, http.StatusInternalServerError, "cannot acquire connection: "+err.Error())
	}
	defer conn.Release()

	return c.SendStatus(http.StatusOK)
}

func (m *IDManager) getData(c *fiber.Ctx, token jwt.Token) error {
	var (
		cred                *Credential
		name, authnProvider string
	)

	err := getFromJWT(token, "credential", cred)
	if err != nil {
		return sendError(c, http.StatusBadRequest, "cannot get credential from token: "+err.Error())
	}
	err = getFromJWT(token, "name", &name)
	if err != nil {
		return sendError(c, http.StatusBadRequest, "cannot get name from token: "+err.Error())
	}
	err = getFromJWT(token, "authn_provider", &authnProvider)
	if err != nil {
		return sendError(c, http.StatusBadRequest, "cannot get authn_provider from token: "+err.Error())
	}

	fmt.Printf("'GetData' event request with parameters: \n"+
		"Credential: %v\n"+
		"ProviderName: %s\n"+
		"AuthN provider: %s\n", cred, name, authnProvider)

	conn, err := m.pool.Acquire(context.Background())
	if err != nil {
		return sendError(c, http.StatusInternalServerError, "cannot acquire connection: "+err.Error())
	}
	defer conn.Release()

	return c.JSON(fiber.Map{"token": token})
}

func (m *IDManager) getMFAData(c *fiber.Ctx, token jwt.Token) error {
	var (
		cred        *Credential
		mfaProvider string
	)

	err := getFromJWT(token, "credential", cred)
	if err != nil {
		return sendError(c, http.StatusBadRequest, "cannot get credential from token: "+err.Error())
	}
	err = getFromJWT(token, "2fa_provider", &mfaProvider)
	if err != nil {
		return sendError(c, http.StatusBadRequest, "cannot get 2fa_provider from token: "+err.Error())
	}

	fmt.Printf("'GetMFAData' event request with parameters: \n"+
		"Credential: %v\n"+
		"2FA provider: %s\n", cred, mfaProvider)

	conn, err := m.pool.Acquire(context.Background())
	if err != nil {
		return sendError(c, http.StatusInternalServerError, "cannot acquire connection: "+err.Error())
	}
	defer conn.Release()

	return c.JSON(fiber.Map{"token": token})
}

func (m *IDManager) update(c *fiber.Ctx, token jwt.Token) error {
	var (
		cred          *Credential
		ident         *Identity
		authnProvider string
	)

	err := getFromJWT(token, "credential", cred)
	if err != nil {
		return sendError(c, http.StatusBadRequest, "cannot get credential from token: "+err.Error())
	}
	err = getFromJWT(token, "identity", ident)
	if err != nil {
		return sendError(c, http.StatusBadRequest, "cannot get identity from token: "+err.Error())
	}
	err = getFromJWT(token, "authn_provider", &authnProvider)
	if err != nil {
		return sendError(c, http.StatusBadRequest, "cannot get authn_provider from token: "+err.Error())
	}

	fmt.Printf("'Update' event request with parameters: \n"+
		"Credential: %v\n"+
		"Identity: %v\n"+
		"AuthN provider: %s\n", cred, ident, authnProvider)

	conn, err := m.pool.Acquire(context.Background())
	if err != nil {
		return sendError(c, http.StatusInternalServerError, "cannot acquire connection: "+err.Error())
	}
	defer conn.Release()

	return c.JSON(fiber.Map{"token": token})
}

func (m *IDManager) checkFeaturesAvailable(c *fiber.Ctx, token jwt.Token) error {
	var requiredFeatures []string
	err := getFromJWT(token, "features", &requiredFeatures)
	if err != nil {
		return sendError(c, http.StatusBadRequest, "cannot get features from token: "+err.Error())
	}

	fmt.Printf("'CheckFeaturesAvailable' event request with parameters: \n"+
		"Required features: %v\n", requiredFeatures)

	for _, f := range requiredFeatures {
		if available, ok := m.features[f]; !ok || !available {
			return sendError(c, http.StatusNotFound, fmt.Sprintf("feature %s hasn't implemented", f))
		}
	}
	return c.SendStatus(http.StatusOK)
}

func sendError(c *fiber.Ctx, statusCode int, message string) error {
	return c.Status(statusCode).JSON(fiber.Map{
		"success": false,
		"error":   message,
	})
}
