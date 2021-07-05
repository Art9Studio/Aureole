package types

import "net/url"

type Authenticator interface {
	Init(string, *url.URL) error
}
