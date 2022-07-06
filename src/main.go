package main

import (
	"aureole/internal/core"
	"log"

	"aureole/internal/configs"
)

func main() {
	conf, err := configs.LoadMainConfig()
	if err != nil {
		log.Panic(err)
	}


	router := core.CreateRouter()
	project := core.InitProject(conf, router)
	err = core.RunServer(project, router)

	if err != nil {
		log.Panic(err)
	}
}
