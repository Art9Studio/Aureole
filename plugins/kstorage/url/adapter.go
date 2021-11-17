package url

import (
	"aureole/internal/plugins/kstorage"
)

const AdapterName = "url"

func init() {
	kstorage.Repository.Register(AdapterName, urlAdapter{})
}

type urlAdapter struct {
}
