package types

import (
	"aureole/internal/plugins/core"
)

type Authenticator interface {
	Initialize(string) error
	GetRoutes() []*core.Route
}
