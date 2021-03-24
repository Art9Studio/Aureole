package types

type Sender interface {
	Send(string, string, string, map[string]interface{}) error
	SendRaw(string, string, string) error
}
