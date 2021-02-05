package main

import (
	"io/ioutil"
	"log"
)

var conf ProjectConfig

func main() {
	data, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		panic(err)
	}
	conf.Init(data)
	log.Fatal(initRouter().Run())
}
