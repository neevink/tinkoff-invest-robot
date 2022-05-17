package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/fatih/color"
	"log"
	"os"
	"time"
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
	n := utils.RequestChoice("📈 Выберите стратегию для тестирования", tradingConfigsInfo, scanner)
	tradingConfig := tradingConfigs[n]

	// TODO задание времени стратегии
	_, _, err = s.GetCandles(tradingConfig.Figi, time.Now(), time.Now(), api.CandleInterval_CANDLE_INTERVAL_5_MIN)
	if err != nil {
		return
	}
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
