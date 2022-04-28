package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v3"

	"tinkoff-invest-bot/internal/robot"
	"tinkoff-invest-bot/robot/pkg/engine"
)

func main() {
	config := loadConfig("./configs/main.yaml")

	ctx := context.Background()

	robotInstance := engine.New(config)
	if err := robotInstance.Run(ctx); err != nil {
		fmt.Println(err)
	}

}

func loadConfig(filename string) *robot.Config {
	config := robot.NewConfig()
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("Config read err #%v", err)
	}
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		log.Fatalf("Config unmarshall err #%v", err)
	}

	return config
}
