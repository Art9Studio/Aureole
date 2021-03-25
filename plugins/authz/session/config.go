package session

type config struct {
	Collection string `mapstructure:"collection"`
	Storage    string `mapstructure:"storage"`
	Domain     string `mapstructure:"domain"`
	Path       string `mapstructure:"path"`
	MaxAge     string `mapstructure:"max_age"`
	Secure     bool   `mapstructure:"secure"`
	HttpOnly   bool   `mapstructure:"http_only"`
}
