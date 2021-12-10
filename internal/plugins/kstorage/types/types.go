package types

type (
	KeyStorage interface {
		Init() error
		Read(v *[]byte) (ok bool, err error)
		Write(v []byte) error
	}
)
