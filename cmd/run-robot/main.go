package main

import (
	"context"
	"log"

	"tinkoff-invest-bot/internal/robot"
	"tinkoff-invest-bot/robot/pkg/engine"
)

func main() {
	config := robot.LoadConfig("./configs/main.yaml")

	ctx := context.Background()

	robotInstance := engine.New(config)
	if err := robotInstance.Run(ctx, "YNDX"); err != nil {
		log.Fatalf("InvestRobot finished with error: %v", err)
	}
}
