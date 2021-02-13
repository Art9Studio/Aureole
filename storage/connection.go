package storage

// RawConnData represents unparsed data from config file
type RawConnData = map[string]interface{}

// ConnConfig represents a parsed connection config
type ConnConfig interface {
	// String returns the connection url that is going to be passed to the adapter
	String() (string, error)

	// AdapterName return the adapter Name, that was used to set up connection
	AdapterName() string
}
