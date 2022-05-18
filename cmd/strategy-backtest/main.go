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

	candles, _, err := s.GetCandles(
		tradingConfig.Figi,
		time.Now().Add(-time.Hour*24), // TODO задание времени стратегии
		time.Now(),
		sdk.IntervalToCandleInterval(tradingConfig.Strategy.Interval),
	)
	if err != nil {
		log.Fatalf("Не удается получить свечи: %v", err)
	}

	//ok
	//ok
	//ok
	//ok
	//ok
	// стратегия называется (cross over EMA — buy, cross below — sell)
	const window = 100                                                  // значение будет подгружено из конфига (окно индикатора EMA)
	duration := sdk.IntervalToDuration(tradingConfig.Strategy.Interval) // интервал свечи, в формате Duration
	series := techan.NewTimeSeries()                                    // история всех свечей (будет использоваться также для графиков)
	closePrices := techan.NewClosePriceIndicator(series)                // отсеивает High, Low, Open, на выходе только Close
	movingAverage := techan.NewEMAIndicator(closePrices, window)        // Создает экспоненциальное средне с окном в n свечей

	// так выглядит создание стратегии
	record := techan.NewTradingRecord() // запись покупок, продаж (будет использоваться также для графиков)
	entryRule := techan.And(            // правило вхождения
		techan.NewCrossUpIndicatorRule(movingAverage, closePrices), // когда свеча закрытия пересечет EMA (станет выше EMA)
		techan.PositionNewRule{})                                   // и сделок не открыто — мы покупаем
	exitRule := techan.And(
		techan.NewCrossDownIndicatorRule(closePrices, movingAverage), // тут соответственно наоборот, стратегия ужасно работает на рынке без тренда
		techan.PositionOpenRule{})
	strategy := techan.RuleStrategy{
		UnstablePeriod: window, // нестабильный период, сюда нужно класть размер окна, так как EMA не будет рассчитываться
		EntryRule:      entryRule,
		ExitRule:       exitRule,
	}

	for i, candle := range candles { // будем попорядку добавлять свечи, имитируя консумер (такая штука может находиться в консумере)
		series.AddCandle(candleFromHistoricCandle(candle, duration)) // мол добавляем пришедшую свечу (неважно откуда)
		if strategy.ShouldEnter(series.LastIndex(), record) {
			// тут соответственно возвращаем сигнал о покупке, бэктестинг это может сохранить, а sdk будет покупать по этому сигналу

			record.Operate(techan.Order{
				Side:          techan.BUY,
				Security:      "uid",
				Price:         big.Decimal{},
				Amount:        big.Decimal{},
				ExecutionTime: series.LastCandle().Period.Start,
			})
			fmt.Println(i, candle.Time.AsTime(), ":", sdk.QuotationToFloat(candle.Close), movingAverage.Calculate(series.LastIndex())) // выведем по приколу сейчасшнюю цену и результат индикатора
			fmt.Println("КУПИЛИ")
		} else {
			if strategy.ShouldExit(series.LastIndex(), record) {
				// тут соответственно возвращаем сигнал о продаже, бэктестинг это может сохранить, а sdk будет продавать по этому сигналу

				record.Operate(techan.Order{
					Side:          techan.SELL,
					Security:      "uid",
					Price:         big.Decimal{},
					Amount:        big.Decimal{},
					ExecutionTime: series.LastCandle().Period.Start, //допустим
				})
				fmt.Println(i, candle.Time.AsTime(), ":", sdk.QuotationToFloat(candle.Close), movingAverage.Calculate(series.LastIndex())) // выведем по приколу сейчасшнюю цену и результат индикатора
				fmt.Println("продали")
			} else {
				//fmt.Println("Действий не требуется")
			}
		}
	}

	fmt.Println("вот трейды:")
	for i, trade := range record.Trades {
		fmt.Println(i, trade.EntranceOrder().ExecutionTime, trade.ExitOrder().ExecutionTime)
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

func timeSeriesFromHistoricCandles(candles []*api.HistoricCandle, period time.Duration) *techan.TimeSeries {
	series := techan.NewTimeSeries()

	for _, c := range candles {
		series.AddCandle(candleFromHistoricCandle(c, period))
	}
	return series
}

func candleFromHistoricCandle(c *api.HistoricCandle, period time.Duration) *techan.Candle {
	timePeriod := techan.NewTimePeriod(c.Time.AsTime(), period)
	candle := techan.NewCandle(timePeriod)

	candle.OpenPrice = big.NewDecimal(sdk.QuotationToFloat(c.Open))
	candle.ClosePrice = big.NewDecimal(sdk.QuotationToFloat(c.Close))
	candle.MaxPrice = big.NewDecimal(sdk.QuotationToFloat(c.High))
	candle.MinPrice = big.NewDecimal(sdk.QuotationToFloat(c.Low))
	candle.Volume = big.NewFromInt(int(c.Volume))
	return candle
}
