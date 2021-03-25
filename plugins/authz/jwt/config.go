package jwt

type config struct {
	Alg     string   `mapstructure:"alg"`
	KidAlg  string   `mapstructure:"kid_alg"`
	Keys    []string `mapstructure:"keys"`
	Payload string   `mapstructure:"payload"`
}
