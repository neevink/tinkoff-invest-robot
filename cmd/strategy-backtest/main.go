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
		tradingConfigsInfo = append(tradingConfigsInfo, configReport(tradingConfig))
	}
	if len(tradingConfigs) == 0 {
		log.Fatalf("Стратегий в %s не было найдено, попробуйте сгенерировать новые", configsPath)
	}
	n := utils.RequestChoice("📈 Выберите стратегию для тестирования", tradingConfigsInfo, scanner)
	tradingConfig := tradingConfigs[n]

	candles, _, err := s.GetCandles(
		tradingConfig.Figi,
		time.Now().Add(-time.Hour*24), // TODO задание времени стратегии
		time.Now(),
		sdk.IntervalToCandleInterval(tradingConfig.Strategy.Interval),
	)
	if err != nil {
		log.Fatalf("Не удается получить свечи: %v", err)
	}

	strategyWrapper, err := strategy.FromConfig(tradingConfig, s, logger)
	if err != nil {
		log.Fatalf("Не удается инициализировать стратегию: %v", err)
	}

	// TODO еще много че не закончено
	for _, candle := range candles { // будем попорядку добавлять свечи, имитируя консумер (такая штука может находиться в консумере)
		op := strategyWrapper.Step(sdk.HistoricCandleToCandle(candle, sdk.IntervalToDuration(tradingConfig.Strategy.Interval)))
		switch op {
		case strategy.Buy:
			strategyWrapper.AddTrade(techan.Order{
				Side:          techan.BUY,
				Security:      "uid",
				Price:         big.Decimal{},
				Amount:        big.Decimal{},
				ExecutionTime: candle.Time.AsTime(),
			})
		case strategy.Sell:
			strategyWrapper.AddTrade(techan.Order{
				Side:          techan.BUY,
				Security:      "uid",
				Price:         big.Decimal{},
				Amount:        big.Decimal{},
				ExecutionTime: candle.Time.AsTime(),
			})
		case strategy.Hold:
			continue
		default:
			panic("не определено")
		}
	}
	for _, trade := range strategyWrapper.GetTrades() {
		fmt.Println("buy:", trade.EntranceOrder().ExecutionTime, trade.EntranceOrder().Amount,
			"sell:", trade.ExitOrder().ExecutionTime, trade.ExitOrder().Amount)
	}
}

// Создает краткую информацию о стратегии
func configReport(tradingConfig *config.TradingConfig) string {
	// TODO pretty input %v %#v или еще что лучше
	return fmt.Sprintf("%s_%s: %s %v",
		tradingConfig.Ticker,
		tradingConfig.AccountId,
		tradingConfig.Strategy.Name,
		tradingConfig.Strategy.Other)
}

func timeSeriesFromHistoricCandles(candles []*investapi.HistoricCandle, period time.Duration) *techan.TimeSeries {
	series := techan.NewTimeSeries()

	for _, c := range candles {
		series.AddCandle(sdk.HistoricCandleToCandle(c, period))
	}
	return series
}
