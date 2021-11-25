package types

import "aureole/internal/plugins"

type Sender interface {
	plugins.MetaDataGetter
	Send(recipient, subject, tmplName string, tmplCtx map[string]interface{}) error
	SendRaw(recipient, subject, message string) error
}
