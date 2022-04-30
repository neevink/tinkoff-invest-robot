package main

import (
	"fmt"
	"log"

	api "tinkoff-invest-bot/investapi"

	"tinkoff-invest-bot/internal/robot"
	"tinkoff-invest-bot/pkg/sdk"
)

func main() {
	log.Print("Config generator is running")

	config := robot.LoadConfig("./configs/main.yaml")

	// ctx := context.Background()

	s, err := sdk.New(config.TinkoffApiEndpoint, config.AccessToken)
	if err != nil {
		log.Fatalf("Can't init sdk: %v", err)
	}

	userInfo, err := s.GetUserInfo()
	if err != nil {
		log.Fatalf("Can't receive user info: %v", err)
	}
	fmt.Printf("User info: %v\n", userInfo)

	log.Println("Select trading account:")
	accounts, err := s.GetAccounts()
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

}

func printMoney(mes string, moneyValue *api.MoneyValue) {
	fmt.Printf("%s: %d.%d%s\n", mes, moneyValue.GetUnits(), moneyValue.GetNano(), moneyValue.GetCurrency())
}
