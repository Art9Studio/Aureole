package types

type Sender interface {
	Init() error
	Send(recipient, subject, tmplName string, tmplCtx map[string]interface{}) error
	SendRaw(recipient, subject, message string) error
}
