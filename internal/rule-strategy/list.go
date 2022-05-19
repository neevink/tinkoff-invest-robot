package rule_strategy

import (
	"github.com/sdcoffey/techan"

	"tinkoff-invest-bot/internal/config"
)

var (
	// List единственное место, где задаются стратегии
	List = map[string]func(tradingConfig config.TradingConfig) (techan.RuleStrategy, *techan.TimeSeries){
		"simpleEma": SimpleEma,
	}
	RequiredParameters = map[string][]string{
		"simpleEma": {"window"},
	}
)
