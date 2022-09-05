package email

type config struct {
	Host               string   `mapstructure:"host" json:"host"`
	Username           string   `mapstructure:"username" json:"username"`
	Password           string   `mapstructure:"password" json:"password"`
	InsecureSkipVerify bool     `mapstructure:"insecure_skip_verify" json:"insecure_skip_verify"`
	From               string   `mapstructure:"from" json:"from"`
	Bcc                []string `mapstructure:"bcc" json:"bcc"`
	Cc                 []string `mapstructure:"cc" json:"cc"`
}
