package config

import (
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
)

type RobotConfig struct {
	TinkoffAccessToken string
	TinkoffApiEndpoint string `yaml:"tinkoff_api_endpoint"`
	// TODO БД
}

func NewRobotConfig() *RobotConfig {
	return &RobotConfig{
		TinkoffAccessToken: getEnv("TINKOFF_ACCESS_TOKEN", ""),
		TinkoffApiEndpoint: "",
	}
}

// LoadRobotConfig Загружает конфигурацию робота из файла
func LoadRobotConfig(filename string) *RobotConfig {
	robotConfig := NewRobotConfig()
	yamlData, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("Ошибка чтения конфигурации робота из файла: %v", err)
	}
	err = yaml.Unmarshal(yamlData, &robotConfig)
	if err != nil {
		log.Fatalf("Ошибка преобразования конфигурации робота: %v", err)
	}
	return robotConfig
}
