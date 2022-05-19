package rule_strategy

import (
	"github.com/sdcoffey/techan"

	"tinkoff-invest-bot/internal/config"
)

const (
	shortWindow  = "short_window"
	middleWindow = "middle_window"
	longWindow   = "long_window"
	window       = "window"
)

// TODO можно сделать RSI, волны боллинджера, CCI, Stochastic, Keltner Channel, MACD
var (
	// List единственное место, где задаются стратегии
	List = map[string]func(tradingConfig config.TradingConfig) (techan.RuleStrategy, *techan.TimeSeries){
		"simpleEMA":   simpleEMA,
		"doubleEMA":   doubleEMA,
		"tripleEMA":   tripleEMA,
		"simpleAroon": simpleAroon,
	}
	RequiredParameters = map[string][]string{
		"simpleEMA":   {window},
		"doubleEMA":   {shortWindow, longWindow},
		"tripleEMA":   {shortWindow, middleWindow, longWindow},
		"simpleAroon": {window},
	}
)
