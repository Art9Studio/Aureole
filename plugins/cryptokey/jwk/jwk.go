package jwk

import (
	"context"
	"github.com/lestrrat-go/jwx/jwk"
	"net/url"
)

type Jwk struct {
	Conf *config
}

func (j *Jwk) Get(path string) (jwk.Set, error) {
	if _, err := url.ParseRequestURI(path); err != nil {
		return jwk.ReadFile(path)
	}

	return jwk.Fetch(context.Background(), path)
}
