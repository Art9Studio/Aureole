package file

import (
	"aureole/internal/plugins"
)

const adapterName = "file"

func init() {
	plugins.KeyStorageRepo.Register(adapterName, fileAdapter{})
}

type fileAdapter struct {
}
