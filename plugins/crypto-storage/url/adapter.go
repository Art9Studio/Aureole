package url

import (
	"aureole/internal/plugins"
)

const adapterName = "url"

func init() {
	plugins.CryptoStorageRepo.Register(adapterName, urlAdapter{})
}

type urlAdapter struct {
}
