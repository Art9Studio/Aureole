package types

type (
	KeyStorage interface {
		GetPluginID() string
		Read(v *[]byte) (ok bool, err error)
		Write(v []byte) error
	}
)
