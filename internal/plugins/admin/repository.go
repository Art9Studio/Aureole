package admin

import (
	"aureole/internal/configs"
	"aureole/internal/plugins/admin/types"
	"aureole/internal/plugins/core"
)

var Repository = core.CreateRepository()

// Adapter defines methods for admin adapter
type Adapter interface {
	// Create returns desired admin plugin depends on the given config
	Create(admin *configs.Admin) types.Admin
}
