package postgresql

import (
	"aureole/internal/configs"
	"aureole/internal/plugins/storage"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Session_RawExec(t *testing.T) {
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
	defer usersSess.Close()

	err = usersSess.RawExec("create table test (id serial);")
	assert.NoError(t, err)

	err = usersSess.RawExec("drop table test;")
	assert.NoError(t, err)
}

func Test_Session_RawQuery(t *testing.T) {
	// todo: check it
	conf := &configs.Storage{
		Type: "postgresql",
		Name: "",
		Config: configs.RawConfig{
			"connection_url": "postgresql://root:password@localhost:5432/aureole",
		},
	}

	usersSess, err := storage.New(conf)
	if err != nil {
		t.Fatalf("open connection by url: %v", err)
	}
	defer usersSess.Close()

	// FIELD WITH NO KEYS
	res, err := usersSess.RawQuery("select username from users where id=$1;", 1)
	if err != nil {
		t.Fatalf("raw query (field, no keys): %v", err)
	}
	switch casted := res.(type) {
	case string:
		fmt.Printf("type is 'string', value is '%v'\n", casted)
	default:
		t.Fatalf("unexpected type %T", casted)
	}

	// FIELDS WITH KEYS
	res, err = usersSess.RawQuery("select row_to_json(t) from (select username from users where id=$1) t;", 1)
	if err != nil {
		t.Fatalf("raw query (fields, keys): %v", err)
	}
	switch casted := res.(type) {
	case map[string]interface{}:
		fmt.Printf("type is 'map[string]interface{}', value is '%v'\n", casted)
	default:
		t.Fatalf("unexpected type %T", casted)
	}

	// FIELD ARRAY WITH NO KEYS
	res, err = usersSess.RawQuery("select json_agg(t.id) from (select p.id from users join posts p on users.id = p.user_id where users.id=$1) t;", 1)
	if err != nil {
		t.Fatalf("raw query (field arr, no keys): %v", err)
	}
	switch casted := res.(type) {
	case []interface{}:
		fmt.Printf("type is '[]interface{}', value is '%v'\n", casted)
	default:
		t.Fatalf("unexpected type %T", casted)
	}

	// FIELD ARRAY WITH KEYS
	res, err = usersSess.RawQuery("select json_agg(t) from (select p.id from users join posts p on users.id = p.user_id where users.id=$1) t;", 1)
	if err != nil {
		t.Fatalf("raw query (field arr, keys): %v", err)
	}
	switch casted := res.(type) {
	case []interface{}:
		fmt.Printf("type is '[]interface{}', value is '%v'\n", casted)
	default:
		t.Fatalf("unexpected type %T", casted)
	}

	// ROW WITH NO KEYS
	res, err = usersSess.RawQuery("select json_build_array(username, password) from users where id=$1;", 1)
	if err != nil {
		t.Fatalf("raw query (row, no keys): %v", err)
	}
	switch casted := res.(type) {
	case interface{}:
		fmt.Printf("type is 'interface{}', value is '%v'\n", casted)
	default:
		t.Fatalf("unexpected type %T", casted)
	}

	// ROW WITH KEYS
	res, err = usersSess.RawQuery("select row_to_json(t) from (select username, password from users where id=$1) t;", 1)
	if err != nil {
		t.Fatalf("raw query (row, keys): %v", err)
	}
	switch casted := res.(type) {
	case map[string]interface{}:
		fmt.Printf("type is 'map[string]interface{}', value is '%v'\n", casted)
	default:
		t.Fatalf("unexpected type %T", casted)
	}

	// ROW ARRAY WITH NO KEYS
	res, err = usersSess.RawQuery("select json_agg(json_build_array(p.id, p.content))  from users join posts p on users.id = p.user_id where users.id=$1;", 1)
	if err != nil {
		t.Fatalf("raw query (row arr, no keys): %v", err)
	}
	switch casted := res.(type) {
	case []interface{}:
		fmt.Printf("type is '[]interface{}', value is '%v'\n", casted)
	default:
		t.Fatalf("unexpected type %T", casted)
	}

	//ROW ARRAY WITH KEYS
	res, err = usersSess.RawQuery("select json_agg(t) from (select p.id, p.content from users join posts p on users.id = p.user_id where users.id=$1 limit 1) t;", 1)
	if err != nil {
		t.Fatalf("raw query (row arr, keys): %v", err)
	}
	switch casted := res.(type) {
	case []interface{}:
		fmt.Printf("type is '[]interface{}', value is '%v'\n", casted)
	default:
		t.Fatalf("unexpected type %T", casted)
	}
}
