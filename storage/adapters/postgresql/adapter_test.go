package postgresql

import (
	adapters "gouth/storage"
	"testing"
)

func TestOpenUrl(t *testing.T) {
	connUrl := "postgresql://root:password@localhost:5432/test"

	sess, err := adapters.Open(ConnectionString{connUrl})
	if err != nil {
		t.Fatalf("open connection by url: %v", err)
	}
	defer sess.Close()

	err = sess.Ping()
	if err != nil {
		t.Fatalf("ping connection: %v", err)
	}
}

func TestOpenConfig(t *testing.T) {
	connConf := ConnectionConfig{
		User:     "root",
		Password: "password",
		Host:     "localhost",
		Port:     "5432",
		Database: "test",
		Options:  nil,
	}

	sess, err := adapters.OpenConfig(connConf)
	if err != nil {
		t.Fatalf("open connection by config: %v", err)
	}
	defer sess.Close()

	err = sess.Ping()
	if err != nil {
		t.Fatalf("ping connection: %v", err)
	}
}
