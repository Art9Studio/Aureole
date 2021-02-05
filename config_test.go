package main

import (
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
	sess := conf.Apps["one"].Session

	err := sess.Ping()
	if err != nil {
		panic("error")
	}
}
