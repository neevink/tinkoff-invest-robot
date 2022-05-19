package strategy

import (
	"github.com/sdcoffey/techan"
	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"tinkoff-invest-bot/internal/config"
	"tinkoff-invest-bot/internal/rule-strategy"
	"tinkoff-invest-bot/pkg/sdk"
)

type Operation int

const (
	Buy Operation = iota
	Sell
	Hold
)

func FromConfig(tradingConfig *config.TradingConfig, s *sdk.SDK, logger *zap.Logger) (*Wrapper, error) {
	f := rule_strategy.List[tradingConfig.Strategy.Name]
	if f == nil {
		return nil, xerrors.Errorf("no ruleStrategy with name %s", tradingConfig.Strategy.Name)
	}

	tradingRecord := techan.NewTradingRecord() // создание структуры стратегии и истории трейдинга
	ruleStrategy, timeSeries := f(*tradingConfig)

	tradingStrategy := Wrapper{
		tradingConfig: tradingConfig,
		sdk:           s,
		logger:        logger,
		TimeSeries:    timeSeries,
		TradingRecord: tradingRecord,
		ruleStrategy:  &ruleStrategy,
	}

	return &tradingStrategy, nil
}
