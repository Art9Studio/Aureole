package main

import (
	"aureole/configs"
	"aureole/context"
	"aureole/context/types"
	"aureole/plugins/authn"
	"aureole/plugins/pwhasher"
	"log"
)

// Project is global object that holds all project level settings variables
var Project types.ProjectCtx

func main() {
	projConf, err := configs.LoadMainConfig()
	if err != nil {
		log.Panic(err)
	}

	pwhasher.InitRepository(&Project)
	authn.InitRepository(&Project)

	if err := context.InitContext(projConf, &Project); err != nil {
		log.Panic(err)
	}

	router, err := initRouter()
	if err != nil {
		log.Panicf("router init: %v", err)
	}

	if err := router.Listen(":3000"); err != nil {
		log.Panicf("router start: %v", err)
	}
}
