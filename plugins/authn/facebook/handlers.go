package facebook

import (
	"context"
	"encoding/json"
	"fmt"
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

func convertUserData(mapIntr map[string]interface{}) map[string]string {
	mapStr := make(map[string]string, len(mapIntr))
	for key, value := range mapIntr {
		mapStr[key] = fmt.Sprintf("%v", value)
	}
	return mapStr
}
