package main

import (
	"context"
	"fmt"
	"log"

	"tinkoff-invest-bot/internal/config"
	"tinkoff-invest-bot/pkg/sdk"
)

const (
	robotConfigPath = "./configs/robot.yaml"
)

// Позволяет вручную делать запросы к API Тинькофф Инвестиций
// Делает это через SDK, тобишь позволяет приложить app-name и токен
// Как будто это делал робот
func main() {
	// Инициализация SDK
	robotConfig := config.LoadRobotConfig(robotConfigPath)
	if robotConfig.TinkoffAccessToken == "" {
		log.Fatalf("Токен доступа (TINKOFF_ACCESS_TOKEN) не был найден в .env")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s, err := sdk.New(robotConfig.TinkoffApiEndpoint, robotConfig.TinkoffAccessToken, robotConfig.AppName, ctx)
	if err != nil {
		log.Fatalf("Не удается инициализировать SDK: %v", err)
	}

	fmt.Printf("Введите accountId:\n")
	var accountId string
	_, err = fmt.Scanf("%s", &accountId)
	if err != nil {
		log.Fatalf("Не получилось считать строку: %v", err)
	}

	fmt.Printf("Введите figi:\n")
	var figi string
	_, err = fmt.Scanf("%s", &figi)
	if err != nil {
		log.Fatalf("Не получилось считать строку: %v", err)
	}

	fmt.Printf("Какой ордер выставить? (b - покупка, s - продажа):\n")
	var order string
	_, err = fmt.Scanf("%s", &order)
	if err != nil {
		log.Fatalf("Не получилось считать строку: %v", err)
	}

	if order == "b" {
		fmt.Printf("Будет выставлен ордер на покупку")
		orderId := sdk.GenerateOrderId()
		resp, trackingId, err := s.RealMarketBuy(figi, 1, accountId, orderId)
		if err != nil {
			fmt.Printf("Error %w", err)
			fmt.Printf("Tracking ID: %v\n", trackingId)
		} else {
			fmt.Printf("Responce: %v\n", resp)
			fmt.Printf("Tracking ID: %v\n", trackingId)
		}
	} else if order == "s" {
		fmt.Printf("Будет выставлен ордер на продажу")
		orderId := sdk.GenerateOrderId()
		resp, trackingId, err := s.RealMarketSell(figi, 1, accountId, orderId)
		if err != nil {
			fmt.Printf("Error %w", err)
			fmt.Printf("Tracking ID: %v\n", trackingId)
		} else {
			fmt.Printf("Responce: %v\n", resp)
			fmt.Printf("Tracking ID: %v\n", trackingId)
		}
	} else {
		log.Fatalf("Не получилось поспознать значение. Введите \"b\" или \"s\"")
	}
}
