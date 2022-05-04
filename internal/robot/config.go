package robot

import (
	"io/fs"
	"io/ioutil"
	"log"
	"strings"

	"golang.org/x/xerrors"
	"gopkg.in/yaml.v3"
)

var (
	writeMode = fs.FileMode(0755)
)

type RobotConfig struct {
	TinkoffApiEndpoint string `yaml:"tinkoff_api_endpoint"`
	AccessToken        string `yaml:"access_token"`
}

type TradingConfig struct {
	AccountId       string `yaml:"account_id"`
	Figi            string `yaml:"figi"`
	TradingStrategy string `yaml:"trading_strategy"`
}

func NewRobotConfig() *RobotConfig {
	return &RobotConfig{
		TinkoffApiEndpoint: "",
		AccessToken:        "",
	}
}

func NewTradingConfig() *TradingConfig {
	return &TradingConfig{
		AccountId:       "",
		Figi:            "",
		TradingStrategy: "",
	}
}

func LoadRobotConfig(filename string) (*RobotConfig, error) {
	config := NewRobotConfig()
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, xerrors.Errorf("RobotConfig read err: %v", err)
	}
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return nil, xerrors.Errorf("RobotConfig unmarshall err: %v", err)
	}
	return config, nil
}

func LoadTradingConfigsFromDir(directoryPath string) []*TradingConfig {
	files, err := ioutil.ReadDir(directoryPath)
	configs := make([]*TradingConfig, 0)
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".yaml") {
			configs = append(configs, LoadTradingConfig(directoryPath+f.Name()))
		}
	}
	return configs
}

func LoadTradingConfig(filename string) *TradingConfig {
	config := NewTradingConfig()
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("TradingConfig read err: %v", err)
	}
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		log.Fatalf("TradingConfig unmarshall err: %v", err)
	}
	return config
}

func WriteTradingConfig(path string, config *TradingConfig) error {
	yamlData, err := yaml.Marshal(config)
	if err != nil {
		return xerrors.Errorf("trading config marshall err: %v", err)
	}
	if err := ioutil.WriteFile(path, yamlData, writeMode); err != nil {
		return xerrors.Errorf("can't write trading config to the file: %v", err)
	}
	return nil
}
