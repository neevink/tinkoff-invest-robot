package strategies

import (
	"fmt"

	"tinkoff-invest-bot/internal/config"
	investsdk "tinkoff-invest-bot/pkg/sdk"
)

type MovingAvgStrategy struct {
	avgPrice      float64
	tradingConf   *config.TradingConfig
	thresholdPerc float64 // percents
	sdk           *investsdk.SDK
}

func NewMovingAvgStrategy(tradingConf *config.TradingConfig, s *investsdk.SDK) *MovingAvgStrategy {
	return &MovingAvgStrategy{
		avgPrice:      0,
		thresholdPerc: 2, // 2%
		tradingConf:   tradingConf,
		sdk:           s,
	}
}

func (a *MovingAvgStrategy) Start() error {
	price, err := a.sdk.GetLastPrice(a.tradingConf.Figi)
	if err != nil {
		return err
	}
	fmt.Printf("Robot started\n")
	investsdk.PrintQuotation(price.Price)
	fmt.Printf("\n")
	return nil
}

func (a *MovingAvgStrategy) Step() error {
	fmt.Printf("Robot steped\n")
	return nil
}

func (a *MovingAvgStrategy) Stop() error {
	fmt.Printf("Robot stopped\n")
	return nil
}
