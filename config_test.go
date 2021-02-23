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
            storages:
              "main db":
                connection_url: "postgresql://root:password@localhost:5432/test?sslmode=disable&search_path=public"
            main:
              user_collection:
                storage: "main db"`)
	conf.Init(yamlContent)
	connSess := conf.Apps["one"].SessByFeature["users"]
	assert.NoError(t, connSess.Ping())

	yamlContent = []byte(`
        api_version: "0.1"
        apps:
          two:
            path_prefix: "/two"
            storages:
              "main db":
                connection_config:
                  adapter: "postgresql"
                  username: "root"
                  password: "password"
                  host: "localhost"
                  port: "5432"
                  db_name: "test"
                  options:
                    sslmode: "disable"
                    search_path: "public"
            main:
              user_collection:
                storage: "main db"`)
	conf.Init(yamlContent)
	connSess = conf.Apps["two"].SessByFeature["users"]
	assert.NoError(t, connSess.Ping())
}
