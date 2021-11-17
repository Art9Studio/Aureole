package file

import (
	"aureole/internal/plugins/kstorage"
)

const AdapterName = "file"

func init() {
	kstorage.Repository.Register(AdapterName, fileAdapter{})
}

type fileAdapter struct {
}
