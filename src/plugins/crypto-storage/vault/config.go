package vault

type config struct {
	Path    string `mapstructure:"path"`
	Token   string `mapstructure:"token"`
	Address string `mapstructure:"address"`
}
