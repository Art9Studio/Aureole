package facebook

import (
	authzT "aureole/internal/plugins/authz/types"
	storageT "aureole/internal/plugins/storage/types"
	"context"
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"net/url"
	"strings"
)

func GetAuthCode(f *facebook) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		u := f.provider.AuthCodeURL("state")
		return c.Redirect(u)
	}
}

func Login(f *facebook) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		state := c.Query("state")
		if state != "state" {
			return sendError(c, fiber.StatusBadRequest, "invalid state")
		}
		code := c.Query("code")
		if code == "" {
			return sendError(c, fiber.StatusBadRequest, "code not found")
		}

		userData, err := getUserData(f, code)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		var (
			socAuth *storageT.SocialAuthData
			user    *storageT.IdentityData
		)
		email := userData["email"]
		/*s := &f.coll.Spec
		filter := []storageT.Filter{
			{Name: s.FieldsMap["email"].Name, Value: email},
			{Name: s.FieldsMap["provider"].Name, Value: Provider},
		}
		exist, err := f.storage.IsSocialAuthExist(s, filter)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		if exist {
			rawSocAuth, err := f.storage.GetSocialAuth(s, []storageT.Filter{
				{Name: s.FieldsMap["email"].Name, Value: email},
				{Name: s.FieldsMap["provider"].Name, Value: Provider},
			})
			if err != nil {
				return sendError(c, fiber.StatusInternalServerError, err.Error())
			}
			socAuth = storageT.NewSocialAuthData(rawSocAuth, s.FieldsMap)

			if socAuth.UserId != nil {
				iSpecs := &f.identity.Collection.Spec
				rawUser, err := f.storage.GetIdentity(f.identity, []storageT.Filter{
					{Name: iSpecs.FieldsMap["id"].Name, Value: socAuth.UserId},
				})
				if err != nil {
					return sendError(c, fiber.StatusInternalServerError, err.Error())
				}
				user = storageT.NewIdentityData(rawUser, iSpecs.FieldsMap)
			}
		} else {*/
		socAuth = &storageT.SocialAuthData{
			SocialId: userData["id"],
			Email:    email,
			Provider: Provider,
			UserData: userData,
		}
		/*user, err = createOrLink(f, socAuth)
			if err != nil {
				return sendError(c, fiber.StatusInternalServerError, err.Error())
			}

			if err = f.storage.SetEmailVerified(&f.identity.Collection.Spec, []storageT.Filter{
				{Name: s.FieldsMap["email"].Name, Value: socAuth.Email},
			}); err != nil {
				return sendError(c, fiber.StatusInternalServerError, err.Error())
			}
		}*/

		payload, err := createAuthzPayload(f, socAuth, user)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}
		return f.authorizer.Authorize(c, payload)
	}
}

func getUserData(f *facebook, code string) (map[string]interface{}, error) {
	ctx := context.Background()
	t, err := f.provider.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}

	u, err := getUserInfoUrl(f)
	if err != nil {
		return nil, err
	}

	client := f.provider.Client(ctx, t)
	resp, err := client.Get(u)
	if err != nil {
		return nil, err
	}

	var userData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&userData); err != nil {
		return nil, err
	}
	return userData, nil
}

func getUserInfoUrl(f *facebook) (string, error) {
	u, err := url.Parse("https://graph.facebook.com/me")
	if err != nil {
		return "", err
	}

	q := u.Query()
	fieldsStr := strings.Join(f.conf.Fields, ",")
	q.Set("fields", fieldsStr)
	u.RawQuery = q.Encode()

	return u.String(), nil
}

/*func createOrLink(f *facebook, socAuth *storageT.SocialAuthData) (*storageT.IdentityData, error) {
	var user *storageT.IdentityData
	i := f.identity
	s := &i.Collection.Spec
	filter := []storageT.Filter{{Name: s.FieldsMap["email"].Name, Value: socAuth.Email}}
	exist, err := f.storage.IsIdentityExist(i, filter)
	if err != nil {
		return nil, err
	}

	if exist {
		rawUser, err := f.storage.GetIdentity(i, filter)
		if err != nil {
			return nil, err
		}
		user = storageT.NewIdentityData(rawUser, s.FieldsMap)
		socAuth.UserId = user.Id
	} else {
		newUser := &storageT.IdentityData{Email: socAuth.Email}
		socAuth.UserId, err = f.storage.InsertIdentity(i, newUser)
		if err != nil {
			return nil, err
		}
	}

	socAuth.Id, err = f.storage.InsertSocialAuth(&f.coll.Spec, socAuth)
	return user, err
}*/

func createAuthzPayload(f *facebook, socAuth *storageT.SocialAuthData, user *storageT.IdentityData) (*authzT.Payload, error) {
	payload := authzT.NewPayload(f.authorizer, f.storage)
	jsonUserData, err := json.Marshal(socAuth.UserData)
	if err != nil {
		return nil, err
	}

	payload.SocialId = socAuth.SocialId
	payload.Email = socAuth.Email
	payload.UserData = string(jsonUserData)

	if user != nil {
		payload.Id = user.Id
		payload.Username = user.Username
		payload.Phone = user.Phone
		payload.Additional = user.Additional
	}

	return payload, nil
}
