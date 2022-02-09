package email

type config struct {
	Host               string   `mapstructure:"host"`
	Username           string   `mapstructure:"username"`
	Password           string   `mapstructure:"password"`
	InsecureSkipVerify bool     `mapstructure:"insecure_skip_verify"`
	From               string   `mapstructure:"from"`
	Bcc                []string `mapstructure:"bcc"`
	Cc                 []string `mapstructure:"cc"`
}
