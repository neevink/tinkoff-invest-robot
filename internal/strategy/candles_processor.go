package strategy

import (
	"fmt"

	"github.com/sdcoffey/big"
	"github.com/sdcoffey/techan"
	"go.uber.org/zap"

	"tinkoff-invest-bot/internal/config"
	"tinkoff-invest-bot/investapi"
	"tinkoff-invest-bot/pkg/sdk"
)

type FinishEvent struct{}

// CandlesStrategyProcessor запускалка всех стратегий, работа которых основана на свечках
type CandlesStrategyProcessor struct {
	tradingConfig *config.TradingConfig
	sdk           *sdk.SDK
	logger        *zap.Logger

	TimeSeries    *techan.TimeSeries
	TradingRecord *techan.TradingRecord
	ruleStrategy  *techan.RuleStrategy

	blockChannel chan FinishEvent
}

func (w CandlesStrategyProcessor) Step(candle *techan.Candle) Operation {
	if w.TimeSeries.AddCandle(candle) {
		fmt.Printf("%v %s: %f\n", w.TimeSeries.LastIndex(), w.tradingConfig.Ticker, candle.ClosePrice.Float())
	} // добавляем пришедшую свечу (неважно откуда)

	if w.ruleStrategy.ShouldEnter(w.TimeSeries.LastIndex(), w.TradingRecord) {
		return Buy
	} else if w.ruleStrategy.ShouldExit(w.TimeSeries.LastIndex(), w.TradingRecord) {
		return Sell
	} else {
		return Hold
	}
}

func (w CandlesStrategyProcessor) Consume(data *investapi.MarketDataResponse) {
	op := w.Step(
		CandleToCandle(
			data.GetCandle(),
			sdk.IntervalToDuration(w.tradingConfig.StrategyConfig.Interval),
		),
	)

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
				zap.String("figi", w.tradingConfig.Figi),
				zap.String("ticker", w.tradingConfig.Ticker),
				zap.Float64("price", sdk.MoneyValueToFloat(resp.GetInitialOrderPrice())),
				zap.Int("candleNo", w.TimeSeries.LastIndex()),
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
				Amount:        big.NewDecimal(sdk.MoneyValueToFloat(resp.GetTotalOrderAmount())),
				ExecutionTime: w.TimeSeries.LastCandle().Period.End,
			})

			w.logger.Info(
				"Buy new share",
				zap.String("accountId", w.tradingConfig.AccountId),
				zap.String("figi", w.tradingConfig.Figi),
				zap.String("ticker", w.tradingConfig.Ticker),
				zap.Float64("price", sdk.MoneyValueToFloat(resp.GetExecutedOrderPrice())),
				zap.Int("candleNo", w.TimeSeries.LastIndex()),
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
				zap.String("figi", w.tradingConfig.Figi),
				zap.String("ticker", w.tradingConfig.Ticker),
				zap.Float64("price", sdk.MoneyValueToFloat(resp.GetInitialOrderPrice())),
				zap.Int("candleNo", w.TimeSeries.LastIndex()),
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
				Amount:        big.NewDecimal(sdk.MoneyValueToFloat(resp.GetTotalOrderAmount())),
				ExecutionTime: w.TimeSeries.LastCandle().Period.End,
			})

			w.logger.Info(
				"Sell share",
				zap.String("accountId", w.tradingConfig.AccountId),
				zap.String("figi", w.tradingConfig.Figi),
				zap.String("ticker", w.tradingConfig.Ticker),
				zap.Float64("price", sdk.MoneyValueToFloat(resp.GetExecutedOrderPrice())),
				zap.Int("candleNo", w.TimeSeries.LastIndex()),
				zap.Float64("income", w.TradingRecord.LastTrade().ExitOrder().Amount.Sub(w.TradingRecord.LastTrade().EntranceOrder().Amount).Float()),
				zap.String("ruleStrategy", w.tradingConfig.StrategyConfig.Name),
				zap.String("orderId", orderId),
				zap.String("trackingId", trackingId),
			)
			// w.blockChannel <- FinishEvent{}
		}
	case Hold:
	default:
	}
}

func (w CandlesStrategyProcessor) Start() error {
	var cons sdk.MarketDataConsumer = w
	err := w.sdk.SubscribeCandles(w.tradingConfig.Figi, sdk.IntervalToSubscriptionInterval(w.tradingConfig.StrategyConfig.Interval), &cons)
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

func (w CandlesStrategyProcessor) Stop() error {
	var cons sdk.MarketDataConsumer = w
	if err := w.sdk.UnsubscribeCandles(w.tradingConfig.Figi, &cons); err != nil {
		return err
	}
	w.logger.Info(
		"Algorithm stopped",
		zap.String("figi", w.tradingConfig.Figi),
		zap.String("ruleStrategy", w.tradingConfig.StrategyConfig.Name),
	)
	return nil
}

func (w *CandlesStrategyProcessor) BlockUntilEnd() {
	<-w.blockChannel
}
