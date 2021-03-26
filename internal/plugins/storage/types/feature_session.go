package types

import "aureole/internal/collections"

type InsertSessionData struct {
	UserId    interface{}
	SessionId interface{}
}

type Session interface {
	CreateSessionColl(collections.Specification) error

	InsertSession(collections.Specification, InsertSessionData) (JSONCollResult, error)

	GetSessionId(collections.Specification, interface{}) (JSONCollResult, error)
}
