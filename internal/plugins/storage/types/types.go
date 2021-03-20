package types

// Storage is an interface that defines methods for database session
type Storage interface {
	Application

	CheckFeaturesAvailable([]string) error

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
