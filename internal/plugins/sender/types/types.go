package types

type Sender interface {
	Initialize() error
	Send(string, string, string, map[string]interface{}) error
	SendRaw(string, string, string) error
}
