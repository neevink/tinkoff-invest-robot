package main

import (
	"context"
	"fmt"
	"log"

	"tinkoff-invest-bot/internal/robot"
	"tinkoff-invest-bot/robot/pkg/engine"
)

func main() {
	config, err := robot.LoadRobotConfig("./configs/robot.yaml")
	if err != nil {
		log.Fatalf("Can't load robot configs: %v", err)
	}
	tradingConfigs := robot.LoadTradingConfigsFromDir("./generated/")

	fmt.Println("Parsed trading configs:")
	for _, conf := range tradingConfigs {
		fmt.Printf("%v\n", conf)
	}

	ctx := context.Background()

	robotInstance := engine.New(config)
	if err := robotInstance.Run(ctx, "YNDX"); err != nil {
		log.Fatalf("InvestRobot finished with error: %v", err)
	}
}
