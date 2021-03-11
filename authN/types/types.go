package types

import "fmt"

type Type int

const (
	PasswordBased Type = iota
	Passwordless
)

var typeNames = [...]string{"password_based", "passwordless"}

func ToAuthNType(authType string) (Type, error) {
	for i, name := range typeNames {
		if name == authType {
			return i, nil
		}
	}

	return 0, fmt.Errorf("authenticate type '%s' is not declared", authType)
}

func (t Type) String() string {
	return typeNames[t]
}
