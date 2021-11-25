package types

import "aureole/internal/plugins"

type PwHasher interface {
	plugins.MetaDataGetter
	HashPw(pw string) (hashPw string, err error)
	ComparePw(pw string, hashPw string) (match bool, err error)
}
