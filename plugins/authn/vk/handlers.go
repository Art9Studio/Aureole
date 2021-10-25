package vk

import (
	authzT "aureole/internal/plugins/authz/types"
	storageT "aureole/internal/plugins/storage/types"
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

		var (
			//filter  []storageT.Filter
			socAuth *storageT.SocialAuthData
			user    *storageT.IdentityData
		)
		//s := &v.coll.Spec
		email, _ := userData["email"]
		/*if ok {
			filter = []storageT.Filter{
				{Name: s.FieldsMap["email"].Name, Value: email},
				{Name: s.FieldsMap["provider"].Name, Value: Provider},
			}
		} else if v.identity.Email.IsRequired {
			return sendError(c, fiber.StatusInternalServerError, "required field email is not provided")
		} else {
			filter = []storageT.Filter{
				{Name: s.FieldsMap["social_id"].Name, Value: userData["user_id"]},
				{Name: s.FieldsMap["provider"].Name, Value: Provider},
			}
		}

		exist, err := v.storage.IsSocialAuthExist(s, filter)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		if exist {
			rawSocAuth, err := v.storage.GetSocialAuth(s, filter)
			if err != nil {
				return sendError(c, fiber.StatusInternalServerError, err.Error())
			}
			socAuth = storageT.NewSocialAuthData(rawSocAuth, s.FieldsMap)

			if socAuth.UserId != nil {
				iSpecs := &v.identity.Collection.Spec
				rawUser, err := v.storage.GetIdentity(v.identity, []storageT.Filter{
					{Name: iSpecs.FieldsMap["id"].Name, Value: socAuth.UserId},
				})
				if err != nil {
					return sendError(c, fiber.StatusInternalServerError, err.Error())
				}
				user = storageT.NewIdentityData(rawUser, iSpecs.FieldsMap)
			}
		} else {*/
		socAuth = &storageT.SocialAuthData{
			SocialId: fmt.Sprintf("%v", userData["user_id"]),
			Email:    email,
			Provider: Provider,
			UserData: userData,
		}
		/*user, err = createOrLink(v, socAuth)
			if err != nil {
				return sendError(c, fiber.StatusInternalServerError, err.Error())
			}
		}*/

		payload, err := createAuthzPayload(socAuth, user)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}
		return v.authorizer.Authorize(c, payload)
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

func createOrLink(v *vk, socAuth *storageT.SocialAuthData) (user *storageT.IdentityData, err error) {
	i := v.identity
	s := &i.Collection.Spec

	if socAuth.Email == nil {
		socAuth.UserId, err = v.storage.InsertIdentity(i, &storageT.IdentityData{})
		if err != nil {
			return nil, err
		}
	} else {
		filter := []storageT.Filter{{Name: s.FieldsMap["email"].Name, Value: socAuth.Email}}
		exist, err := v.storage.IsIdentityExist(i, filter)
		if err != nil {
			return nil, err
		}

		if exist {
			rawUser, err := v.storage.GetIdentity(i, filter)
			if err != nil {
				return nil, err
			}
			user = storageT.NewIdentityData(rawUser, s.FieldsMap)
			socAuth.UserId = user.Id
		} else {
			newUser := &storageT.IdentityData{Email: socAuth.Email}
			socAuth.UserId, err = v.storage.InsertIdentity(i, newUser)
			if err != nil {
				return nil, err
			}
		}
	}

	socAuth.Id, err = v.storage.InsertSocialAuth(&v.coll.Spec, socAuth)
	return user, err
}

func createAuthzPayload(socAuth *storageT.SocialAuthData, user *storageT.IdentityData) (*authzT.Payload, error) {
	var payload *authzT.Payload
	jsonUserData, err := json.Marshal(socAuth.UserData)
	if err != nil {
		return nil, err
	}

	if user != nil {
		payload = &authzT.Payload{
			Id:         user.Id,
			SocialId:   socAuth.SocialId,
			Username:   user.Username,
			Phone:      user.Phone,
			Email:      user.Email,
			UserData:   socAuth.UserData,
			Additional: user.Additional,
		}
	} else {
		payload = &authzT.Payload{
			SocialId: socAuth.SocialId,
			Email:    socAuth.Email,
			UserData: string(jsonUserData),
		}
	}
	return payload, nil
}
