package facebook

import (
	"aureole/internal/identity"
	authzT "aureole/internal/plugins/authz/types"
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

		if ok, err := f.app.Filter(convertUserData(userData), f.rawConf.Filter); err != nil {
			return sendError(c, fiber.StatusBadRequest, err.Error())
		} else if !ok {
			return sendError(c, fiber.StatusBadRequest, "apple: input data doesn't pass filters")
		}

		var i map[string]interface{}
		if f.manager != nil {
			i, err = f.manager.OnUserAuthenticated(
				&identity.Credential{
					Name:  "email",
					Value: userData["email"],
				},
				&identity.Identity{
					Email: userData["email"].(string),
				},
				AdapterName,
				map[string]interface{}{
					"social_id": userData["id"],
					"user_data": userData,
				})
			if err != nil {
				return sendError(c, fiber.StatusInternalServerError, err.Error())
			}
		} else {
			i = map[string]interface{}{
				"email":     userData["email"],
				"provider":  AdapterName,
				"social_id": userData["id"],
				"user_data": userData,
			}
		}

		return f.authorizer.Authorize(c, authzT.NewPayload(f.authorizer, nil, i))
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
