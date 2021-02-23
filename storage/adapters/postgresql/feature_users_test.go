package postgresql

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"gouth/storage"
	"testing"
)

func createSession(t *testing.T) storage.ConnSession {
	rawConnData := storage.RawStorageConfig{
		"connection_url": "postgresql://root:password@localhost:5432/test",
	}

	features := []string{"users"}
	sess, err := storage.Open(rawConnData, features)
	if err != nil {
		t.Fatalf("open connection by url: %v", err)
	}

	return sess
}

func Test_Session_IsCollExists(t *testing.T) {
	connSess := createSession(t)
	defer connSess.Close()

	res, err := connSess.IsCollExists(
		*storage.NewCollConfig("users", "id"))
	assert.NoError(t, err)
	assert.Equal(t, res, true)

	res, err = connSess.IsCollExists(
		*storage.NewCollConfig("other", "id"))
	assert.NoError(t, err)
	assert.Equal(t, res, false)
}

func Test_Session_CreateUserColl(t *testing.T) {
	connSess := createSession(t)
	defer connSess.Close()

	err := connSess.CreateUserColl(*storage.NewUserCollConfig("users", "id", "username", "password"))
	assert.Contains(t, err.Error(), "already exists")

	err = connSess.CreateUserColl(*storage.NewUserCollConfig("other", "id", "username", "password"))
	assert.NoError(t, err)

	err = connSess.CreateUserColl(*storage.NewUserCollConfig("); drop table other; --", "id", "username", "password"))
	assert.NoError(t, err)

	isOtherExist, err := connSess.IsCollExists(*storage.NewCollConfig("other", "id"))
	assert.True(t, isOtherExist)

	isDropExist, err := connSess.IsCollExists(*storage.NewCollConfig("); drop table other; --", "id"))
	assert.True(t, isDropExist)
}

func Test_Session_InsertUser(t *testing.T) {
	connSess := createSession(t)
	defer connSess.Close()

	res, err := connSess.InsertUser(
		*storage.NewUserCollConfig("users", "id", "username", "password"),
		*storage.NewInsertUserData("hello", "secret"),
	)
	assert.NoError(t, err)
	fmt.Printf("new id: %v\n", res)
}
