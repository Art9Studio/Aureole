package storage

// Open attempts to establish a connection with a database by ConnectionString
func Open(connUrl ConnectionString) (Session, error) {
	a, err := GetAdapter(connUrl.AdapterName())
	if err != nil {
		return nil, err
	}

	return a.OpenUrl(connUrl)
}

// OpenConfig attempts to establish a connection with a database by ConnectionConfig
func OpenConfig(connConf ConnectionConfig) (Session, error) {
	a, err := GetAdapter(connConf.AdapterName())
	if err != nil {
		return nil, err
	}

	return a.OpenConfig(connConf)
}
