package postgresql

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"gouth/storage"
	"testing"
)

func Test_Session_IsCollExists(t *testing.T) {
	rawConnData := storage.RawConnData{
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
	rawConnData := storage.RawConnData{
		"connection_url": "postgresql://root:password@localhost:5432/test",
	}

	sess, err := storage.Open(rawConnData)
	if err != nil {
		t.Fatalf("open connection by url: %v", err)
	}
	defer sess.Close()

	err = sess.CreateUserColl(*storage.NewUserCollConfig("users", "id", "username", "password"))
	assert.Error(t, err, "Name users already exists")

	err = sess.CreateUserColl(*storage.NewUserCollConfig("other", "id", "username", "password"))
	assert.NoError(t, err)
}

func Test_Session_InsertUser(t *testing.T) {
	rawConnData := storage.RawConnData{
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
