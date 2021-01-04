package main

import (
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

//App ...
type App struct {
	PathPrefix string `yaml:"pathPrefix"`
	Data       string `yaml:"data"`
}

//Config ...
type Config struct {
	APIVersion string         `yaml:"apiVersion"`
	Apps       map[string]App `yaml:"apps"`
}

//GetConfig ...
func GetConfig(path string) (Config, error) {
	var conf Config
	data, err := ioutil.ReadFile(path)

	if err != nil {
		return conf, err
	}

	err = yaml.Unmarshal(data, &conf)

	if err != nil {
		return conf, err
	}

	return conf, nil
}
