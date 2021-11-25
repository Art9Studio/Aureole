package types

import "aureole/internal/plugins"

type (
	Storage interface {
		plugins.MetaDataGetter
		Set(k string, v interface{}, exp int) error
		Get(k string, v interface{}) (ok bool, err error)
		Delete(k string) error
		Exists(k string) (found bool, err error)
		Close() error
	}
)
