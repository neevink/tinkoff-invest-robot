package main

import (
	"fmt"
	"log"

	api "tinkoff-invest-bot/investapi"

	"tinkoff-invest-bot/internal/robot"
	"tinkoff-invest-bot/pkg/sdk"
)

func main() {
	log.Print("RobotConfig generator is running")
	config, err := robot.LoadRobotConfig("./configs/robot.yaml")
	if err != nil {
		log.Fatalf("Can't load robot configs: %v", err)
	}

	s, err := sdk.New(config.TinkoffApiEndpoint, config.AccessToken)
	if err != nil {
		log.Fatalf("Can't init sdk: %v", err)
	}

	fmt.Printf("Input new config name without spaces:\n")
	var fileName string
	if _, err := fmt.Scan(&fileName); err != nil {
		log.Fatalf("Scan for accountIndex failed, due to %e", err)
	}
	newConfigsPath := "./generated/" + fileName + ".yaml"

	userInfo, err := s.GetUserInfo()
	if err != nil {
		log.Fatalf("Can't receive user info: %v", err)
	}
	fmt.Printf("User info: %v\n", userInfo)

	log.Println("Select trading account:")
	accounts, err := s.GetAccounts()
	if err != nil {
		log.Fatalf("Can't receive accounts: %v", err)
	}
	for i, account := range accounts {
		fmt.Printf("%d. %s (status: %s, account id: %s)\n", i, account.GetName(), account.GetStatus(), account.GetId())
	}
	var accountIndex int
	if _, err := fmt.Scan(&accountIndex); err != nil {
		log.Fatalf("Scan for accountIndex failed, due to %e", err)
	}
	account := accounts[accountIndex]
	fmt.Printf("Selected account with id=%s\n", account.GetId())

	portf, err := s.GetPortfolio(account.GetId())
	if err != nil {
		log.Fatalf("Can't receive portfolio info: %v", err)
	}
	printMoney("All shares in account costs", portf.GetTotalAmountShares())

	marginAttrs, err := s.GetMarginAttributes(account.GetId())
	if err != nil {
		log.Fatalf("Can't receive margin attributes info: %v", err)
	}
	printMoney("liquid_portfolio", marginAttrs.GetLiquidPortfolio())
	printMoney("starting_margin", marginAttrs.GetStartingMargin())
	printMoney("minimal_margin", marginAttrs.GetMinimalMargin())
	printMoney("amount_of_missing_funds", marginAttrs.GetAmountOfMissingFunds())

	fmt.Printf("Input figi of share (example: BBG00RZ9HFD6):\n")
	var figi string
	if _, err := fmt.Scan(&figi); err != nil {
		log.Fatalf("Scan for figi failed: %e", err)
	}

	share, err := s.GetInstrumentByFigi(figi)
	if err != nil {
		log.Fatalf("Can't receive share: %v", err)
	}
	fmt.Printf("Share name: %s, currency: %s, instrument: %s\n", share.GetName(), share.GetCurrency(), share.GetInstrumentType())

	newConfig := &robot.TradingConfig{
		AccountId:       account.GetId(),
		Figi:            figi,
		TradingStrategy: "simple",
	}
	if err := robot.WriteTradingConfig(newConfigsPath, newConfig); err != nil {
		log.Fatalf("Saving error %e", err)
	}
	fmt.Printf("New trading config added successfully!")
}

func printMoney(mes string, moneyValue *api.MoneyValue) {
	fmt.Printf("%s: %d.%d%s\n", mes, moneyValue.GetUnits(), moneyValue.GetNano(), moneyValue.GetCurrency())
}
