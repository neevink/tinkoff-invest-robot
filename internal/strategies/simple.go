package strategies

import (
	"go.uber.org/zap"

	"tinkoff-invest-bot/internal/config"
	api "tinkoff-invest-bot/investapi"
	investsdk "tinkoff-invest-bot/pkg/sdk"
)

type operation uint64

const (
	buy operation = iota
	sell
)

type SimpleStrategy struct {
	tradingConf   *config.TradingConfig
	thresholdPerc float64 // percents
	sdk           *investsdk.SDK
	logger        *zap.Logger

	startPrice    float64
	nextOperation operation
}

func NewSimpleStrategy(tradingConf *config.TradingConfig, s *investsdk.SDK, logger *zap.Logger) *TradingStrategy {
	var strategy TradingStrategy = &SimpleStrategy{
		startPrice:    0,
		thresholdPerc: 1,
		tradingConf:   tradingConf,
		sdk:           s,
		logger:        logger,

		nextOperation: buy,
	}
	return &strategy
}

func (a *SimpleStrategy) Name() string {
	return "Simple"
}

func (a *SimpleStrategy) Consume(data *api.MarketDataResponse) {
	lastPrice := data.GetLastPrice()
	if lastPrice == nil {
		return
	}

	if lastPrice.Figi != a.tradingConf.Figi {
		return
	}

	price := investsdk.QuotationToFloat(lastPrice.GetPrice())

	if a.nextOperation == buy && price < a.startPrice {
		orderId := investsdk.GenerateOrderId()

		_, trackingId, err := a.sdk.SandboxMarketBuy(
			a.tradingConf.Figi,
			1,
			a.tradingConf.AccountId,
			orderId,
		)

		if err == nil {
			a.nextOperation = sell
			a.logger.Info(
				"Buy new share",
				zap.String("figi", a.tradingConf.Figi),
				zap.Float64("price", price),
				zap.String("strategy", a.Name()),
				zap.String("orderId", orderId),
				zap.String("trackingId", trackingId),
			)
		} else {
			a.logger.Info(
				"Can't buy share",
				zap.String("figi", a.tradingConf.Figi),
				zap.Float64("price", price),
				zap.String("strategy", a.Name()),
				zap.String("orderId", orderId),
				zap.String("trackingId", trackingId),
				zap.Error(err),
			)
		}
	}

	if a.nextOperation == sell && price > a.startPrice {
		orderId := investsdk.GenerateOrderId()

		_, trackingId, err := a.sdk.SandboxMarketSell(
			a.tradingConf.Figi,
			1,
			a.tradingConf.AccountId,
			orderId,
		)

		if err == nil {
			a.nextOperation = buy
			a.logger.Info(
				"Sell share",
				zap.String("figi", a.tradingConf.Figi),
				zap.Float64("price", price),
				zap.String("strategy", a.Name()),
				zap.String("orderId", orderId),
				zap.String("trackingId", trackingId),
			)
		} else {
			a.logger.Info(
				"Can't sell new share",
				zap.String("figi", a.tradingConf.Figi),
				zap.Float64("price", price),
				zap.String("strategy", a.Name()),
				zap.String("orderId", orderId),
				zap.String("trackingId", trackingId),
				zap.Error(err),
			)
		}
	}

}

func (a *SimpleStrategy) Start() error {
	var cons investsdk.TickerPriceConsumerInterface = a
	err := a.sdk.SubscribeMarketData(a.tradingConf.Figi, &cons)
	if err != nil {
		return err
	}

	price, _, err := a.sdk.GetLastPrice(a.tradingConf.Figi)
	if err != nil {
		return err
	}
	a.startPrice = investsdk.QuotationToFloat(price.GetPrice())

	a.logger.Info(
		"Algorithm started",
		zap.String("figi", a.tradingConf.Figi),
		zap.String("strategy", a.Name()),
		zap.Float64("start price", a.startPrice),
	)

	return nil
}

func (a *SimpleStrategy) Stop() error {
	var cons investsdk.TickerPriceConsumerInterface = a
	if err := a.sdk.UnsubscribeMarketData(a.tradingConf.Figi, &cons); err != nil {
		return err
	}
	a.logger.Info(
		"Invest robot stopped",
		zap.String("figi", a.tradingConf.Figi),
		zap.String("strategy", a.Name()),
	)
	return nil
}
