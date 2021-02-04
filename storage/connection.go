package storage

type ConnectionString interface {
	// String returns the connection url that is going to be passed to the adapter
	String() string

	// AdapterName return the adapter name, that was used to set up connection
	AdapterName() string
}

// ConnectionConfig represents a parsed connection url
type ConnectionConfig interface {
	// String returns the connection url that is going to be passed to the adapter
	String() (string, error)

	// AdapterName return the adapter name, that was used to set up connection
	AdapterName() string
}
