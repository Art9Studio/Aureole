package main

import (
	"aureole/internal/configs"
	"aureole/internal/context"
	pluginCore "aureole/internal/plugins/core"
	"aureole/internal/router"
	"log"
)

var Project *context.ProjectCtx

func main() {
	projConf, err := configs.LoadMainConfig()
	if err != nil {
		log.Panic(err)
	}

	Project = &context.ProjectCtx{}

	r := router.Init()
	pluginCore.InitApi(Project, r)
	if err := context.Init(projConf, Project); err != nil {
		log.Panic(err)
	}

	server, err := router.CreateServer(Project.Apps)
	if err != nil {
		log.Panicf("router init: %v", err)
	}

	if err := server.Listen(":3000"); err != nil {
		log.Panicf("router start: %v", err)
	}
}
