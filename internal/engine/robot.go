package engine

import (
	"tinkoff-invest-bot/internal/strategies"

	"golang.org/x/xerrors"

	"tinkoff-invest-bot/internal/config"
	"tinkoff-invest-bot/pkg/sdk"
)

type investRobot struct {
	config   *config.Config
	strategy strategies.TradingStrategy
}

func New(conf *config.Config, tradingConf *config.TradingConfig) (*investRobot, error) {
	s, err := sdk.New(conf.TinkoffApiEndpoint, conf.AccessToken)
	if err != nil {
		return nil, xerrors.Errorf("can't init sdk: %v", err)
	}

	strategy, err := strategies.FromConfig(tradingConf, s)
	if err != nil {
		return nil, err
	}

	return &investRobot{
		config:   conf,
		strategy: strategy,
	}, nil
}

func (r *investRobot) Run() error {
	err := r.strategy.Start()
	if err != nil {
		return xerrors.Errorf("can't start robot strategy")
	}

	err = r.strategy.Step()
	if err != nil {
		return xerrors.Errorf("can't step robot strategy")
	}

	err = r.strategy.Stop()
	if err != nil {
		return xerrors.Errorf("can't stop robot strategy")
	}
	return nil
}
