package main

import (
	"aureole/internal/configs"
	"aureole/internal/router"
	"aureole/internal/state"
	"log"
)

func main() {
	conf, err := configs.LoadMainConfig()
	if err != nil {
		log.Panic(err)
	}

	project := &state.Project{}
	state.Init(conf, project)
	state.ListPluginStatus(project)

	server, err := router.CreateServer(project.Apps)
	if err != nil {
		log.Panicf("router init: %v", err)
	}

	if err := server.Listen(":3000"); err != nil {
		log.Panicf("router start: %v", err)
	}
}
