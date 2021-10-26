package types

import (
	_interface "aureole/internal/state/interface"
)

type Authenticator interface {
	Init(app _interface.AppState) error
}
