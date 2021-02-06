package storage

type CollConfig struct {
	name string `yaml:"name"`
	pk   string `yaml:"pk,omitempty"`
}

type UserCollConfig struct {
	CollConfig
	userID      string `yaml:"user_id"`
	userConfirm string ` yaml:"user_confirm"`
}

type InsertUserData struct {
	userID      string
	userConfirm string
}

type Application interface {
	// IsCollExists checks whether the given collection exists
	IsCollExists(CollConfig) (bool, error)

	// CreateUserColl creates user collection with traits passed by UserCollectionConfig
	CreateUserColl(UserCollConfig) error

	// InsertUser inserts user entity in the user collection
	InsertUser(UserCollConfig, InsertUserData) (JSONCollResult, error)

	GetUserPassword() error
}

func NewUserCollConfig(name string, pk string, userID string, userConfirm string) *UserCollConfig {
	return &UserCollConfig{CollConfig: CollConfig{name: name, pk: pk}, userID: userID, userConfirm: userConfirm}
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

func NewInsertUserData(userID string, userConfirm string) *InsertUserData {
	return &InsertUserData{userID: userID, userConfirm: userConfirm}
}

func (i InsertUserData) UserID() string {
	return i.userID
}

func (i InsertUserData) UserConfirm() string {
	return i.userConfirm
}
