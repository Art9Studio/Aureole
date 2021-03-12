package collections

import (
	"gouth/configs"
)

// todo: reorganize this structures
type (
	Collection struct {
		Type          string
		Name          string
		UseExistent   bool
		Specification Specification
	}

	Specification struct {
		Name      string
		Pk        string
		FieldsMap map[string]string
	}
)

func NewCollection(collType string, conf *configs.Collections) (Collection, error) {

	//switch collConf.Type {
	//case "identity":
	//	identityColl := storage.NewIdentityCollection(collConf.Storage, collConf.Specification)
	//	identityStorage := conf.Storages[collConf.Storage]
	//
	//	isExists, err := identityStorage.IsCollExists(identityColl.ToCollConfig())
	//	if err != nil {
	//		return err
	//	}
	//
	//	useExistent := collConf.Specification["use_existent"].(bool)
	//	if !useExistent && !isExists {
	//		if err = identityStorage.CreateUserColl(*identityColl); err != nil {
	//			return err
	//		}
	//	} else if useExistent && !isExists {
	//		return fmt.Errorf("identity collection '%s' is not found", collConf.Name)
	//	}
	//
	//	conf.Collections[collConf.Name] = identityColl
	//case "session":
	//	sessionColl := storage.NewSessionCollection(collConf.Storage, collConf.Specification)
	//	sessionStorage := conf.Storages[collConf.Storage]
	//
	//	isExists, err := sessionStorage.IsCollExists(sessionColl.ToCollConfig())
	//	if err != nil {
	//		return err
	//	}
	//
	//	useExistent := collConf.Specification["use_existent"].(bool)
	//	if !useExistent && !isExists {
	//		if err = sessionStorage.CreateUserColl(*sessionColl); err != nil {
	//			return err
	//		}
	//	} else if useExistent && !isExists {
	//		return fmt.Errorf("session collection '%s' is not found", collConf.Name)
	//	}
	//
	//	conf.Collections[collConf.Name] = sessionColl
	//}
	return Collection{}, nil
}
