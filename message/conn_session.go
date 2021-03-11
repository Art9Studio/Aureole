package message

// ConnSession is an interface that defines methods for database session
type ConnSession interface {
	Send()

	// ConnConfig returns the connection config that was used to set up the adapter
	GetConfig() ConnConfig

	// Close terminates the currently active connection
	Close() error
}
