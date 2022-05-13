package config

import (
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/ilyakaznacheev/cleanenv"
	"golang.org/x/xerrors"
	"gopkg.in/yaml.v3"
)

type StrategyConfig struct {
	Threshold    int `yaml:"threshold"`
	CandlesCount int `yaml:"candles_count"`
}

type Strategy struct {
	Name           string         `yaml:"name"`
	StrategyConfig StrategyConfig `yaml:"configuration"`
}

type TradingConfig struct {
	AccountId string   `yaml:"account_id"`
	Figi      string   `yaml:"figi"`
	Ticker    string   `yaml:"ticker"`
	Exchange  string   `yaml:"exchange"`
	Strategy  Strategy `yaml:"strategy"`
}

var tradingCfg TradingConfig

// LoadTradingsConfig Загружает торговую конфигурацию из файла
func LoadTradingsConfig(filename string) *TradingConfig {
	if err := cleanenv.ReadConfig(filename, tradingCfg); err != nil {
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
			configs = append(configs, LoadTradingsConfig(dirname+f.Name()))
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

func createDirIfNotExist(dirname string) error {
	if _, err := os.Stat(dirname); os.IsNotExist(err) {
		if err := os.Mkdir(dirname, writeMode); err != nil {
			return xerrors.Errorf("Ошибка создания папки: %v", err)
		}
	}
	return nil
}
