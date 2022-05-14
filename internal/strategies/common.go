package strategies

import (
	"golang.org/x/xerrors"

	"tinkoff-invest-bot/internal/config"
	investsdk "tinkoff-invest-bot/pkg/sdk"
)

func FromConfig(conf *config.TradingConfig, s *investsdk.SDK) (TradingStrategy, error) {
	switch conf.Strategy.Name {
	case "mov_avg":
		var movAvg TradingStrategy = NewMovingAvgStrategy(conf, s)
		return movAvg, nil
	default:
		return nil, xerrors.Errorf("no strategy with name %v", conf.Strategy)
	}
}
