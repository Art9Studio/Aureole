package types

import (
	coll "aureole/internal/collections"
	"time"
)

type InsertSessionData struct {
	UserId       int
	SessionToken string
	Expiration   time.Duration
}

type Session interface {
	SetGCInterval(time.Duration)

	CreateSessionColl(coll.Specification) error

	GetSession(coll.Specification, int) (JSONCollResult, error)

	InsertSession(coll.Specification, InsertSessionData) (JSONCollResult, error)

	DeleteSession(coll.Specification, int) (JSONCollResult, error)
}
