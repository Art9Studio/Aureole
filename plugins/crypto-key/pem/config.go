package pem

type config struct {
	Alg             string `mapstructure:"alg" json:"alg"`
	Storage         string `mapstructure:"storage" json:"storage"`
	RefreshInterval int    `mapstructure:"refresh_interval" json:"refresh_interval"`
	RetriesNum      int    `mapstructure:"retries_num" json:"retries_num"`
	RetryInterval   int    `mapstructure:"retry_interval" json:"retry_interval"`
	PathPrefix      string
}
