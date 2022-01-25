package vault

import (
	"aureole/internal/plugins"
)

const adapterName = "vault"

func init() {
	plugins.CryptoStorageRepo.Register(adapterName, vaultAdapter{})
}

type vaultAdapter struct {
}
