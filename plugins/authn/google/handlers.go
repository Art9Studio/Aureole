package google

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/lestrrat-go/jwx/jwt"
)

func getAuthCode(g *google) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		// todo: save state and compare later #1
		u := g.provider.AuthCodeURL("state")
		return c.Redirect(u)
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
