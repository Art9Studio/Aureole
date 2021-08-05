package apple

import (
	authzT "aureole/internal/plugins/authz/types"
	storageT "aureole/internal/plugins/storage/types"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/lestrrat-go/jwx/jwt"
	"strconv"
)

func GetAuthCode(a *apple) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		u := a.provider.AuthCodeURL("state")
		return c.Redirect(u)
	}
}

func Login(a *apple) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		data := struct {
			State string
			Code  string
		}{}
		if err := c.BodyParser(&data); err != nil {
			return err
		}
		if data.State != "state" {
			return sendError(c, fiber.StatusBadRequest, "invalid state")
		}
		if data.Code == "" {
			return sendError(c, fiber.StatusBadRequest, "code not found")
		}

		jwtT, err := getJwt(a, data.Code)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		var (
			socAuth *storageT.SocialAuthData
			user    *storageT.IdentityData
		)
		email, _ := jwtT.Get("email")
		s := &a.coll.Spec
		filter := []storageT.Filter{
			{Name: s.FieldsMap["email"].Name, Value: email},
			{Name: s.FieldsMap["provider"].Name, Value: Provider},
		}
		exist, err := a.storage.IsSocialAuthExist(s, filter)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		if exist {
			rawSocAuth, err := a.storage.GetSocialAuth(s, filter)
			if err != nil {
				return sendError(c, fiber.StatusInternalServerError, err.Error())
			}
			socAuth = storageT.NewSocialAuthData(rawSocAuth, s.FieldsMap)

			if socAuth.UserId != nil {
				iSpecs := &a.identity.Collection.Spec
				rawUser, err := a.storage.GetIdentity(a.identity, []storageT.Filter{
					{Name: iSpecs.FieldsMap["id"].Name, Value: socAuth.UserId},
				})
				if err != nil {
					return sendError(c, fiber.StatusInternalServerError, err.Error())
				}
				user = storageT.NewIdentityData(rawUser, iSpecs.FieldsMap)
			}
		} else {
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
			user, err = createOrLink(a, socAuth)
			if err != nil {
				return sendError(c, fiber.StatusInternalServerError, err.Error())
			}

			if v, ok := jwtT.Get("email_verified"); ok {
				if verified, err := strconv.ParseBool(v.(string)); err == nil && verified {
					if err := a.storage.SetEmailVerified(&a.identity.Collection.Spec, []storageT.Filter{
						{Name: s.FieldsMap["email"].Name, Value: socAuth.Email},
					}); err != nil {
						return sendError(c, fiber.StatusInternalServerError, err.Error())
					}
				}
			}
		}

		authzCtx, err := createAuthzCtx(socAuth, user)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}
		return a.authorizer.Authorize(c, authzCtx)
	}
}

func getJwt(a *apple, code string) (jwt.Token, error) {
	t, err := a.provider.Exchange(code)
	if err != nil {
		return nil, err
	}
	idToken := t["id_token"]

	keySet := a.publicKey.GetPublicSet()
	if keySet == nil {
		return nil, fmt.Errorf("apple: cannot get public set from %s", a.conf.PublicKey)
	}

	return jwt.ParseString(
		idToken.(string),
		jwt.WithAudience(a.provider.ClientId),
		jwt.WithKeySet(keySet))
}

func createOrLink(a *apple, socAuth *storageT.SocialAuthData) (*storageT.IdentityData, error) {
	var user *storageT.IdentityData
	i := a.identity
	s := &i.Collection.Spec
	filter := []storageT.Filter{{Name: s.FieldsMap["email"].Name, Value: socAuth.Email}}
	exist, err := a.storage.IsIdentityExist(i, filter)
	if err != nil {
		return nil, err
	}

	if exist {
		rawUser, err := a.storage.GetIdentity(i, filter)
		if err != nil {
			return nil, err
		}
		user = storageT.NewIdentityData(rawUser, s.FieldsMap)
		socAuth.UserId = user.Id
	} else {
		newUser := &storageT.IdentityData{Email: socAuth.Email}
		socAuth.UserId, err = a.storage.InsertIdentity(i, newUser)
		if err != nil {
			return nil, err
		}
	}

	socAuth.Id, err = a.storage.InsertSocialAuth(&a.coll.Spec, socAuth)
	return user, err
}

func createAuthzCtx(socAuth *storageT.SocialAuthData, user *storageT.IdentityData) (*authzT.Context, error) {
	var authzCtx *authzT.Context
	jsonUserData, err := json.Marshal(socAuth.UserData)
	if err != nil {
		return nil, err
	}

	if user != nil {
		authzCtx = &authzT.Context{
			Id:         user.Id,
			SocialId:   socAuth.SocialId,
			Username:   user.Username,
			Phone:      user.Phone,
			Email:      user.Email,
			UserData:   socAuth.UserData,
			Additional: user.Additional,
		}
	} else {
		authzCtx = &authzT.Context{
			SocialId: socAuth.SocialId,
			Email:    socAuth.Email,
			UserData: string(jsonUserData),
		}
	}
	return authzCtx, nil
}
