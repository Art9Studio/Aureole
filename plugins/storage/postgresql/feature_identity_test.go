package postgresql

import (
	"aureole/configs"
	"aureole/internal/collections"
	"aureole/internal/plugins/storage"
	"aureole/internal/plugins/storage/types"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func createConnSess(t *testing.T) types.Storage {
	conf := &configs.Storage{
		Type: "",
		Name: "",
		Config: configs.RawConfig{
			"connection_url": "postgresql://root:password@localhost:5432/test",
		},
	}

	usersSess, err := storage.New(conf)
	if err != nil {
		t.Fatalf("open connection by url: %v", err)
	}

	return usersSess
}

func Test_Session_IsCollExists(t *testing.T) {
	usersSess := createConnSess(t)
	defer usersSess.Close()

	res, err := usersSess.IsCollExists(collections.Specification{Name: "users", Pk: "id"})
	assert.NoError(t, err)
	assert.Equal(t, res, true)

	res, err = usersSess.IsCollExists(collections.Specification{Name: "users", Pk: "id"})
	assert.NoError(t, err)
	assert.Equal(t, res, false)
}

func Test_Session_CreateIdentitytColl(t *testing.T) {
	usersSess := createConnSess(t)
	defer usersSess.Close()

	err := usersSess.CreateIdentityColl(collections.Specification{
		Name:      "users",
		Pk:        "id",
		FieldsMap: map[string]string{"identity": "username", "password": "password"},
	})
	assert.Contains(t, err.Error(), "already exists")

	err = usersSess.CreateIdentityColl(collections.Specification{
		Name:      "other",
		Pk:        "id",
		FieldsMap: map[string]string{"identity": "username", "password": "password"},
	})
	assert.NoError(t, err)

	err = usersSess.CreateIdentityColl(collections.Specification{
		Name:      "); drop table other; --",
		Pk:        "id",
		FieldsMap: map[string]string{"identity": "username", "password": "password"},
	})
	assert.NoError(t, err)

	isOtherExist, err := usersSess.IsCollExists(collections.Specification{Name: "other", Pk: "id"})
	assert.True(t, isOtherExist)

	isDropExist, err := usersSess.IsCollExists(collections.Specification{Name: "); drop table other; --", Pk: "id"})
	assert.True(t, isDropExist)
}

func Test_Session_InsertIdentity(t *testing.T) {
	usersSess := createConnSess(t)
	defer usersSess.Close()

	res, err := usersSess.InsertIdentity(
		collections.Specification{
			Name:      "users",
			Pk:        "id",
			FieldsMap: map[string]string{"identity": "username", "password": "password"},
		},
		types.InsertIdentityData{Identity: "hello", UserConfirm: "secret"},
	)
	assert.NoError(t, err)
	fmt.Printf("new id: %v\n", res)
}
