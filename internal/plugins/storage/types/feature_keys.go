package types

type KeysReadWrite interface {
	Read() ([]byte, error)
	Write([]byte) error
}
