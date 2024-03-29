package vk

import (
	"aureole/internal/core"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func getAuthCode(v *vk) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		u := v.provider.AuthCodeURL(core.State)
		return c.Redirect(u)
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

	request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, u, http.NoBody)
	if err != nil {
		return nil, err
	}

	client := v.provider.Client(ctx, t)
	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	userArr := data["response"].([]interface{})
	userData := userArr[0].(map[string]interface{})
	userData["email"] = t.Extra("email")
	userData["user_id"] = t.Extra("user_id")
	userData["plugin_id"] = fmt.Sprintf("%d", v.GetMetadata().PluginID)
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

func convertUserData(mapIntr map[string]interface{}) map[string]string {
	mapStr := make(map[string]string, len(mapIntr))
	for key, value := range mapIntr {
		mapStr[key] = fmt.Sprintf("%v", value)
	}
	return mapStr
}
