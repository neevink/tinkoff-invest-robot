package strategy

import (
	"github.com/sdcoffey/big"
	"github.com/sdcoffey/techan"
	"go.uber.org/zap"

	"tinkoff-invest-bot/internal/config"
	"tinkoff-invest-bot/investapi"
	"tinkoff-invest-bot/pkg/sdk"
)

type Wrapper struct {
	tradingConfig *config.TradingConfig
	sdk           *sdk.SDK
	logger        *zap.Logger

	TimeSeries    *techan.TimeSeries
	TradingRecord *techan.TradingRecord
	ruleStrategy  *techan.RuleStrategy
}

func (W Wrapper) Step(candle *techan.Candle) Operation {
	W.TimeSeries.AddCandle(candle) // добавляем пришедшую свечу (неважно откуда)

	if W.ruleStrategy.ShouldEnter(W.TimeSeries.LastIndex(), W.TradingRecord) {
		return Buy
	} else if W.ruleStrategy.ShouldExit(W.TimeSeries.LastIndex(), W.TradingRecord) {
		return Sell
	} else {
		return Hold
	}
}

func (W Wrapper) Consume(data *investapi.MarketDataResponse) {
	op := W.Step(CandleToCandle(data.GetCandle(), sdk.IntervalToDuration(W.tradingConfig.StrategyConfig.Interval)))
	// TODO тут можно сократить текст на много
	switch op {
	case Buy:
		orderId := sdk.GenerateOrderId()

		if resp, trackingId, err := W.sdk.SandboxMarketBuy(
			W.tradingConfig.Figi,
			W.tradingConfig.StrategyConfig.Quantity,
			W.tradingConfig.AccountId,
			orderId,
		); err != nil {
			W.logger.Info(
				"Can't Buy share",
				zap.String("figi", W.tradingConfig.Figi),
				zap.Float64("price", sdk.MoneyValueToFloat(resp.GetInitialOrderPrice())),
				zap.String("ruleStrategy", W.tradingConfig.StrategyConfig.Name),
				zap.String("orderId", orderId),
				zap.String("trackingId", trackingId),
				zap.Error(err),
			)
		} else {
			W.TradingRecord.Operate(techan.Order{
				Side:          techan.OrderSide(op),
				Security:      orderId,
				Price:         big.NewDecimal(sdk.MoneyValueToFloat(resp.GetExecutedOrderPrice())),
				Amount:        big.NewFromInt(int(resp.GetLotsExecuted())),
				ExecutionTime: W.TimeSeries.LastCandle().Period.End,
			})

			W.logger.Info(
				"Buy new share",
				zap.String("figi", W.tradingConfig.Figi),
				zap.Float64("price", sdk.MoneyValueToFloat(resp.GetExecutedOrderPrice())),
				zap.String("ruleStrategy", W.tradingConfig.StrategyConfig.Name),
				zap.String("orderId", orderId),
				zap.String("trackingId", trackingId),
			)
		}
	case Sell:
		orderId := sdk.GenerateOrderId()

		if resp, trackingId, err := W.sdk.SandboxMarketSell(
			W.tradingConfig.Figi,
			W.tradingConfig.StrategyConfig.Quantity,
			W.tradingConfig.AccountId,
			orderId,
		); err != nil {
			W.logger.Info(
				"Can't Sell new share",
				zap.String("figi", W.tradingConfig.Figi),
				zap.Float64("price", sdk.MoneyValueToFloat(resp.GetInitialOrderPrice())),
				zap.String("ruleStrategy", W.tradingConfig.StrategyConfig.Name),
				zap.String("orderId", orderId),
				zap.String("trackingId", trackingId),
				zap.Error(err),
			)
		} else {
			W.TradingRecord.Operate(techan.Order{
				Side:          techan.OrderSide(op),
				Security:      orderId,
				Price:         big.NewDecimal(sdk.MoneyValueToFloat(resp.GetExecutedOrderPrice())),
				Amount:        big.NewFromInt(int(resp.GetLotsExecuted())),
				ExecutionTime: W.TimeSeries.LastCandle().Period.End,
			})

			W.logger.Info(
				"Sell share",
				zap.String("figi", W.tradingConfig.Figi),
				zap.Float64("price", sdk.MoneyValueToFloat(resp.GetExecutedOrderPrice())),
				zap.String("ruleStrategy", W.tradingConfig.StrategyConfig.Name),
				zap.String("orderId", orderId),
				zap.String("trackingId", trackingId),
			)
		}
	case Hold:
		W.logger.Info(
			"Share ждет",
			zap.String("figi", W.tradingConfig.Figi),
			zap.Float64("curPrice", W.TimeSeries.LastCandle().ClosePrice.Float()),
			zap.String("ruleStrategy", W.tradingConfig.StrategyConfig.Name),
		)
	default:
		panic("нет такого")
	}
}

func (W Wrapper) Start() error {
	var cons sdk.TickerPriceConsumerInterface = W
	err := W.sdk.SubscribeCandles(W.tradingConfig.Figi, sdk.IntervalToSubscriptionInterval(W.tradingConfig.StrategyConfig.Interval), &cons)
	if err != nil {
		return err
	}

	W.logger.Info(
		"Algorithm started",
		zap.String("figi", W.tradingConfig.Figi),
		zap.String("ruleStrategy", W.tradingConfig.StrategyConfig.Name),
	)

	return nil
}

func (W Wrapper) Stop() error {
	var cons sdk.TickerPriceConsumerInterface = W
	if err := W.sdk.UnsubscribeCandles(W.tradingConfig.Figi, &cons); err != nil {
		return err
	}
	W.logger.Info(
		"Invest robot stopped",
		zap.String("figi", W.tradingConfig.Figi),
		zap.String("ruleStrategy", W.tradingConfig.StrategyConfig.Name),
	)
	return nil
}
