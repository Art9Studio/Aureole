package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_ProjectConfig_Init(t *testing.T) {
	conf := ProjectConfig{}

	yamlContent := []byte(`
        api_version: "0.1"
        apps:
          one:
            path_prefix: "/one"
            storage:
              connection_url: "postgresql://root:password@localhost:5432/test?sslmode=disable&search_path=public"`)
	conf.Init(yamlContent)
	sess := conf.Apps["one"].Conn
	assert.NoError(t, sess["users"].Ping())

	yamlContent = []byte(`
        api_version: "0.1"
        apps:
          three:
            path_prefix: "/three"
            storage:
              connection_config:
                adapter: "postgresql"
                username: "root"
                password: "password"
                host: "localhost"
                port: "5432"
                db_name: "test"
                options:
                  sslmode: "disable"
                  search_path: "public"`)
	conf.Init(yamlContent)
	sess = conf.Apps["three"].Conn
	assert.NoError(t, sess["users"].Ping())

	yamlContent = []byte(`
        api_version: "0.1"
        apps:
          one:
            path_prefix: "/one"
            storage:
              connection_url: "postgresql://root:password@localhost:5432/test?sslmode=disable&search_path=public"
            auth:
              use_existent_collection: false
              user_collection:
                name: "users"
                pk: "id"
                user_unique: "username"
                user_confirm: "password"`)
	conf.Init(yamlContent)
}
