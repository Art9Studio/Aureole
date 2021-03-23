package types

type Sender interface {
	Send(string, string, map[string]interface{}) error
	SendRaw(string, string) error
}
