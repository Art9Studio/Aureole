package storage

type CollConfig struct {
	Name string
	Pk   string
}

type IdentityCollConfig struct {
	Name     string
	Pk       string
	Identity string
	Password string
}

type InsertIdentityData struct {
	Identity    interface{}
	UserConfirm interface{}
}

type Application interface {
	// IsCollExists checks whether the given collection exists
	IsCollExists(CollConfig) (bool, error)

	// CreateUserColl creates user collection with traits passed by UserCollectionConfig
	CreateIdentityColl(IdentityCollConfig) error

	// InsertUser inserts user entity in the user collection
	InsertIdentity(IdentityCollConfig, InsertIdentityData) (JSONCollResult, error)

	GetPasswordByIdentity(IdentityCollConfig, interface{}) (JSONCollResult, error)
}

func NewCollConfig(name string, pk string) *CollConfig {
	return &CollConfig{Name: name, Pk: pk}
}

func NewIdentityCollConfig(name string, pk string, userUnique string, userConfirm string) *IdentityCollConfig {
	return &IdentityCollConfig{Name: name, Pk: pk, Identity: userUnique, Password: userConfirm}
}

func NewInsertIdentityData(userUnique interface{}, userConfirm interface{}) *InsertIdentityData {
	return &InsertIdentityData{Identity: userUnique, UserConfirm: userConfirm}
}

func (conf IdentityCollConfig) ToCollConfig() CollConfig {
	return CollConfig{
		Name: conf.Name,
		Pk:   conf.Pk,
	}
}
