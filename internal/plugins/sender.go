package plugins

import (
	"aureole/internal/configs"
	"fmt"
)

var SenderRepo = createRepository()

// SenderAdapter defines methods for authentication plugins
type (
	SenderAdapter interface {
		// Create returns desired messenger depends on the given config
		Create(*configs.Sender) Sender
	}

	Sender interface {
		MetaDataGetter
		Send(recipient, subject, tmplName string, tmplCtx map[string]interface{}) error
		SendRaw(recipient, subject, message string) error
	}
)

// NewSender returns desired messenger depends on the given config
func NewSender(conf *configs.Sender) (Sender, error) {
	a, err := SenderRepo.Get(conf.Type)
	if err != nil {
		return nil, err
	}

	adapter, ok := a.(SenderAdapter)
	if !ok {
		return nil, fmt.Errorf("trying to cast adapter was failed: %v", err)
	}

	return adapter.Create(conf), nil
}
