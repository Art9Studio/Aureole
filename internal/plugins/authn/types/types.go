package types

type Authenticator interface {
	Initialize(string) error
}
