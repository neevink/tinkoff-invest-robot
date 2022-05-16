package strategies

import (
	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"tinkoff-invest-bot/internal/config"
	investsdk "tinkoff-invest-bot/pkg/sdk"
)

func FromConfig(conf *config.TradingConfig, s *investsdk.SDK, logger *zap.Logger) (*TradingStrategy, error) {
	switch conf.Strategy.Name {
	case "simple":
		return NewSimpleStrategy(conf, s, logger), nil
	default:
		return nil, xerrors.Errorf("no strategy with name %v", conf.Strategy)
	}
}
