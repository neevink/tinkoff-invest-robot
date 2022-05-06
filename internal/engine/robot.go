package engine

import (
	"context"
	"fmt"

	"golang.org/x/xerrors"

	"tinkoff-invest-bot/internal/config"
	"tinkoff-invest-bot/pkg/sdk"
)

type investRobot struct {
	config *config.Config
}

func New(config *config.Config) *investRobot {
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
