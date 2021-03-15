package pwhasher

import (
	"aureole/configs"
	"aureole/plugins/pwhasher/types"
)

// New returns desired PwHasher depends on the given config
func New(conf *configs.Hasher) (types.PwHasher, error) {
	adapter, err := GetAdapter(conf.Type)
	if err != nil {
		return nil, err
	}

	return adapter.Get(conf)
}
