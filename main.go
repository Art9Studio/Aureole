package main

import (
	"aureole/internal/configs"
	"aureole/internal/core"
	"log"
)

func main() {
	conf, err := configs.LoadMainConfig()
	if err != nil {
		log.Panic(err)
	}

	core.Init(conf)
	err = core.RunServer()
	if err != nil {
		log.Panic(err)
	}
}
