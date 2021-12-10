package internal

import (
	"aureole/internal/encrypt"
	"aureole/internal/jwt"
	state "aureole/internal/state/interface"
)

func InitCore(p state.ProjectState) {
	jwt.Init(p)
	encrypt.Init(p)
}
