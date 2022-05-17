package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/fatih/color"
	"tinkoff-invest-bot/internal/config"
	api "tinkoff-invest-bot/investapi"
	"tinkoff-invest-bot/pkg/sdk"
	"tinkoff-invest-bot/pkg/utils"
)

const (
	configsPath     = "./configs/generated/"
	robotConfigPath = "./configs/robot.yaml"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
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

	// TODO задание времени стратегии
	candles, _, err := s.GetCandles(
		tradingConfig.Figi,
		time.Now().Add(-time.Hour*24),
		time.Now(),
		sdk.ConvertIntervalToCandleInterval(tradingConfig.Interval),
	)
	if err != nil {
		log.Fatalf("Не удается получить свечи: %v", err)
	}

	fmt.Println(convertCandles(candles))
}

// Создает краткую информацию о стратегии
func configReport(tradingConfig *config.TradingConfig) string {
	// TODO pretty input %v %#v или еще что лучше
	return fmt.Sprintf("%s_%s: %s %v",
		tradingConfig.Ticker,
		tradingConfig.AccountId,
		tradingConfig.Strategy.Name,
		tradingConfig.Strategy.Config)
}

// Формат свечи: Timestamp, Open, Close, High, Low, volume
func convertCandles(candles []*api.HistoricCandle) [][]float64 {
	var convertedCandles [][]float64
	for _, candle := range candles {
		convertedCandles = append(convertedCandles, []float64{
			float64(candle.Time.Seconds),
			sdk.ConvertQuotation(candle.Open),
			sdk.ConvertQuotation(candle.Close),
			sdk.ConvertQuotation(candle.High),
			sdk.ConvertQuotation(candle.Low),
			float64(candle.Volume),
		})
	}
	return convertedCandles
}
