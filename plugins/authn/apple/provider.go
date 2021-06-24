package apple

import (
	"encoding/json"
	"net/http"
	"net/url"
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
	resp, err := http.PostForm(c.Endpoint.TokenUrl, v)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return data, nil
}
