package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/sdcoffey/big"
	"github.com/sdcoffey/techan"
	"go.uber.org/zap"

	"tinkoff-invest-bot/internal/config"
	"tinkoff-invest-bot/internal/strategy"
	"tinkoff-invest-bot/investapi"
	"tinkoff-invest-bot/pkg/sdk"
	"tinkoff-invest-bot/pkg/utils"
)

const (
	configsPath     = "./configs/generated/"
	robotConfigPath = "./configs/robot.yaml"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Cant create production logger: %v", err)
	}
	fmt.Println(color.GreenString("🤖 Бэктестинг ассистент для торгового робота запущен!"))
	fmt.Println("Вы можете протестировать", color.MagentaString("сгенерированную стратегию 💫"))
	fmt.Println("На", color.MagentaString("исторических данных 🦕"), "доступных в Тинькофф Инвестиции")

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

	// Предложение с выбором конфига
	tradingConfigs := config.LoadTradingConfigsFromDir(configsPath)
	var tradingConfigsInfo []string
	for _, tradingConfig := range tradingConfigs {
		tradingConfigsInfo = append(tradingConfigsInfo,
			fmt.Sprintf("%s: %s_%s", tradingConfig.Strategy.Name, tradingConfig.Ticker, tradingConfig.AccountId),
		)
	}
	if len(tradingConfigs) == 0 {
		log.Fatalf("Стратегий в %s не было найдено, попробуйте сгенерировать новые", configsPath)
	}
	n := utils.RequestChoice("📈 Выберите стратегию для тестирования", tradingConfigsInfo, scanner)
	tradingConfig := tradingConfigs[n]

	vars := []string{"За последние сутки", "За последнюю неделю", "За последний месяц", "Свой промежуток (не больше месяца)"}
	vals := []time.Duration{1, 7, 30, 0}
	n = utils.RequestChoice("🕰 На каком отрезке протестировать стратегию?", vars, scanner)
	var from, to time.Time
	var candles []*investapi.HistoricCandle
	if vals[n] == 0 {
		for {
			from = utils.RequestDate("🎬 Введите дату начала в формате DDMMYY", scanner)
			to = utils.RequestDate("🎬 Введите дату конца в формате DDMMYY", scanner)
			if from.After(to) {
				color.Yellow("Дата начала позже даты конца")
			} else if to.Sub(from) > time.Hour*24*31 {
				color.Yellow("Промежуток должен быть не больше месяца")
			} else {
				break
			}
		}
	} else {
		to = time.Now()
		from = to.Add(-time.Hour * 24 * vals[n])
	}
	for from.Before(to) {
		c, _, err := s.GetCandles(
			tradingConfig.Figi,
			from,
			from.AddDate(0, 0, 1),
			sdk.IntervalToCandleInterval(tradingConfig.Strategy.Interval),
		)
		if err != nil {
			log.Fatalf("Не удается получить свечи: %v", err)
		}
		candles = append(candles, c...)
		from = from.AddDate(0, 0, 1)
	}

	strategyWrapper, err := strategy.FromConfig(tradingConfig, s, logger)
	if err != nil {
		log.Fatalf("Не удается инициализировать стратегию: %v", err)
	}

	for _, candle := range candles {
		op := strategyWrapper.Step(sdk.HistoricCandleToCandle(candle, sdk.IntervalToDuration(tradingConfig.Strategy.Interval)))
		switch op {
		case strategy.Buy:
			fallthrough
		case strategy.Sell:
			strategyWrapper.TradingRecord.Operate(techan.Order{
				Side:          techan.OrderSide(op),
				Price:         big.NewDecimal(sdk.QuotationToFloat(candle.Close)),
				Amount:        big.NewFromInt(int(tradingConfig.Strategy.Quantity)),
				ExecutionTime: candle.Time.AsTime(),
			})
		case strategy.Hold:
			continue
		default:
			panic("Значение не определено")
		}
	}

	income := 0.0
	for _, trade := range strategyWrapper.TradingRecord.Trades {
		res := trade.ExitOrder().Price.Sub(trade.EntranceOrder().Price).Float()
		fmt.Printf("res: %f\n", res)
		income += res
	}
	fmt.Println("income:", income)
}
