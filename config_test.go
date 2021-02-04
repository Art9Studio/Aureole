package main

import (
	"gouth/config"
	"testing"
)

func TestConfig(t *testing.T) {
	conf := config.ProjectConfig{}
	yamlContent := []byte(`
        api_version: "0.1"
        apps:
          one:
            path_prefix: "/one"
            storage:
              connection_url: "postgresql://root:password@localhost:5432/test?sslmode=disable&search_path=public"`)

	conf.init(yamlContent)
	version := conf.APIVersion
	println(version)

}
