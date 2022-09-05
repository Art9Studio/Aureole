package twilio

type config struct {
	AccountSid  string `mapstructure:"account_sid" json:"account_sid"`
	AuthToken   string `mapstructure:"auth_token" json:"auth_token"`
	From        string `mapstructure:"from" json:"from"`
	Endpoint    string `mapstructure:"endpoint" json:"endpoint"`
	MessageType string `mapstructure:"message_type" json:"message_type"`
}
