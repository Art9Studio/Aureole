package internal

import (
	"aureole/internal/jwt"
	_interface "aureole/internal/state/interface"
)

func InitCore(p _interface.ProjectState) {
	jwt.Init(p)
}
