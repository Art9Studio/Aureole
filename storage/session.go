package storage

// Session is an interface that defines methods for database session
type Session interface {
	Application

	// ConnConfig returns the connection config that was used to set up the adapter
	ConnConfig() ConnConfig

	// RelInfo returns information about tables relationships
	RelInfo() map[CollPair]RelInfo

	// Ping returns an error if the DBMS could not be reached
	Ping() error

	// RawExec executes the given sql query with no returning results
	RawExec(string, ...interface{}) error

	// RawQuery executes the given sql query and returns results
	RawQuery(string, ...interface{}) (JSONCollResult, error)

	// Read
	Read(string) (JSONCollResult, error)

	// Close terminates the currently active connection to the DBMS
	Close() error
}

type CollPair struct {
	from, to string
}

type RelInfo struct {
	// isO2M says about relationship between tables. One-to-Many or Many-to-One
	isO2M      bool
	fromFields []string
	toFields   []string
}

type JSONCollResult = interface{}
