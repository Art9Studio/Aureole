package authN

import "gouth/config"

type Controller interface {
	GetRoutes() []Route
}

type Route struct {
	Path    string
	Handler func(interface{}) interface{}
}

type Type int

const (
	PasswordBased Type = iota
	Passwordless
)

func (d Type) String() string {
	return [...]string{"password_based", "passwordless"}[d]
}

func New(algoName string, rawConf *config.RawConfig) (Controller, error) {
	adapter, err := GetAdapter(algoName)
	if err != nil {
		return nil, err
	}

	return adapter.GetAuthNController(rawConf)
}
