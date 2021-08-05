package types

import _interface "aureole/internal/context/interface"

type Authenticator interface {
	Init(app _interface.AppCtx) error
}
