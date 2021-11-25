package second_factor

import (
	"aureole/internal/configs"
	"aureole/internal/plugins/2fa/types"
	"aureole/internal/plugins/core"
)

var Repository = core.CreateRepository()

type Adapter interface {
	Create(*configs.SecondFactor) types.SecondFactor
}
