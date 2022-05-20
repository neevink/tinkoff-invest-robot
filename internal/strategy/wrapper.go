package strategy

import (
	"github.com/sdcoffey/big"
	"github.com/sdcoffey/techan"
	"go.uber.org/zap"

	"tinkoff-invest-bot/internal/config"
	"tinkoff-invest-bot/investapi"
	"tinkoff-invest-bot/pkg/sdk"
)

type FinishEvent struct{}

type Wrapper struct {
	tradingConfig *config.TradingConfig
	sdk           *sdk.SDK
	logger        *zap.Logger

	TimeSeries    *techan.TimeSeries
	TradingRecord *techan.TradingRecord
	ruleStrategy  *techan.RuleStrategy

	blockChannel chan FinishEvent
}

func (w Wrapper) Step(candle *techan.Candle) Operation {
	w.TimeSeries.AddCandle(candle) // добавляем пришедшую свечу (неважно откуда)

	if w.ruleStrategy.ShouldEnter(w.TimeSeries.LastIndex(), w.TradingRecord) {
		return Buy
	} else if w.ruleStrategy.ShouldExit(w.TimeSeries.LastIndex(), w.TradingRecord) {
		return Sell
	} else {
		return Hold
	}
}

func (w Wrapper) Consume(data *investapi.MarketDataResponse) {
	if data == nil {
		return
	}

	if data.GetCandle() == nil {
		return
	}

	op := w.Step(CandleToCandle(data.GetCandle(), sdk.IntervalToDuration(w.tradingConfig.StrategyConfig.Interval)))
	// TODO тут можно сократить текст на много
	switch op {
	case Buy:
		orderId := sdk.GenerateOrderId()

		var resp *investapi.PostOrderResponse
		var trackingId string
		var err error
		if w.tradingConfig.IsSandbox {
			resp, trackingId, err = w.sdk.SandboxMarketBuy(
				w.tradingConfig.Figi,
				w.tradingConfig.StrategyConfig.Quantity,
				w.tradingConfig.AccountId,
				orderId,
			)
		} else {
			resp, trackingId, err = w.sdk.RealMarketBuy(
				w.tradingConfig.Figi,
				w.tradingConfig.StrategyConfig.Quantity,
				w.tradingConfig.AccountId,
				orderId,
			)
		}

		if err != nil {
			w.logger.Info(
				"Can't Buy share",
				zap.String("accountId", w.tradingConfig.AccountId),
				zap.Bool("isSandbox", w.tradingConfig.IsSandbox),
				zap.String("figi", w.tradingConfig.Figi),
				zap.Float64("price", sdk.MoneyValueToFloat(resp.GetInitialOrderPrice())),
				zap.String("ruleStrategy", w.tradingConfig.StrategyConfig.Name),
				zap.String("orderId", orderId),
				zap.String("trackingId", trackingId),
				zap.Error(err),
			)
		} else {
			w.TradingRecord.Operate(techan.Order{
				Side:          techan.OrderSide(op),
				Security:      orderId,
				Price:         big.NewDecimal(sdk.MoneyValueToFloat(resp.GetExecutedOrderPrice())),
				Amount:        big.NewFromInt(int(resp.GetLotsExecuted())),
				ExecutionTime: w.TimeSeries.LastCandle().Period.End,
			})

			w.logger.Info(
				"Buy new share",
				zap.String("accountId", w.tradingConfig.AccountId),
				zap.Bool("isSandbox", w.tradingConfig.IsSandbox),
				zap.String("figi", w.tradingConfig.Figi),
				zap.Float64("price", sdk.MoneyValueToFloat(resp.GetExecutedOrderPrice())),
				zap.String("ruleStrategy", w.tradingConfig.StrategyConfig.Name),
				zap.String("orderId", orderId),
				zap.String("trackingId", trackingId),
			)
		}
	case Sell:
		orderId := sdk.GenerateOrderId()

		var resp *investapi.PostOrderResponse
		var trackingId string
		var err error
		if w.tradingConfig.IsSandbox {
			resp, trackingId, err = w.sdk.SandboxMarketSell(
				w.tradingConfig.Figi,
				w.tradingConfig.StrategyConfig.Quantity,
				w.tradingConfig.AccountId,
				orderId,
			)
		} else {
			resp, trackingId, err = w.sdk.RealMarketSell(
				w.tradingConfig.Figi,
				w.tradingConfig.StrategyConfig.Quantity,
				w.tradingConfig.AccountId,
				orderId,
			)
		}

		if err != nil {
			w.logger.Info(
				"Can't sell new share",
				zap.String("accountId", w.tradingConfig.AccountId),
				zap.Bool("isSandbox", w.tradingConfig.IsSandbox),
				zap.String("figi", w.tradingConfig.Figi),
				zap.Float64("price", sdk.MoneyValueToFloat(resp.GetInitialOrderPrice())),
				zap.String("ruleStrategy", w.tradingConfig.StrategyConfig.Name),
				zap.String("orderId", orderId),
				zap.String("trackingId", trackingId),
				zap.Error(err),
			)
		} else {
			w.TradingRecord.Operate(techan.Order{
				Side:          techan.OrderSide(op),
				Security:      orderId,
				Price:         big.NewDecimal(sdk.MoneyValueToFloat(resp.GetExecutedOrderPrice())),
				Amount:        big.NewFromInt(int(resp.GetLotsExecuted())),
				ExecutionTime: w.TimeSeries.LastCandle().Period.End,
			})

			w.logger.Info(
				"Sell share",
				zap.String("accountId", w.tradingConfig.AccountId),
				zap.Bool("isSandbox", w.tradingConfig.IsSandbox),
				zap.String("figi", w.tradingConfig.Figi),
				zap.Float64("price", sdk.MoneyValueToFloat(resp.GetExecutedOrderPrice())),
				zap.String("ruleStrategy", w.tradingConfig.StrategyConfig.Name),
				zap.String("orderId", orderId),
				zap.String("trackingId", trackingId),
			)
			// w.blockChannel <- FinishEvent{}
		}
	case Hold:
		w.logger.Info(
			"Algorithm is waiting",
			zap.String("figi", w.tradingConfig.Figi),
			zap.Float64("curPrice", w.TimeSeries.LastCandle().ClosePrice.Float()),
			zap.String("ruleStrategy", w.tradingConfig.StrategyConfig.Name),
		)
	default:
	}
}

func (w Wrapper) Start() error {
	var cons sdk.TickerPriceConsumerInterface = w
	err := w.sdk.SubscribeMarketData(w.tradingConfig.Figi, sdk.IntervalToSubscriptionInterval(w.tradingConfig.StrategyConfig.Interval), &cons)
	if err != nil {
		return err
	}

	w.logger.Info(
		"Algorithm started",
		zap.String("figi", w.tradingConfig.Figi),
		zap.String("ruleStrategy", w.tradingConfig.StrategyConfig.Name),
	)

	return nil
}

func (w Wrapper) Stop() error {
	var cons sdk.TickerPriceConsumerInterface = w
	if err := w.sdk.UnsubscribeMarketData(w.tradingConfig.Figi, &cons); err != nil {
		return err
	}
	w.logger.Info(
		"Algorithm stopped",
		zap.String("figi", w.tradingConfig.Figi),
		zap.String("ruleStrategy", w.tradingConfig.StrategyConfig.Name),
	)
	return nil
}

func (w *Wrapper) BlockUntilEnd() {
	<-w.blockChannel
}
