package types

import (
	coll "aureole/internal/collections"
	"github.com/gofrs/uuid"
	"time"
)

type InsertSessionData struct {
	UserId       int
	SessionToken uuid.UUID
	Expiration   time.Time
}

type Session interface {
	SetCleanInterval(time.Duration)

	StartCleaning(spec coll.Specification)

	CreateSessionColl(coll.Specification) error

	GetSession(coll.Specification, int) (JSONCollResult, error)

	InsertSession(coll.Specification, InsertSessionData) (JSONCollResult, error)

	DeleteSession(coll.Specification, int) (JSONCollResult, error)
}
