package types

import "aureole/internal/plugins"

type KeyStorage interface {
	plugins.MetaDataGetter
	Read(v *[]byte) (ok bool, err error)
	Write(v []byte) error
}
