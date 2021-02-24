package storage

type CollConfig struct {
	Name string `yaml:"name"`
	Pk   string `yaml:"pk,omitempty"`
}

type UserCollConfig struct {
	StorageName string `yaml:"storage"`
	Name        string `yaml:"name"`
	Pk          string `yaml:"pk,omitempty"`
	UserUnique  string `yaml:"user_unique"`
	UserConfirm string `yaml:"user_confirm"`
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
