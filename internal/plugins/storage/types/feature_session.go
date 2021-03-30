package types

import (
	coll "aureole/internal/collections"
	"time"
)

type Session interface {
	SetGCInterval(time.Duration)

	CreateSessionColl(coll.Specification) error

	// Get gets the value for the given key.
	// `nil, nil` is returned when the key does not exist
	Get(coll.Specification, string) ([]byte, error)

	// Set stores the given value for the given key along
	// with an expiration value, 0 means no expiration.
	// Empty key or value will be ignored without an error.
	Set(coll.Specification, string, []byte, time.Duration) error

	// Delete deletes the value for the given key.
	// It returns no error if the storage does not contain the key,
	Delete(coll.Specification, string) error

	// Reset resets the storage and delete all keys.
	Reset(coll.Specification) error
}
