package types

type Authenticator interface {
	Init(string) error
}
