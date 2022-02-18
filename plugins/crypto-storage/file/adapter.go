package file

import (
	"aureole/internal/plugins"
)

const adapterName = "file"

func init() {
	plugins.CryptoStorageRepo.Register(adapterName, fileAdapter{})
}

type fileAdapter struct {
}
