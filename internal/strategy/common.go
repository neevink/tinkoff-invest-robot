package strategy

import (
	"go.uber.org/zap"
	"golang.org/x/xerrors"
	rule_trategy "tinkoff-invest-bot/internal/rule-strategy"

	"tinkoff-invest-bot/internal/config"
	investsdk "tinkoff-invest-bot/pkg/sdk"
)

type Operation int

const (
	Buy Operation = iota
	Sell
	Hold
)

func FromConfig(tradingConfig *config.TradingConfig, s *investsdk.SDK, logger *zap.Logger) (*Wrapper, error) {
	f := rule_trategy.List[tradingConfig.Strategy.Name]
	if f == nil {
		return nil, xerrors.Errorf("no ruleStrategy with name %s", tradingConfig.Strategy.Name)
	}

	ruleStrategy, timeSeries, tradingRecord := f(*tradingConfig)

	tradingStrategy := Wrapper{
		tradingConfig: tradingConfig,
		sdk:           s,
		logger:        logger,
		timeSeries:    timeSeries,
		tradingRecord: tradingRecord,
		ruleStrategy:  &ruleStrategy,
	}

	return &tradingStrategy, nil
}
