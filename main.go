package main

import (
	"aureole/configs"
	"aureole/context"
	"aureole/context/types"
	"aureole/internal/plugins/authn"
	"aureole/internal/plugins/authz"
	"aureole/internal/plugins/core"
	"aureole/internal/plugins/cryptokey"
	"aureole/internal/plugins/pwhasher"
	"aureole/internal/plugins/sender"
	"aureole/internal/plugins/storage"
	"log"
)

var Project types.ProjectCtx

func main() {

	projConf, err := configs.LoadMainConfig()
	if err != nil {
		log.Panic(err)
	}

	pluginApi := core.InitPluginApi(&Project)

	pwhasher.InitRepository(pluginApi)
	storage.InitRepository(pluginApi)
	authn.InitRepository(pluginApi)
	authz.InitRepository(pluginApi)
	sender.InitRepository(pluginApi)
	cryptokey.InitRepository(pluginApi)

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
