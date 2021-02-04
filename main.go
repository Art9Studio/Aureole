package main

import (
	"gouth/config"
	"io/ioutil"
	"log"
)

var conf config.ProjectConfig

func main() {
	data, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		panic(err)
	}
	conf.Init(data)
	log.Fatal(initRouter().Run())
}
