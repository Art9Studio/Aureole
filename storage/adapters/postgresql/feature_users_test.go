package postgresql

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"gouth/storage"
	"testing"
)

func Test_Session_IsCollExists(t *testing.T) {
	rawConnData := storage.RawStorageConfig{
		"connection_url": "postgresql://root:password@localhost:5432/test",
	}

	sess, err := storage.Open(rawConnData)
	if err != nil {
		t.Fatalf("open connection by url: %v", err)
	}
	defer sess.Close()

	res, err := sess.IsCollExists(
		*storage.NewCollConfig("users", "id"))
	assert.NoError(t, err)
	assert.Equal(t, res, true)

	res, err = sess.IsCollExists(
		*storage.NewCollConfig("other", "id"))
	assert.NoError(t, err)
	assert.Equal(t, res, false)
}

func Test_Session_CreateUserColl(t *testing.T) {
	rawConnData := storage.RawStorageConfig{
		"connection_url": "postgresql://root:password@localhost:5432/test",
	}

	sess, err := storage.Open(rawConnData)
	if err != nil {
		t.Fatalf("open connection by url: %v", err)
	}
	defer sess.Close()

	err = sess.CreateUserColl(*storage.NewUserCollConfig("users", "id", "username", "password"))
	assert.Contains(t, err.Error(), "already exists")

	err = sess.CreateUserColl(*storage.NewUserCollConfig("other", "id", "username", "password"))
	assert.NoError(t, err)

	err = sess.CreateUserColl(*storage.NewUserCollConfig("); drop table other; --", "id", "username", "password"))
	assert.NoError(t, err)

	isOtherExist, err := sess.IsCollExists(*storage.NewCollConfig("other", "id"))
	assert.True(t, isOtherExist)

	isDropExist, err := sess.IsCollExists(*storage.NewCollConfig("); drop table other; --", "id"))
	assert.True(t, isDropExist)
}

func Test_Session_InsertUser(t *testing.T) {
	rawConnData := storage.RawStorageConfig{
		"connection_url": "postgresql://root:password@localhost:5432/test",
	}

	sess, err := storage.Open(rawConnData)
	if err != nil {
		t.Fatalf("open connection by url: %v", err)
	}
	defer sess.Close()

	res, err := sess.InsertUser(
		*storage.NewUserCollConfig("users", "id", "username", "password"),
		*storage.NewInsertUserData("hello", "secret"),
	)
	assert.NoError(t, err)
	fmt.Printf("new id: %v\n", res)
}
