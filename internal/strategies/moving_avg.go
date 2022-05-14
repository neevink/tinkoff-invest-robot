package strategies

import (
	"fmt"
	"math"

	"tinkoff-invest-bot/internal/config"
	api "tinkoff-invest-bot/investapi"
	investsdk "tinkoff-invest-bot/pkg/sdk"
)

type operation uint64

const (
	buy operation = iota
	sell
)

type MovingAvgStrategy struct {
	tradingConf   *config.TradingConfig
	thresholdPerc float64 // percents
	sdk           *investsdk.SDK

	startPrice    float64
	nextOperation operation
}

func NewMovingAvgStrategy(tradingConf *config.TradingConfig, s *investsdk.SDK) *MovingAvgStrategy {
	return &MovingAvgStrategy{
		startPrice:    0,
		thresholdPerc: 1,
		tradingConf:   tradingConf,
		sdk:           s,

		nextOperation: buy,
	}
}

func (a *MovingAvgStrategy) Name() string {
	return "Moving average"
}

func (a *MovingAvgStrategy) Consume(data *api.MarketDataResponse) {
	lastPrice := data.GetLastPrice()
	if lastPrice == nil {
		return
	}

	if lastPrice.Figi != a.tradingConf.Figi {
		return
	}

	price := investsdk.QuotationToFloat(lastPrice.GetPrice())

	if a.nextOperation == buy && price < a.startPrice && math.Abs(a.startPrice-price)/a.startPrice > a.thresholdPerc {
		_, err := a.sdk.PostSandboxMarketOrder(
			a.tradingConf.Figi,
			1,
			true,
			a.tradingConf.AccountId,
		)
		if err != nil {
			a.nextOperation = sell
		}
	}

	if a.nextOperation == sell && price > a.startPrice && math.Abs(a.startPrice-price)/a.startPrice > a.thresholdPerc {
		_, err := a.sdk.PostSandboxMarketOrder(
			a.tradingConf.Figi,
			1,
			false,
			a.tradingConf.AccountId,
		)
		if err != nil {
			a.nextOperation = buy
		}
	}

	fmt.Printf("lastPrice figi %s is: %f\n", lastPrice.GetFigi(), investsdk.QuotationToFloat(lastPrice.GetPrice()))

}

func (a *MovingAvgStrategy) Start() error {
	fmt.Printf("Robot started\n")

	var cons investsdk.TickerPriceConsumerInterface = a
	err := a.sdk.SubscribeMarketData(a.tradingConf.Figi, &cons)
	if err != nil {
		return err
	}

	price, err := a.sdk.GetLastPrice(a.tradingConf.Figi)
	if err != nil {
		return err
	}
	a.startPrice = investsdk.QuotationToFloat(price.GetPrice())
	return nil
}

func (a *MovingAvgStrategy) Stop() error {
	var cons investsdk.TickerPriceConsumerInterface = a
	if err := a.sdk.UnsubscribeMarketData(a.tradingConf.Figi, &cons); err != nil {
		return err
	}
	fmt.Printf("Robot stopped\n")
	return nil
}
