package config

import (
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"golang.org/x/xerrors"
	"gopkg.in/yaml.v3"
)

const (
	writeMode = fs.FileMode(0755)
)

type Share struct {
	Ticker string `yaml:"ticker"`
	Figi   string `yaml:"figi"`
}

type Config struct {
	TinkoffApiEndpoint string  `yaml:"tinkoff_api_endpoint"`
	AccessToken        string  `yaml:"access_token"`
	AccountId          string  `yaml:"account_id"`
	Shares             []Share `yaml:"shares"`
}

type TradingConfig struct { // per ticker config
	AccountId string `yaml:"account_id"`
	Figi      string `yaml:"figi"`
	Strategy  string `yaml:"strategy"`
}

func LoadConfig(filename string) *Config {
	config := &Config{}
	yamlData, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("Ошибка чтения конфига из файла: %v", err)
	}
	err = yaml.Unmarshal(yamlData, &config)
	if err != nil {
		log.Fatalf("Ошибка преобразования конфига: %v", err)
	}
	return config
}

func LoadTradingsConfig(filename string) *TradingConfig {
	config := &TradingConfig{}
	yamlData, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("Ошибка чтения конфига из файла: %v", err)
	}
	err = yaml.Unmarshal(yamlData, &config)
	if err != nil {
		log.Fatalf("Ошибка преобразования конфига: %v", err)
	}
	return config
}

func LoadTradingConfigsFromDir(directoryPath string) []*TradingConfig {
	err := os.MkdirAll(directoryPath, writeMode)
	if err != nil {
		log.Fatalf("Ошибка создания папки для сгенерированных конфигов: %v", err)
	}
	files, err := ioutil.ReadDir(directoryPath)
	configs := make([]*TradingConfig, 0)
	if err != nil {
		log.Fatalf("Ошибка чтения папки с конфигами: %v", err)
	}
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".yaml") {
			configs = append(configs, LoadTradingsConfig(directoryPath+f.Name()))
		}
	}
	return configs
}

func WriteConfig(path string, config *Config) error {
	yamlData, err := yaml.Marshal(config)
	if err != nil {
		return xerrors.Errorf("Ошибка преобразования конфига: %v", err)
	}
	if err := ioutil.WriteFile(path, yamlData, writeMode); err != nil {
		return xerrors.Errorf("Ошибка записи конфига в файл: %v", err)
	}
	return nil
}
