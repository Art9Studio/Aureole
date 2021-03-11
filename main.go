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
	router, err := initRouter()
	if err != nil {
		log.Panicf("router init: %v", err)
	}

	if err := router.Listen(":3000"); err != nil {
		log.Panicf("router start: %v", err)
	}
}
