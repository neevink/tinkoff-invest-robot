package engine

import (
	"context"
	"fmt"

	"golang.org/x/xerrors"

	"tinkoff-invest-bot/internal/robot"
	"tinkoff-invest-bot/pkg/sdk"
)

type investRobot struct {
	config *robot.RobotConfig
}

func New(config *robot.RobotConfig) *investRobot {
	return &investRobot{
		config: config,
	}
}

func (r *investRobot) Run(ctx context.Context, share string) error {
	s, err := sdk.New(r.config.TinkoffApiEndpoint, r.config.AccessToken)
	if err != nil {
		return xerrors.Errorf("can't init sdk: %v", err)
	}

	acc, err := s.GetAccounts()
	fmt.Printf("%v", acc)
	return nil
}
