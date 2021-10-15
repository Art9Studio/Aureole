package google

import (
	authzT "aureole/internal/plugins/authz/types"
	storageT "aureole/internal/plugins/storage/types"
	"context"
	"encoding/json"
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
			return sendError(c, fiber.StatusBadRequest, "invalid state")
		}
		code := c.Query("code")
		if code == "" {
			return sendError(c, fiber.StatusBadRequest, "code not found")
		}

		jwtT, err := getJwt(g, code)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, errors.Wrap(err, "error while exchange").Error())
		}

		var (
			socAuth *storageT.SocialAuthData
			user    *storageT.IdentityData
		)
		email, _ := jwtT.Get("email")
		/*s := &g.coll.Spec
		filter := []storageT.Filter{
			{Name: s.FieldsMap["email"].Name, Value: email},
			{Name: s.FieldsMap["provider"].Name, Value: Provider},
		}
		exist, err := g.storage.IsSocialAuthExist(s, filter)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		if exist {
			rawSocAuth, err := g.storage.GetSocialAuth(s, filter)
			if err != nil {
				return sendError(c, fiber.StatusInternalServerError, err.Error())
			}
			socAuth = storageT.NewSocialAuthData(rawSocAuth, s.FieldsMap)

			if socAuth.UserId != nil {
				iSpecs := &g.identity.Collection.Spec
				rawUser, err := g.storage.GetIdentity(g.identity, []storageT.Filter{
					{Name: iSpecs.FieldsMap["id"].Name, Value: socAuth.UserId},
				})
				if err != nil {
					return sendError(c, fiber.StatusInternalServerError, err.Error())
				}
				user = storageT.NewIdentityData(rawUser, iSpecs.FieldsMap)
			}
		} else {*/
		userData, err := jwtT.AsMap(context.Background())
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		socAuth = &storageT.SocialAuthData{
			Email:    email,
			Provider: Provider,
			UserData: userData,
		}
		socAuth.SocialId, _ = jwtT.Get("sub")
		/*user, err = createOrLink(g, socAuth)
			if err != nil {
				return sendError(c, fiber.StatusInternalServerError, err.Error())
			}

			if verified, ok := jwtT.Get("email_verified"); ok && verified.(bool) {
				if err = g.storage.SetEmailVerified(&g.identity.Collection.Spec, []storageT.Filter{
					{Name: s.FieldsMap["email"].Name, Value: socAuth.Email},
				}); err != nil {
					return sendError(c, fiber.StatusInternalServerError, err.Error())
				}
			}
		}*/

		payload, err := createAuthzPayload(socAuth, user)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}
		return g.authorizer.Authorize(c, payload)
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

func createOrLink(g *google, socAuth *storageT.SocialAuthData) (*storageT.IdentityData, error) {
	var user *storageT.IdentityData
	i := g.identity
	s := &i.Collection.Spec
	filter := []storageT.Filter{{Name: s.FieldsMap["email"].Name, Value: socAuth.Email}}
	exist, err := g.storage.IsIdentityExist(i, filter)
	if err != nil {
		return nil, err
	}

	if exist {
		rawUser, err := g.storage.GetIdentity(i, filter)
		if err != nil {
			return nil, err
		}
		user := storageT.NewIdentityData(rawUser, s.FieldsMap)
		socAuth.UserId = user.Id
	} else {
		newUser := &storageT.IdentityData{Email: socAuth.Email}
		socAuth.UserId, err = g.storage.InsertIdentity(i, newUser)
		if err != nil {
			return nil, err
		}
	}

	socAuth.Id, err = g.storage.InsertSocialAuth(&g.coll.Spec, socAuth)
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
