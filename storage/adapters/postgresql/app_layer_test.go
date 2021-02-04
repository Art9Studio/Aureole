package postgresql

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	adapters "gouth/storage"
	"testing"
)

func TestIsUserCollectionExists(t *testing.T) {
	connUrl := "postgresql://root:password@localhost:5432/test"

	sess, err := adapters.Open(ConnectionString{connUrl})
	if err != nil {
		t.Fatalf("open connection by url: %v", err)
	}
	defer sess.Close()

	res, err := sess.IsUserCollectionExists(
		UserCollectionConfig{collection: "users", pk: "id", userId: "username", userConfirm: "password"})
	assert.NoError(t, err)
	assert.Equal(t, res, true)

	res, err = sess.IsUserCollectionExists(
		UserCollectionConfig{collection: "other", pk: "id", userId: "username", userConfirm: "password"})
	assert.NoError(t, err)
	assert.Equal(t, res, false)
}

func TestCreateUserCollection(t *testing.T) {
	connUrl := "postgresql://root:password@localhost:5432/test"

	sess, err := adapters.Open(ConnectionString{connUrl})
	if err != nil {
		t.Fatalf("open connection by url: %v", err)
	}
	defer sess.Close()

	// ALREADY EXISTS
	err = sess.CreateUserCollection(UserCollectionConfig{collection: "users", pk: "id", userId: "username", userConfirm: "password"})
	assert.Error(t, err, "Collection users already exists")

	// SUCCESS
	err = sess.CreateUserCollection(UserCollectionConfig{collection: "other", pk: "id", userId: "username", userConfirm: "password"})
	assert.NoError(t, err)
}

func TestInsertUser(t *testing.T) {
	connUrl := "postgresql://root:password@localhost:5432/test"

	sess, err := adapters.Open(ConnectionString{connUrl})
	if err != nil {
		t.Fatalf("open connection by url: %v", err)
	}
	defer sess.Close()

	res, err := sess.InsertUser(
		UserCollectionConfig{collection: "users", pk: "id", userId: "username", userConfirm: "password"},
		InsertUserData{userId: "hello", userConfirm: "secret_password"},
	)
	assert.NoError(t, err)

	fmt.Printf("new id: %v\n", res)
}
