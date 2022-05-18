package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"strings"

	"github.com/fatih/color"

	"tinkoff-invest-bot/internal/config"
	"tinkoff-invest-bot/internal/rule-strategy"
	"tinkoff-invest-bot/investapi"
	"tinkoff-invest-bot/pkg/sdk"
	"tinkoff-invest-bot/pkg/utils"
)

var (
	scanner = bufio.NewScanner(os.Stdin)
	bold    = color.New(color.Bold).SprintfFunc()
)

const (
	configsPath     = "./configs/generated/"
	robotConfigPath = "./configs/robot.yaml"
)

func main() {
	fmt.Println(color.GreenString("🤖 Генератор конфига для торгового робота запущен!"))
	fmt.Println("Робот создан для торговли", color.MagentaString("базовыми акциями 📈"), "в Тинькофф Инвестиции")
	fmt.Println("Еще", color.MagentaString("немного текста"), "который можно в любой момент изменить 💫")

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

	// Формирование информации об аккаунтах
	accounts, _, err := s.GetAccounts()
	if err != nil {
		log.Fatalf("Не удается получить информацию об аккаунтах: %v", err)
	}
	invalidAccounts := 0
	var validAccounts []*investapi.Account
	var accountsInfo []string
	for _, account := range accounts {
		// Фильтрация аккаунтов на валидные и нет
		if account.GetType() == investapi.AccountType_ACCOUNT_TYPE_UNSPECIFIED ||
			account.GetStatus() != investapi.AccountStatus_ACCOUNT_STATUS_OPEN ||
			account.GetAccessLevel() != investapi.AccessLevel_ACCOUNT_ACCESS_LEVEL_FULL_ACCESS {
			invalidAccounts++
			continue
		}
		// Получение краткой информации об аккаунте
		var accountInfo string
		switch account.GetType() {
		case investapi.AccountType_ACCOUNT_TYPE_INVEST_BOX:
			accountInfo += "🐷 "
		case investapi.AccountType_ACCOUNT_TYPE_TINKOFF_IIS:
			accountInfo += "🏦 "
		case investapi.AccountType_ACCOUNT_TYPE_TINKOFF:
			accountInfo += "💰 "
		}
		accountInfo += account.GetName() + " "
		portfolio, _, err := s.GetPortfolio(account.GetId())
		if err != nil {
			log.Fatalf("Не удается получить портфолио аккаунта %s: %v", account.GetId(), err)
		}
		accountInfo += portfolioReport(portfolio)
		accountsInfo = append(accountsInfo, accountInfo)
		validAccounts = append(validAccounts, account)
	}

	// Выбор аккаунта для торговли
	if invalidAccounts > 0 {
		fmt.Printf(color.YellowString("Найдено аккаунтов без доступа к торговле")+": %d\n", invalidAccounts)
	}
	if invalidAccounts >= len(accounts) {
		log.Fatalln("По данному токену не найдено аккаунтов с доступом к торговле")
	}
	n := utils.RequestChoice("👤 Выберите аккаунт для торговли", accountsInfo, scanner)
	account := validAccounts[n]

	// Конфигурация стратегии
	var ruleStrategyNames []string
	for name := range rule_strategy.List {
		ruleStrategyNames = append(ruleStrategyNames, name)
	}
	n = utils.RequestChoice("🕹 Выберите стратегию из предложенных", ruleStrategyNames, scanner)
	ruleStrategy := ruleStrategyNames[n]
	// TODO задание параметров стратегии ("window"), интервала, количества лотов
	strategyConfig := config.StrategyConfig{
		Name:     ruleStrategy,
		Interval: "1_MIN",
		Other:    make(map[string]int, 0),
		Quantity: 1,
	}

	// Выбор акций для торговли
	responseShares, _, err := s.GetShares()
	if err != nil {
		log.Fatalf("Не удается получить информацию об акциях: %v", err)
	}
	input := utils.RequestParameter("🛍 Введите тикеры акций для торговли (через пробел)", true, scanner)
	tickers := strings.Split(input, " ")
	for i := 0; i < len(tickers); i++ {
		tickers[i] = strings.ToUpper(tickers[i])
	}

TickerLoop:
	// Создание конфигурации для каждой акции
	for _, ticker := range tickers {
		for {
			for _, share := range responseShares {
				if share.GetTicker() == ticker {
					tradingConfig := config.TradingConfig{
						AccountId: account.GetId(),
						Ticker:    ticker,
						Figi:      share.GetFigi(),
						Strategy:  strategyConfig,
						Exchange:  share.GetExchange(),
					}
					filename := ticker + "_" + account.GetId() + ".yaml"
					err := config.WriteTradingConfig(configsPath, filename, &tradingConfig)
					if err != nil {
						fmt.Println(color.YellowString("Торговая конфигурация %s не была записана %v", filename, err))
					}
					continue TickerLoop
				}
			}
			fmt.Println(color.YellowString("Инструмент с тикером \"" + ticker + "\" не найден!"))
			ticker = strings.ToUpper(utils.RequestParameter("🖍 Уточните или пропустите тикер", false, scanner))
			if ticker == "" {
				continue TickerLoop
			}
		}
	}

	fmt.Println(color.GreenString("👍 Удачной торговли!"))
}

func portfolioReport(portfolio *investapi.PortfolioResponse) string {
	totalAmount := sdk.MoneyValueToFloat(portfolio.GetTotalAmountCurrencies()) +
		sdk.MoneyValueToFloat(portfolio.GetTotalAmountBonds()) +
		sdk.MoneyValueToFloat(portfolio.GetTotalAmountShares()) +
		sdk.MoneyValueToFloat(portfolio.GetTotalAmountEtf()) +
		sdk.MoneyValueToFloat(portfolio.GetTotalAmountFutures())

	expectedYield := sdk.QuotationToFloat(portfolio.ExpectedYield)

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
