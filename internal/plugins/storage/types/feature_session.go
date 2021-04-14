package types

import (
	coll "aureole/internal/collections"
	"github.com/gofrs/uuid"
	"time"
)

type InsertSessionData struct {
	UserId       int32
	SessionToken uuid.UUID
	Expiration   time.Time
}

type Session interface {
	CreateSessionColl(coll.Specification) error

	GetSession(coll.Specification, int) (JSONCollResult, error)

	InsertSession(coll.Specification, InsertSessionData) (JSONCollResult, error)

	DeleteSession(coll.Specification, int) (JSONCollResult, error)

	SetCleanInterval(int)

	StartCleaning(spec coll.Specification)
}
