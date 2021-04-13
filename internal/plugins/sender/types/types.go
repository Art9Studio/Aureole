package types

import (
	"aureole/internal"
)

type Sender interface {
	internal.Initializer
	Send(string, string, string, map[string]interface{}) error
	SendRaw(string, string, string) error
}
