package main

import (
	"gouth/config"
	"log"
)

// Project is global object that holds all project level settings variables
var Project config.Project

func main() {

	if err := config.LoadMainConfig(&Project); err != nil {
		log.Panic(err)
	}

	if err := initRouter().Run(); err != nil {
		log.Panicf("router init: %v", err)
	}
}
