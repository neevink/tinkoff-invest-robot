package main

import (
	"context"
	"fmt"
	"log"

	"go.uber.org/zap"

	"tinkoff-invest-bot/internal/config"
	"tinkoff-invest-bot/internal/engine"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Cant create production logger: %v", err)
	}

	robotConfig := config.LoadRobotConfig("./configs/robot.yaml")

	tradingConfigs := config.LoadTradingConfigsFromDir("./configs/generated/")

	fmt.Println("Parsed trading configs:")
	for _, conf := range tradingConfigs {
		fmt.Printf("%v\n", conf)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	robotInstance, err := engine.New(robotConfig, tradingConfigs[0], logger, ctx)
	if err != nil {
		logger.Fatal("Cant create robot instance", zap.Error(err))
	}
	if err := robotInstance.Run(); err != nil {
		logger.Fatal("Cant create robot instance", zap.Error(err))
	}
}
