package types

import (
	"aureole/internal/collections"
	"time"

	"github.com/gofrs/uuid"
)

type InsertSessionData struct {
	UserId       interface{}
	SessionToken uuid.UUID
	Expiration   time.Time
}

type Session interface {
	GetSession(collections.Spec, int) (JSONCollResult, error)

	InsertSession(collections.Spec, InsertSessionData) (JSONCollResult, error)

	DeleteSession(collections.Spec, int) (JSONCollResult, error)

	SetCleanInterval(int)

	StartCleaning(spec collections.Spec)
}
