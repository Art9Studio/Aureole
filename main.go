package main

import (
	"io/ioutil"
	"log"
)

// conf is global object that holds all project level settings variables
var conf ProjectConfig

func main() {
	data, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		panic(err)
	}
	conf.Init(data)
	log.Fatal(initRouter().Run())
}
