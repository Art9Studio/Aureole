package types

type Sender interface {
	GetPluginID() string
	Send(recipient, subject, tmplName string, tmplCtx map[string]interface{}) error
	SendRaw(recipient, subject, message string) error
}
