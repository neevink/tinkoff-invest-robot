package main

import (
	"bufio"
	"fmt"
	"github.com/fatih/color"
	"log"
	"math"
	"os"
	"strconv"
	"strings"

	api "tinkoff-invest-bot/investapi"

	"tinkoff-invest-bot/internal/config"
	"tinkoff-invest-bot/pkg/sdk"
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
	// TODO работают ли емоджи на линухе?
	fmt.Println(color.GreenString("\U0001F916 Генератор конфига для торгового робота запущен!"))
	fmt.Println("Робот создан для торговли", color.MagentaString("базовыми акциями 📈"), "на MOEX и SPB")
	fmt.Println("Еще", color.MagentaString("немного текста"), "который можно в любой момент изменить 💫")

	// Инициализация SDK
	robotConfig := config.LoadRobotConfig(robotConfigPath)
	if robotConfig.TinkoffAccessToken == "" {
		log.Fatalf("Токен доступа (TINKOFF_ACCESS_TOKEN) не был найден в .env")
	}

	s, err := sdk.New(robotConfig.TinkoffApiEndpoint, robotConfig.TinkoffAccessToken)
	if err != nil {
		log.Fatalf("Не удается инициализировать SDK: %v", err)
	}

	// Формирование информации об аккаунтах
	accounts, err := s.GetAccounts()
	if err != nil {
		log.Fatalf("Не удается получить информацию об аккаунтах: %v", err)
	}
	invalidAccounts := 0
	var accountsInfo []string
	for _, account := range accounts {
		// Фильтрация аккаунтов на валидные и нет
		if account.GetType() == api.AccountType_ACCOUNT_TYPE_UNSPECIFIED ||
			account.GetStatus() != api.AccountStatus_ACCOUNT_STATUS_OPEN ||
			account.GetAccessLevel() != api.AccessLevel_ACCOUNT_ACCESS_LEVEL_FULL_ACCESS {
			invalidAccounts++
			continue
		}
		// Получение краткой информации об аккаунте
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
		portfolio, err := s.GetPortfolio(account.GetId())
		if err != nil {
			log.Fatalf("Не удается получить портфолио аккаунта %s: %v", account.GetId(), err)
		}
		accountInfo += portfolioReport(portfolio)
		accountsInfo = append(accountsInfo, accountInfo)
	}

	// Выбор аккаунта для торговли
	if invalidAccounts > 0 {
		fmt.Printf(color.YellowString("Найдено аккаунтов без возможности торговли")+": %d\n", invalidAccounts)
	}
	n := requestChoice("👤 Выберите аккаунт для торговли", accountsInfo)
	account := accounts[n]

	// Конфигурация стратегии
	// TODO выбор и задание параметров стратегии (будет использоваться StrategyList)
	strategy := config.Strategy{
		Name: "",
		StrategyConfig: config.StrategyConfig{
			Threshold:    0,
			CandlesCount: 0,
		},
	}

	// Выбор акций для торговли
	responseShares, err := s.GetShares()
	if err != nil {
		log.Fatalf("Не удается получить информацию об акциях: %v", err)
	}
	input := requestParameter("🛍 Введите тикеры акций для торговли", true)
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
						Strategy:  strategy,
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
			ticker = strings.ToUpper(requestParameter("🖍 Уточните или пропустите тикер", false))
			if ticker == "" {
				continue TickerLoop
			}
		}
	}

	fmt.Println(color.GreenString("👍 Удачной торговли!"))
}

// Запросить у пользователя параметр в виде строки
func requestParameter(msg string, required bool) string {
	for {
		fmt.Printf(color.BlueString(msg) + ": ")
		if !scanner.Scan() {
			if scanner.Err() == nil {
				log.Fatalf("Ввод из консоли принудительно завершен")
			} else {
				fmt.Println(color.YellowString("Не удалось прочитать из консоли"))
				continue
			}
		}
		parameter := scanner.Text()
		if required && parameter == "" {
			fmt.Println(color.YellowString("Этот параметр является обязательным"))
			continue
		}
		return parameter
	}
}

// Запросить у пользователя выбор строки из предложенных строк
func requestChoice(msg string, a []string) int {
	if len(a) <= 0 {
		log.Fatalf("Ошибка, передано 0 возможных значений")
	}
	for i, aa := range a {
		fmt.Printf("%d. %s\n", i, aa)
	}
	for {
		input := requestParameter(msg, true)
		n, err := strconv.Atoi(input)
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
