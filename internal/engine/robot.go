package engine

import (
	"context"
	"time"

	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"tinkoff-invest-bot/internal/config"
	"tinkoff-invest-bot/internal/strategy"
	"tinkoff-invest-bot/pkg/sdk"
)

type investRobot struct {
	robotConfig     *config.RobotConfig
	tradingConfig   *config.TradingConfig
	tradingStrategy *strategy.Wrapper
	logger          *zap.Logger
}

func New(conf *config.RobotConfig, tradingConfig *config.TradingConfig, logger *zap.Logger, ctx context.Context) (*investRobot, error) {
	s, err := sdk.New(conf.TinkoffApiEndpoint, conf.TinkoffAccessToken, conf.AppName, ctx)
	if err != nil {
		return nil, xerrors.Errorf("can't init sdk: %v", err)
	}
	s.Run()

	tradingStrategy, err := strategy.FromConfig(tradingConfig, s, logger)
	if err != nil {
		return nil, err
	}

	return &investRobot{
		robotConfig:     conf,
		tradingConfig:   tradingConfig,
		tradingStrategy: tradingStrategy,
		logger:          logger,
	}, nil
}

func (r *investRobot) Run() error {
	err := (*r.tradingStrategy).Start()
	if err != nil {
		return xerrors.Errorf("can't start robot tradingStrategy, %v", err)
	}

	r.logger.Info(
		"Invest robot successfully run",
		zap.String("figi", r.tradingConfig.Figi),
		zap.String("tradingStrategy", r.tradingConfig.StrategyConfig.Name),
	)

	time.Sleep(6000 * time.Second)

	err = (*r.tradingStrategy).Stop()
	if err != nil {
		return xerrors.Errorf("can't stop robot tradingStrategy, %v", err)
	}
	return nil
}
