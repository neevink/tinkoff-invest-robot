package rule_strategy

import (
	"github.com/sdcoffey/techan"

	"tinkoff-invest-bot/internal/config"
)

func SimpleEma(tradingConfig config.TradingConfig) (techan.RuleStrategy, *techan.TimeSeries, *techan.TradingRecord) {
	var window = tradingConfig.Strategy.Other["window"]          // TODO как обработать отсутствие значения
	series := techan.NewTimeSeries()                             // история всех свечей
	closePrices := techan.NewClosePriceIndicator(series)         // отсеивает High, Low, Open, на выходе только Close
	movingAverage := techan.NewEMAIndicator(closePrices, window) // Создает экспоненциальное средне с окном в n свечей

	// создание структуры стратегии и истории трейдинга
	tradingRecord := techan.NewTradingRecord()
	entryRule := techan.And( // правило входа
		techan.NewCrossUpIndicatorRule(movingAverage, closePrices), // когда свеча закрытия пересечет EMA (станет выше EMA)
		techan.PositionNewRule{})                                   // и сделок не открыто — мы покупаем
	exitRule := techan.And( // правило выхода
		techan.NewCrossDownIndicatorRule(closePrices, movingAverage), // когда свеча закроется ниже EMA
		techan.PositionOpenRule{})                                    // и сделка открыта — продаем
	ruleStrategy := techan.RuleStrategy{
		UnstablePeriod: window,
		EntryRule:      entryRule,
		ExitRule:       exitRule,
	}
	return ruleStrategy, series, tradingRecord
}
