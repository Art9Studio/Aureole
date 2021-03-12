package main

import (
	"gouth/adapters/authn"
	"gouth/config"
	"gouth/context"
	"gouth/context/types"
	"log"
)

// Project is global object that holds all project level settings variabces
var Project types.ProjectCtx

func main() {
	projConf, err := config.LoadMainConfig()
	if err != nil {
		log.Panic(err)
	}

	if err := context.InitContext(projConf, &Project); err != nil {
		log.Panic(err)
	}

	authn.InitRepository(&Project)

	router, err := initRouter()
	if err != nil {
		log.Panicf("router init: %v", err)
	}

	if err := router.Listen(":3000"); err != nil {
		log.Panicf("router start: %v", err)
	}
}
