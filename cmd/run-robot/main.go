package main

import (
	"context"
	"fmt"
	"log"

	"tinkoff-invest-bot/internal/config"
	"tinkoff-invest-bot/internal/engine"
)

func main() {
	robotConfig := config.LoadConfig("./configs/robot.yaml")

	tradingConfigs := config.LoadConfigsFromDir("./generated/")

	fmt.Println("Parsed trading configs:")
	for _, conf := range tradingConfigs {
		fmt.Printf("%v\n", conf)
	}

	ctx := context.Background()

	robotInstance := engine.New(robotConfig)
	if err := robotInstance.Run(ctx, "YNDX"); err != nil {
		log.Fatalf("InvestRobot finished with error: %v", err)
	}
}
