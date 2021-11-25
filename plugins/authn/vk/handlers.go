package vk

import (
	"aureole/internal/identity"
	authzT "aureole/internal/plugins/authz/types"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"net/url"
	"strings"
)

func GetAuthCode(v *vk) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		u := v.provider.AuthCodeURL("state")
		return c.Redirect(u)
	}
}

func Login(v *vk) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		state := c.Query("state")
		if state != "state" {
			return sendError(c, fiber.StatusBadRequest, "invalid state")
		}
		code := c.Query("code")
		if code == "" {
			return sendError(c, fiber.StatusBadRequest, "code not found")
		}

		userData, err := getUserData(v, code)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		if ok, err := v.app.Filter(convertUserData(userData), v.rawConf.Filter); err != nil {
			return sendError(c, fiber.StatusBadRequest, err.Error())
		} else if !ok {
			return sendError(c, fiber.StatusBadRequest, "apple: input data doesn't pass filters")
		}

		var i map[string]interface{}
		if v.manager != nil {
			i, err = v.manager.OnUserAuthenticated(
				&identity.Credential{
					Name:  identity.Email,
					Value: userData["email"].(string),
				},
				&identity.Identity{
					Email: userData["email"].(string),
				},
				AdapterName,
				map[string]interface{}{
					"social_id": userData["user_id"],
					"user_data": userData,
				})
			if err != nil {
				return sendError(c, fiber.StatusInternalServerError, err.Error())
			}
		} else {
			i = map[string]interface{}{
				identity.Email: userData["email"],
				"provider":     AdapterName,
				"social_id":    userData["user_id"],
				"user_data":    userData,
			}
		}

		return v.authorizer.Authorize(c, authzT.NewPayload(v.authorizer, nil, i))
	}
}

func getUserData(v *vk, code string) (map[string]interface{}, error) {
	ctx := context.Background()
	t, err := v.provider.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}

	u, err := getUserInfoUrl(v)
	if err != nil {
		return nil, err
	}

	client := v.provider.Client(ctx, t)
	resp, err := client.Get(u)
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	userArr := data["response"].([]interface{})
	userData := userArr[0].(map[string]interface{})
	userData["email"] = t.Extra("email")
	userData["user_id"] = t.Extra("user_id")
	return userData, nil
}

func getUserInfoUrl(v *vk) (string, error) {
	u, err := url.Parse("https://api.vk.com/method/users.get")
	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Set("v", fmt.Sprintf("%f", 5.131))
	fieldsStr := strings.Join(v.conf.Fields, ",")
	q.Set("fields", fieldsStr)
	u.RawQuery = q.Encode()

	return u.String(), nil
}
