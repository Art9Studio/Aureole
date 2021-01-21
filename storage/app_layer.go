package storage

type UserCollectionConfig interface {
	Collection() string
	Pk() string
	UserId() string
	UserConfirm() string
}

type InsertUserData interface {
	UserId() string
	UserConfirm() string
}

type Application interface {
	// IsUserCollectionExists checks whether the given collection exists
	IsUserCollectionExists(UserCollectionConfig) (bool, error)

	// CreateUserCollection creates user collection with traits passed by UserCollectionConfig
	CreateUserCollection(UserCollectionConfig) error

	// InsertUser inserts user entity in the user collection
	InsertUser(UserCollectionConfig, InsertUserData) (JSONCollectionResult, error)

	GetUserPassword() error
}
