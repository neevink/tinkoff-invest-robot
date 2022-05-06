package main

import (
	"bufio"
	"fmt"
	"github.com/fatih/color"
	"log"
	"math"
	"os"
	"strconv"
	api "tinkoff-invest-bot/investapi"

	"tinkoff-invest-bot/internal/config"
	"tinkoff-invest-bot/pkg/sdk"
)

var (
	scanner = bufio.NewScanner(os.Stdin)
	bold    = color.New(color.Bold).SprintfFunc()
)

func main() {
	fmt.Println(color.GreenString("🤖 Генератор конфига для торгового робота запущен!"))
	fmt.Println("Следуйте инструкциям 📝", color.MagentaString("блабла"), "можете использовать значения по умолчанию", color.MagentaString("бла"))
	fmt.Println("Еще немного текста который можно в любой момент изменить 💫")

	commonConfig := config.LoadConfig("./configs/common.yaml")

	tinkoffApiEndpoint := requestParameter(color.BlueString("📬 Адрес сервиса"), commonConfig.TinkoffApiEndpoint)
	accessToken := requestParameter(color.BlueString("🔑 Токен доступа"), commonConfig.AccessToken)

	s, err := sdk.New(tinkoffApiEndpoint, accessToken)
	if err != nil {
		log.Fatalf("Не удается инициализировать SDK: %v", err)
	}

	accounts, err := s.GetAccounts()
	if err != nil {
		log.Fatalf("Не удается получить информацию об аккаунтах: %v", err)
	}
	accountsAndPortfolios := make(map[*api.Account]*api.PortfolioResponse)
	for _, account := range accounts {
		portfolio, err := s.GetPortfolio(account.GetId())
		if err != nil {
			log.Fatalf("Не удается получить портфолио аккаунта %s: %v", account.GetId(), err)
		}
		accountsAndPortfolios[account] = portfolio
	}
	accountsInfo, invalidAccounts := accountsReport(accountsAndPortfolios)
	if invalidAccounts > 0 {
		fmt.Printf(color.YellowString("😵 Найдено невалидных аккаунтов для торговли: %d\n", invalidAccounts))
	}
	n := requestChoice(color.BlueString("👤 Выберите аккаунт для торговли"), accountsInfo)
	account := accounts[n]

	fmt.Printf("Selected account with id=%s\n", account.GetId())

	fmt.Printf("Input figi of share (example: BBG00RZ9HFD6):\n")
	var figi string
	if _, err := fmt.Scanln(&figi); err != nil {
		log.Fatalf("Scan for figi failed: %e", err)
	}

	share, err := s.GetInstrumentByFigi(figi)
	if err != nil {
		log.Fatalf("Can't receive share: %v", err)
	}
	fmt.Printf("Share name: %s, currency: %s, instrument: %s\n", share.GetName(), share.GetCurrency(), share.GetInstrumentType())

	newConfig := &config.Config{
		TinkoffApiEndpoint: tinkoffApiEndpoint,
		AccessToken:        accessToken,
		AccountId:          account.GetId(),
		Figi:               figi,
	}

	fmt.Printf("Input new config name without spaces:\n")
	var fileName string
	if _, err := fmt.Scanln(&fileName); err != nil {
		log.Fatalf("Scan for accountIndex failed, due to %e", err)
	}
	newConfigsPath := "./generated/" + fileName + ".yaml"

	if err := config.WriteConfig(newConfigsPath, newConfig); err != nil {
		log.Fatalf("Saving error %v", err)
	}
	fmt.Printf("New trading config added successfully!")
}

func requestParameter(msg string, common string) string {
	fmt.Printf("%s: (%s) ", msg, common)
	for {
		if !scanner.Scan() {
			if scanner.Err() == nil {
				log.Fatalf("Ввод из консоли принудительно завершен")
			} else {
				fmt.Println(color.YellowString("Не удалось прочитать из консоли"))
				continue
			}
		}
		parameter := scanner.Text()
		if parameter == "" {
			return common
		} else {
			return parameter
		}
	}
}

func accountsReport(accountsAndPortfolios map[*api.Account]*api.PortfolioResponse) ([]string, int) {
	var accountsInfo []string
	var invalidAccounts int
	i := 0
	for account, portfolio := range accountsAndPortfolios {
		if account.GetType() == api.AccountType_ACCOUNT_TYPE_UNSPECIFIED ||
			account.GetStatus() != api.AccountStatus_ACCOUNT_STATUS_OPEN ||
			account.GetAccessLevel() != api.AccessLevel_ACCOUNT_ACCESS_LEVEL_FULL_ACCESS {
			invalidAccounts++
		} else {
			var accountInfo string
			switch account.GetType() {
			case api.AccountType_ACCOUNT_TYPE_INVEST_BOX:
				accountInfo += "🐷 "
			case api.AccountType_ACCOUNT_TYPE_TINKOFF_IIS:
				accountInfo += "🏦 "
			case api.AccountType_ACCOUNT_TYPE_TINKOFF:
				accountInfo += "💰 "
			}
			accountInfo += account.GetName() + " "
			accountInfo += portfolioReport(portfolio)
			accountsInfo = append(accountsInfo, accountInfo)
		}
		i++
	}
	return accountsInfo, invalidAccounts
}

func portfolioReport(portfolio *api.PortfolioResponse) string {
	totalAmount := convertMoneyValue(portfolio.GetTotalAmountCurrencies()) +
		convertMoneyValue(portfolio.GetTotalAmountBonds()) +
		convertMoneyValue(portfolio.GetTotalAmountShares()) +
		convertMoneyValue(portfolio.GetTotalAmountEtf()) +
		convertMoneyValue(portfolio.GetTotalAmountFutures())

	expectedYield := float64(portfolio.ExpectedYield.Units) + float64(portfolio.ExpectedYield.Nano)/1000000000

	report := bold("%.2f₽ ", totalAmount)
	income := fmt.Sprintf("%.2f₽ (%.2f%%)", totalAmount*expectedYield/100, math.Abs(expectedYield))
	switch {
	case expectedYield < 0:
		report += color.RedString(income)
	case expectedYield > 0:
		report += color.GreenString(income)
	default:
		report += color.WhiteString(income)
	}
	return report
}

func convertMoneyValue(moneyValue *api.MoneyValue) float64 {
	return float64(moneyValue.Units) + float64(moneyValue.Nano)/1000000000
}

func requestChoice(msg string, a []string) int {
	fmt.Printf("%s:\n", msg)
	if len(a) <= 0 {
		log.Fatalf("Ошибка, передано 0 возможных значений")
	}
	for i, aa := range a {
		fmt.Printf("%d. %s\n", i, aa)
	}
	for {
		if !scanner.Scan() {
			if scanner.Err() == nil {
				log.Fatalf("Ввод из консоли принудительно завершен")
			} else {
				fmt.Println(color.YellowString("Не удалось прочитать из консоли"))
				continue
			}
		}
		n, err := strconv.Atoi(scanner.Text())
		if err != nil {
			fmt.Println(color.YellowString("Ошибка конвертации в целое число"))
			continue
		}
		if n < 0 || n >= len(a) {
			fmt.Println(color.YellowString("Введите число в промежутке [%d, %d]", 0, len(a)-1))
			continue
		}
		return n
	}
}
