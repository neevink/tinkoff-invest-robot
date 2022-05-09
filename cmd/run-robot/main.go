package main

import (
	"fmt"
	"log"

	"tinkoff-invest-bot/internal/config"
	"tinkoff-invest-bot/internal/engine"
)

func main() {
	robotConfig := config.LoadConfig("./configs/robot.yaml")

	tradingConfigs := config.LoadTradingConfigsFromDir("./configs/generated/")

	fmt.Println("Parsed trading configs:")
	for _, conf := range tradingConfigs {
		fmt.Printf("%v\n", conf)
	}

	robotInstance, err := engine.New(robotConfig, tradingConfigs[0])
	if err != nil {
		log.Fatalf("Cant create robot instance: %v", err)
	}
	if err := robotInstance.Run(); err != nil {
		log.Fatalf("InvestRobot finished with error: %v", err)
	}
}
