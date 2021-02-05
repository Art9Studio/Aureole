package storage

// TODO: solve naming interface-struct pairs problem

type CollectionConfig interface {
	Name() string
	PK() string
}

type UserCollectionConfig interface {
	CollectionConfig
	UserID() string
	UserConfirm() string
}

type InsertionUserData interface {
	UserID() string
	UserConfirm() string
}

type Application interface {
	// IsCollExists checks whether the given collection exists
	IsCollExists(CollectionConfig) (bool, error)

	// CreateUserColl creates user collection with traits passed by UserCollectionConfig
	CreateUserColl(UserCollectionConfig) error

	// InsertUser inserts user entity in the user collection
	InsertUser(UserCollectionConfig, InsertionUserData) (JSONCollResult, error)

	GetUserPassword() error
}

type UserCollConfig struct {
	name        string `yaml:"name"`
	pk          string `yaml:"pk,omitempty`
	userID      string `yaml:"user_id"`
	userConfirm string ` yaml:"user_confirm"`
}

func NewUserCollConfig(name string, pk string, userID string, userConfirm string) UserCollConfig {
	return UserCollConfig{name: name, pk: pk, userID: userID, userConfirm: userConfirm}
}

func (u UserCollConfig) Name() string {
	return u.name
}

func (u UserCollConfig) PK() string {
	return u.pk
}

func (u UserCollConfig) UserID() string {
	return u.userID
}

func (u UserCollConfig) UserConfirm() string {
	return u.userConfirm
}

type InsertUserData struct {
	userID      string
	userConfirm string
}

func NewInsertUserData(userID string, userConfirm string) *InsertUserData {
	return &InsertUserData{userID: userID, userConfirm: userConfirm}
}

func (i InsertUserData) UserID() string {
	return i.userID
}

func (i InsertUserData) UserConfirm() string {
	return i.userConfirm
}
