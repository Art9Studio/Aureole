package google

import (
	"aureole/internal/identity"
	authzT "aureole/internal/plugins/authz/types"
	"aureole/internal/router"
	"context"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/pkg/errors"
)

func GetAuthCode(g *google) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		// todo: save state and compare later #1
		u := g.provider.AuthCodeURL("state")
		return c.Redirect(u)
	}
}

func Login(g *google) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		// todo: save state and compare later #2
		state := c.Query("state")
		if state != "state" {
			return router.SendError(c, fiber.StatusBadRequest, "invalid state")
		}
		code := c.Query("code")
		if code == "" {
			return router.SendError(c, fiber.StatusBadRequest, "code not found")
		}

		jwtT, err := getJwt(g, code)
		if err != nil {
			return router.SendError(c, fiber.StatusInternalServerError, errors.Wrap(err, "error while exchange").Error())
		}

		email, ok := jwtT.Get("email")
		if !ok {
			return router.SendError(c, fiber.StatusInternalServerError, "can't get 'email' from token")
		}
		socialId, ok := jwtT.Get("sub")
		if !ok {
			return router.SendError(c, fiber.StatusInternalServerError, "can't get 'social_id' from token")
		}
		userData, err := jwtT.AsMap(context.Background())
		if err != nil {
			return router.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		if ok, err := g.app.Filter(convertUserData(userData), g.rawConf.Filter); err != nil {
			return router.SendError(c, fiber.StatusBadRequest, err.Error())
		} else if !ok {
			return router.SendError(c, fiber.StatusBadRequest, "apple: input data doesn't pass filters")
		}

		var i map[string]interface{}
		if g.manager != nil {
			i, err = g.manager.OnUserAuthenticated(
				&identity.Credential{
					Name:  identity.Email,
					Value: email.(string),
				},
				&identity.Identity{
					Email: email.(string),
				},
				AdapterName,
				map[string]interface{}{
					"social_id": socialId,
					"user_data": userData,
				})
			if err != nil {
				return router.SendError(c, fiber.StatusInternalServerError, err.Error())
			}
		} else {
			i = map[string]interface{}{
				identity.Email: email,
				"provider":     AdapterName,
				"social_id":    socialId,
				"user_data":    userData,
			}
		}

		return g.authorizer.Authorize(c, authzT.NewPayload(g.authorizer, nil, i))
	}
}

func getJwt(g *google, code string) (jwt.Token, error) {
	t, err := g.provider.Exchange(context.Background(), code)
	if err != nil {
		return nil, err
	}
	idToken := t.Extra("id_token").(string)
	return jwt.ParseString(idToken)
}

func convertUserData(mapIntr map[string]interface{}) map[string]string {
	mapStr := make(map[string]string, len(mapIntr))
	for key, value := range mapIntr {
		mapStr[key] = fmt.Sprintf("%v", value)
	}
	return mapStr
}
