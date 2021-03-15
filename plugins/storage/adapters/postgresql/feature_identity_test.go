package postgresql

import (
	"aureole/collections"
	"aureole/plugins/storage"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func createConnSess(t *testing.T) storage.ConnSession {
	rawConnData := storage.RawStorageConfig{
		"connection_url": "postgresql://root:password@localhost:5432/test",
	}

	usersSess, err := storage.New(rawConnData)
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
		storage.NewInsertIdentityData("hello", "secret"),
	)
	assert.NoError(t, err)
	fmt.Printf("new id: %v\n", res)
}
