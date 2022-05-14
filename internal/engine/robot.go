package engine

import (
	"time"

	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"tinkoff-invest-bot/internal/config"
	"tinkoff-invest-bot/internal/strategies"
	"tinkoff-invest-bot/pkg/sdk"
)

type investRobot struct {
	config      *config.RobotConfig
	tradingConf *config.TradingConfig
	strategy    strategies.TradingStrategy
	logger      *zap.Logger
}

func New(conf *config.RobotConfig, tradingConf *config.TradingConfig, logger *zap.Logger) (*investRobot, error) {
	s, err := sdk.New(conf.TinkoffApiEndpoint, conf.TinkoffAccessToken)
	if err != nil {
		return nil, xerrors.Errorf("can't init sdk: %v", err)
	}

	if s.Run(); err != nil {
		return nil, xerrors.Errorf("can't init sdk: %v", err)
	}

	strategy, err := strategies.FromConfig(tradingConf, s)
	if err != nil {
		return nil, err
	}

	return &investRobot{
		config:      conf,
		tradingConf: tradingConf,
		strategy:    strategy,
		logger:      logger,
	}, nil
}

func (r *investRobot) Run() error {
	err := r.strategy.Start()
	if err != nil {
		return xerrors.Errorf("can't start robot strategy, %v", err)
	}

	r.logger.Info(
		"Invest robot successfully run",
		zap.String("figi", r.tradingConf.Figi),
		zap.String("strategy", r.strategy.Name()),
	)

	time.Sleep(60 * time.Second)

	err = r.strategy.Stop()
	if err != nil {
		return xerrors.Errorf("can't stop robot strategy, %v", err)
	}
	return nil
}
