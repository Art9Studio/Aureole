package storage

type CollConfig struct {
	Name string
	Pk   string
}

type IdentityCollConfig struct {
	StorageName string `c:"storage" v:"required"`
	Name        string `c:"name" v:"required"`
	Pk          string `c:"pk" v:"required"`
	Identity    string `c:"identity" v:"required"`
	Password    string `c:"password" v:"required"`
}

type InsertUserData struct {
	Identity    interface{}
	UserConfirm interface{}
}

type Application interface {
	// IsCollExists checks whether the given collection exists
	IsCollExists(CollConfig) (bool, error)

	// CreateUserColl creates user collection with traits passed by UserCollectionConfig
	CreateIdentityColl(IdentityCollConfig) error

	// InsertUser inserts user entity in the user collection
	InsertIdentity(IdentityCollConfig, InsertUserData) (JSONCollResult, error)

	GetPasswordByIdentity(IdentityCollConfig, interface{}) (JSONCollResult, error)
}

func NewCollConfig(name string, pk string) *CollConfig {
	return &CollConfig{Name: name, Pk: pk}
}

func NewIdentityCollConfig(name string, pk string, userUnique string, userConfirm string) *IdentityCollConfig {
	return &IdentityCollConfig{Name: name, Pk: pk, Identity: userUnique, Password: userConfirm}
}

func NewInsertUserData(userUnique interface{}, userConfirm interface{}) *InsertUserData {
	return &InsertUserData{Identity: userUnique, UserConfirm: userConfirm}
}

func (conf IdentityCollConfig) ToCollConfig() CollConfig {
	return CollConfig{
		Name: conf.Name,
		Pk:   conf.Pk,
	}
}
