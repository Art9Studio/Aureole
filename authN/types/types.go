package types

type Type int

const (
	PasswordBased Type = iota
	Passwordless
)

func (t Type) String() string {
	return [...]string{"password_based", "passwordless"}[t]
}
