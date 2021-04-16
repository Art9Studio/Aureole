package main

import (
	"aureole/configs"
	"aureole/context"
	"aureole/context/types"
	"aureole/internal/plugins/core"
	"aureole/internal/router"
	"log"
)

var Project types.ProjectCtx

func main() {
	projConf, err := configs.LoadMainConfig()
	if err != nil {
		log.Panic(err)
	}

	core.Init(&Project)
	if err := context.Init(projConf, &Project); err != nil {
		log.Panic(err)
	}

	r, err := router.Init(&Project)
	if err != nil {
		log.Panicf("router init: %v", err)
	}

	if err := r.Listen(":3000"); err != nil {
		log.Panicf("router start: %v", err)
	}
}
