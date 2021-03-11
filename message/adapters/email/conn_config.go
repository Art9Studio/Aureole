package email

// ConnConfig represents a parsed PostgreSQL connection URL
type ConnConfig struct {
	User     string
	Password string
	Host     string
	Port     string
	Database string
	Options  map[string]string
}

// AdapterName return the adapter name, that was used to set up connection
func (connConf ConnConfig) AdapterName() string {
	return AdapterName
}
