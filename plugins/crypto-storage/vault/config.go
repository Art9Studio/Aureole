package vault

type config struct {
	Path    string `mapstructure:"path" json:"path"`
	Token   string `mapstructure:"token" json:"token"`
	Address string `mapstructure:"address" json:"address"`
}
