package strategies

import (
	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"tinkoff-invest-bot/internal/config"
	investsdk "tinkoff-invest-bot/pkg/sdk"
)

var (
	StrategyList = map[string]func(tradingConf *config.TradingConfig, s *investsdk.SDK, logger *zap.Logger) *TradingStrategy{
		"simple": NewSimpleStrategy,
	}
)

func FromConfig(conf *config.TradingConfig, s *investsdk.SDK, logger *zap.Logger) (*TradingStrategy, error) {
	f := StrategyList[conf.Strategy.Name]
	if f == nil {
		return nil, xerrors.Errorf("no strategy with name %s", conf.Strategy.Name)
	}
	return f(conf, s, logger), nil
}
