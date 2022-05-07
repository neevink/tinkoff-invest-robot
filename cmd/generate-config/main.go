package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"

	api "tinkoff-invest-bot/investapi"

	"tinkoff-invest-bot/internal/config"
	"tinkoff-invest-bot/pkg/sdk"
)

var (
	scanner          = bufio.NewScanner(os.Stdin)
	bold             = color.New(color.Bold).SprintfFunc()
	configsPath      = "./configs/generated/"
	commonConfigPath = "./configs/common.yaml"
)

func main() {
	fmt.Println(color.GreenString("🤖 Генератор конфига для торгового робота запущен!"))
	fmt.Println("Робот создан для торговли", color.MagentaString("базовыми акциями 📈"), "на MOEX и SPB")
	fmt.Println("Еще немного текста который можно в любой момент изменить 💫")
	commonConfig := config.LoadConfig(commonConfigPath)
	tinkoffApiEndpoint := requestParameter("📬 Адрес сервиса", commonConfig.TinkoffApiEndpoint)
	accessToken := requestParameter("🔑 Токен доступа", commonConfig.AccessToken)

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
		fmt.Printf(color.YellowString("Найдено невалидных аккаунтов для торговли")+": %d\n", invalidAccounts)
	}
	n := requestChoice("👤 Выберите аккаунт для торговли", accountsInfo)
	account := accounts[n]

	responseShares, err := s.GetShares()
	if err != nil {
		log.Fatalf("Не удается получить информацию об акциях: %v", err)
	}
	var commonTickers []string
	for _, ticker := range commonConfig.Shares {
		commonTickers = append(commonTickers, ticker.Ticker)
	}
	input := requestParameter("🛍 Введите тикеры акций для торговли", strings.Trim(fmt.Sprint(commonTickers), "[]"))
	var tickers []string
	if input == "" {
		tickers = commonTickers
	} else {
		tickers = strings.Split(input, " ")
	}
	for i := 0; i < len(tickers); i++ {
		tickers[i] = strings.ToUpper(tickers[i])
	}
	var shares []config.Share
TickerLoop:
	for _, ticker := range tickers {
		for _, share := range responseShares {
			if share.GetTicker() == ticker {
				shares = append(shares, config.Share{Ticker: ticker, Figi: share.GetFigi()})
				continue TickerLoop
			}
		}
		fmt.Println(color.YellowString("Инструмент с тикером \"" + ticker + "\" не найден!"))
	}

	newConfig := &config.Config{
		TinkoffApiEndpoint: tinkoffApiEndpoint,
		AccessToken:        accessToken,
		AccountId:          account.GetId(),
		Shares:             shares,
	}

	fileName := requestParameter("📄 Название нового конфига", "config_at_"+time.Now().Format("02-01-06_15:04.05"))
	if strings.Contains(fileName, "/") {
		log.Fatalf("Использование подпапок недопустимо")
	}
	newConfigPath := configsPath + fileName + ".yaml"
	if err := config.WriteConfig(newConfigPath, newConfig); err != nil {
		log.Fatalf("Ошибка сохранения конфига %v", err)
	}
	fmt.Println(color.GreenString("👍 Конфиг успешно сохранен, удачной торговли!"))
}

func requestParameter(msg string, common string) string {
	fmt.Printf(color.BlueString(msg)+": (%s) ", common)
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
	if len(a) <= 0 {
		log.Fatalf("Ошибка, передано 0 возможных значений")
	}
	for {
		for i, aa := range a {
			fmt.Printf("%d. %s\n", i, aa)
		}
		input := requestParameter(msg, "0")
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
