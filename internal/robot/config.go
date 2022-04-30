package robot

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v3"
)

type Config struct {
	TinkoffApiEndpoint string `yaml:"tinkoff_api_endpoint"`
	AccessToken        string `yaml:"access_token"`
}

func NewConfig() *Config {
	return &Config{
		TinkoffApiEndpoint: "",
		AccessToken:        "",
	}
}

func LoadConfig(filename string) *Config {
	config := NewConfig()
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("Config read err: %v", err)
	}
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		log.Fatalf("Config unmarshall err: %v", err)
	}
	return config
}
