package types

type CollPair struct {
	from, to string //nolint
}

type RelInfo struct {
	// isO2M says about relationship between tables. One-to-Many or Many-to-One
	isO2M      bool     //nolint
	fromFields []string //nolint
	toFields   []string //nolint
}

type JSONCollResult = interface{}
