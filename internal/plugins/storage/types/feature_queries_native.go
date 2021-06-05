package types

type NativeQueries interface {
	NativeQuery(string, ...interface{}) (JSONCollResult, error)
}
