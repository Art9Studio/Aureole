package storage

// ConnConfig represents a parsed connection configs
type ConnConfig interface {
	// String returns the connection url that is going to be passed to the adapter
	String() (string, error)

	// AdapterName return the adapter Name, that was used to set up connection
	AdapterName() string
}
