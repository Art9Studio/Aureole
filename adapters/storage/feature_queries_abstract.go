package storage

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
