package vault

import (
	"aureole/internal/plugins/kstorage"
)

const AdapterName = "vault"

func init() {
	kstorage.Repository.Register(AdapterName, vaultAdapter{})
}

type vaultAdapter struct {
}
