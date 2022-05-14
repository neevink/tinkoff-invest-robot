package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type RobotConfig struct {
	TinkoffAccessToken string `env:"TINKOFF_ACCESS_TOKEN"`
	TinkoffApiEndpoint string `yaml:"tinkoff_api_endpoint"`
	// TODO БД
}

// LoadRobotConfig Загружает конфигурацию робота из файла
func LoadRobotConfig(filename string) *RobotConfig {
	var robotCfg RobotConfig
	err := cleanenv.ReadConfig(filename, &robotCfg)
	if err != nil {
		log.Fatalf("Ошибка чтения конфигурации робота %s: %v", filename, err)
	}
	return &robotCfg
}
