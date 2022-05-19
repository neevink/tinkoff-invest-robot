package config

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/ilyakaznacheev/cleanenv"
	"golang.org/x/xerrors"
	"gopkg.in/yaml.v3"
)

type StrategyConfig struct {
	Name     string         `yaml:"name"`
	Interval string         `yaml:"interval"`
	Quantity int64          `yaml:"quantity"`
	Other    map[string]int `yaml:"other"`
}

type TradingConfig struct {
	AccountId      string         `yaml:"account_id"`
	IsSandbox      bool           `yaml:"is_sandbox"`
	Ticker         string         `yaml:"ticker"`
	Figi           string         `yaml:"figi"`
	Exchange       string         `yaml:"exchange"`
	StrategyConfig StrategyConfig `yaml:"strategy"`
}

// LoadTradingsConfig Загружает торговую конфигурацию из файла
func LoadTradingsConfig(filename string) *TradingConfig {
	var tradingCfg TradingConfig
	if err := cleanenv.ReadConfig(filename, &tradingCfg); err != nil {
		fmt.Printf("%v", err)
		log.Fatalf("Ошибка чтения торговой конфигурации %s: %v", filename, err)
	}
	return &tradingCfg
}

// LoadTradingConfigsFromDir Загружает торговые конфигурации из папки
func LoadTradingConfigsFromDir(dirname string) []*TradingConfig {
	if err := createDirIfNotExist(dirname); err != nil {
		log.Fatalf("Ошибка создания папки для сгенерированных конфигов: %v", err)
	}
	files, err := ioutil.ReadDir(dirname)
	configs := make([]*TradingConfig, 0)
	if err != nil {
		log.Fatalf("Ошибка чтения папки с конфигами: %v", err)
	}
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".yaml") {
			newCfg := LoadTradingsConfig(dirname + f.Name())
			configs = append(configs, newCfg)
		}
	}
	return configs
}

// WriteTradingConfig Сохраняет торговый конфиг
func WriteTradingConfig(dirname string, filename string, config *TradingConfig) error {
	if err := createDirIfNotExist(dirname); err != nil {
		log.Fatalf("Ошибка создания папки для сгенерированных конфигов: %v", err)
	}
	yamlData, err := yaml.Marshal(config)
	if err != nil {
		return xerrors.Errorf("Ошибка преобразования конфига: %v", err)
	}
	if err := ioutil.WriteFile(dirname+filename, yamlData, writeMode); err != nil {
		return xerrors.Errorf("Ошибка записи конфига в файл: %v", err)
	}
	return nil
}

// Создает папку если еще не была создана
func createDirIfNotExist(dirname string) error {
	if _, err := os.Stat(dirname); os.IsNotExist(err) {
		if err := os.Mkdir(dirname, writeMode); err != nil {
			return xerrors.Errorf("Ошибка создания папки: %v", err)
		}
	}
	return nil
}
