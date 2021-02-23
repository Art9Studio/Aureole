package postgresql

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"gouth/storage"
	"testing"
)

func createUsersSess(t *testing.T) storage.ConnSession {
	rawConnData := storage.RawStorageConfig{
		"connection_url": "postgresql://root:password@localhost:5432/test",
	}

	features := []string{"users"}
	usersSess, err := storage.Open(rawConnData, features)
	if err != nil {
		t.Fatalf("open connection by url: %v", err)
	}

	return usersSess
}

func Test_Session_IsCollExists(t *testing.T) {
	usersSess := createUsersSess(t)
	defer usersSess.Close()

	res, err := usersSess.IsCollExists(
		*storage.NewCollConfig("users", "id"))
	assert.NoError(t, err)
	assert.Equal(t, res, true)

	res, err = usersSess.IsCollExists(
		*storage.NewCollConfig("other", "id"))
	assert.NoError(t, err)
	assert.Equal(t, res, false)
}

func Test_Session_CreateUserColl(t *testing.T) {
	usersSess := createUsersSess(t)
	defer usersSess.Close()

	err := usersSess.CreateUserColl(*storage.NewUserCollConfig("users", "id", "username", "password"))
	assert.Contains(t, err.Error(), "already exists")

	err = usersSess.CreateUserColl(*storage.NewUserCollConfig("other", "id", "username", "password"))
	assert.NoError(t, err)

	err = usersSess.CreateUserColl(*storage.NewUserCollConfig("); drop table other; --", "id", "username", "password"))
	assert.NoError(t, err)

	isOtherExist, err := usersSess.IsCollExists(*storage.NewCollConfig("other", "id"))
	assert.True(t, isOtherExist)

	isDropExist, err := usersSess.IsCollExists(*storage.NewCollConfig("); drop table other; --", "id"))
	assert.True(t, isDropExist)
}

func Test_Session_InsertUser(t *testing.T) {
	usersSess := createUsersSess(t)
	defer usersSess.Close()

	res, err := usersSess.InsertUser(
		*storage.NewUserCollConfig("users", "id", "username", "password"),
		*storage.NewInsertUserData("hello", "secret"),
	)
	assert.NoError(t, err)
	fmt.Printf("new id: %v\n", res)
}
