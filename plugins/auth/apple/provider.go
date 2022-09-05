package apple

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

const (
	authUrl  string = "https://appleid.apple.com/auth/authorize"
	tokenUrl string = "https://appleid.apple.com/auth/token"
)

type (
	providerConfig struct {
		clientId     string
		teamId       string
		keyId        string
		clientSecret string
		endpoint     endpoint
		redirectUrl  string
		scopes       []string
	}

	endpoint struct {
		authUrl  string
		tokenUrl string
	}
)

func (c *providerConfig) authCodeURL(state string) string {
	u, _ := url.Parse(c.endpoint.authUrl)
	v := u.Query()
	v.Set("response_type", "code")
	v.Set("state", state)
	v.Set("client_id", c.clientId)
	v.Set("response_mode", "form_post")
	v.Set("redirect_uri", c.redirectUrl)
	v.Set("scope", strings.Join(c.scopes, " "))
	u.RawQuery = v.Encode()
	return u.String()
}

func (c *providerConfig) exchange(code string) (map[string]interface{}, error) {
	v := url.Values{}
	v.Set("client_id", c.clientId)
	v.Set("client_secret", c.clientSecret)
	v.Set("code", code)
	v.Set("grant_type", "authorization_code")
	return doRequest(c, v)
}

func doRequest(c *providerConfig, v url.Values) (map[string]interface{}, error) {
	ctx := context.Background()
	r, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint.tokenUrl, strings.NewReader(v.Encode()))
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
