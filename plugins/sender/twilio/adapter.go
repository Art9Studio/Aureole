package twilio

import (
	"aureole/internal/configs"
	"aureole/internal/plugins"
)

// adapterName is the internal name of the adapter
const adapterName = "twilio"

// init initializes package by register adapter
func init() {
	plugins.SenderRepo.Register(adapterName, adapter{})
}

// twilioAdapter represents adapter for the email messenger
type adapter struct {
}

func (adapter) Create(conf *configs.Sender) plugins.Sender {
	return &sender{rawConf: conf}
}
