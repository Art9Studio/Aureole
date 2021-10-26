package apple

import (
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	AuthUrl  string = "https://appleid.apple.com/auth/authorize"
	TokenUrl string = "https://appleid.apple.com/auth/token"
)

type (
	Config struct {
		ClientId     string
		TeamId       string
		KeyId        string
		ClientSecret string
		Endpoint     Endpoint
		RedirectUrl  string
		Scopes       []string
	}

	Endpoint struct {
		AuthUrl  string
		TokenUrl string
	}
)

func (c *Config) AuthCodeURL(state string) string {
	u, _ := url.Parse(c.Endpoint.AuthUrl)
	v := u.Query()
	v.Set("response_type", "code")
	v.Set("state", state)
	v.Set("client_id", c.ClientId)
	v.Set("response_mode", "form_post")
	v.Set("redirect_uri", c.RedirectUrl)
	v.Set("scope", strings.Join(c.Scopes, " "))
	u.RawQuery = v.Encode()
	return u.String()
}

func (c *Config) Exchange(code string) (map[string]interface{}, error) {
	v := url.Values{}
	v.Set("client_id", c.ClientId)
	v.Set("client_secret", c.ClientSecret)
	v.Set("code", code)
	v.Set("grant_type", "authorization_code")
	return doRequest(c, v)
}

func doRequest(c *Config, v url.Values) (map[string]interface{}, error) {
	ctx := context.Background()
	r, err := http.NewRequestWithContext(ctx, http.MethodPost, c.Endpoint.TokenUrl, strings.NewReader(v.Encode()))
	if err != nil {
		return nil, err
	}

	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Content-Length", strconv.Itoa(len(v.Encode())))

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data = make(map[string]interface{})
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(b, &data); err != nil {
		return nil, errors.Wrapf(err, "decode - %s", string(b))
	}

	return data, nil
}
