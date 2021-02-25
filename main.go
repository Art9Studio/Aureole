package main

import "log"

// conf is global object that holds all project level settings variables
var conf ProjectConfig

func main() {
	if err := conf.Init(); err != nil {
		log.Panic(err)
	}

	if err := initRouter().Run(); err != nil {
		log.Panicf("router init: %v", err)
	}
}
