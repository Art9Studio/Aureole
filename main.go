package main

import (
	"aureole/internal/configs"
	pluginCore "aureole/internal/plugins/core"
	"aureole/internal/router"
	"aureole/internal/state"
	"log"
)

func main() {
	projConf, err := configs.LoadMainConfig()
	if err != nil {
		log.Panic(err)
	}

	project := &state.Project{}
	pluginCore.InitApi(project, router.Init())
	state.Init(projConf, project)
	state.ListPluginStatus(project)

	server, err := router.CreateServer(project.Apps)
	if err != nil {
		log.Panicf("router init: %v", err)
	}

	if err := server.Listen(":3000"); err != nil {
		log.Panicf("router start: %v", err)
	}
}
