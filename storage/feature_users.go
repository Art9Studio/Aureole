package storage

type CollConfig struct {
	Name string `config:"name"`
	Pk   string `config:"pk,omitempty"`
}

type UserCollConfig struct {
	StorageName string `config:"storage"`
	Name        string `config:"name"`
	Pk          string `config:"pk,omitempty"`
	UserUnique  string `config:"user_unique"`
	UserConfirm string `config:"user_confirm"`
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
