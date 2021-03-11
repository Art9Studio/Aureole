package storage

type CollConfig struct {
	Name string
	Pk   string
}

type UserCollConfig struct {
	StorageName string `c:"storage" v:"required"`
	Name        string `c:"name" v:"required"`
	Pk          string `c:"pk" v:"required"`
	UserUnique  string `c:"identity" v:"required"`
	UserConfirm string `c:"password" v:"required"`
}

type InsertUserData struct {
	UserUnique  interface{}
	UserConfirm interface{}
}

type Application interface {
	// IsCollExists checks whether the given collection exists
	IsCollExists(CollConfig) (bool, error)

	// CreateUserColl creates user collection with traits passed by UserCollectionConfig
	CreateUserColl(UserCollConfig) error

	// InsertUser inserts user entity in the user collection
	InsertUser(UserCollConfig, InsertUserData) (JSONCollResult, error)

	GetUserPassword(UserCollConfig, interface{}) (JSONCollResult, error)
}

func NewCollConfig(name string, pk string) *CollConfig {
	return &CollConfig{Name: name, Pk: pk}
}

func NewUserCollConfig(name string, pk string, userUnique string, userConfirm string) *UserCollConfig {
	return &UserCollConfig{Name: name, Pk: pk, UserUnique: userUnique, UserConfirm: userConfirm}
}

func NewInsertUserData(userUnique interface{}, userConfirm interface{}) *InsertUserData {
	return &InsertUserData{UserUnique: userUnique, UserConfirm: userConfirm}
}

func (conf UserCollConfig) ToCollConfig() CollConfig {
	return CollConfig{
		Name: conf.Name,
		Pk:   conf.Pk,
	}
}
