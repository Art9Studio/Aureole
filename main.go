package main

import "log"

func main() {
	conf.init("config.yaml")
	log.Fatal(initRouter().Run())
}
