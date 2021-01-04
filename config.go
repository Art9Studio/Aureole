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

	if err = yaml.Unmarshal(data, &conf); err != nil {
		return conf, err
	}

	return conf, nil
}
