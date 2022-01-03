package url

import (
	"aureole/internal/plugins"
)

const adapterName = "url"

func init() {
	plugins.KeyStorageRepo.Register(adapterName, urlAdapter{})
}

type urlAdapter struct {
}
