package vault

import (
	"aureole/internal/plugins"
)

const adapterName = "vault"

func init() {
	plugins.KeyStorageRepo.Register(adapterName, vaultAdapter{})
}

type vaultAdapter struct {
}
