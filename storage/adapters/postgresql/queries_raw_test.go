package postgresql

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"gouth/storage"
	"testing"
)

func Test_Session_RawExec(t *testing.T) {
	rawConnData := storage.RawConnData{
		"connection_url": "postgresql://root:password@localhost:5432/test",
	}

	sess, err := storage.Open(rawConnData)
	if err != nil {
		t.Fatalf("open connection by url: %v", err)
	}
	defer sess.Close()

	err = sess.RawExec("create table test (id serial);")
	assert.NoError(t, err)

	err = sess.RawExec("drop table test;")
	assert.NoError(t, err)
}

func Test_Session_RawQuery(t *testing.T) {
	rawConnData := storage.RawConnData{
		"connection_url": "postgresql://root:password@localhost:5432/test",
	}

	sess, err := storage.Open(rawConnData)
	if err != nil {
		t.Fatalf("open connection by url: %v", err)
	}
	defer sess.Close()

	// FIELD WITH NO KEYS
	res, err := sess.RawQuery("select username from users where id=1;")
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
	res, err = sess.RawQuery("select row_to_json(t) from (select username from users where id=1) t;")
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
	res, err = sess.RawQuery("select json_agg(t.id) from (select p.id from users join posts p on users.id = p.user_id where users.id=1) t;")
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
	res, err = sess.RawQuery("select json_agg(t) from (select p.id from users join posts p on users.id = p.user_id where users.id=1) t;")
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
	res, err = sess.RawQuery("select json_build_array(username, password) from users where id=1;")
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
	res, err = sess.RawQuery("select row_to_json(t) from (select username, password from users where id=1) t;")
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
	res, err = sess.RawQuery("select json_agg(json_build_array(p.id, p.content))  from users join posts p on users.id = p.user_id where users.id=1;")
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
	res, err = sess.RawQuery("select json_agg(t) from (select p.id, p.content from users join posts p on users.id = p.user_id where users.id=1 limit 1) t;")
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
