package strategy

import (
	"github.com/sdcoffey/big"
	"github.com/sdcoffey/techan"
	"time"
	"tinkoff-invest-bot/investapi"
	"tinkoff-invest-bot/pkg/sdk"
)

func HistoricCandleToCandle(c *investapi.HistoricCandle, period time.Duration) *techan.Candle {
	timePeriod := techan.NewTimePeriod(c.Time.AsTime(), period)
	candle := techan.NewCandle(timePeriod)

	candle.OpenPrice = big.NewDecimal(sdk.QuotationToFloat(c.Open))
	candle.ClosePrice = big.NewDecimal(sdk.QuotationToFloat(c.Close))
	candle.MaxPrice = big.NewDecimal(sdk.QuotationToFloat(c.High))
	candle.MinPrice = big.NewDecimal(sdk.QuotationToFloat(c.Low))
	candle.Volume = big.NewFromInt(int(c.Volume))
	return candle
}

func CandleToCandle(c *investapi.Candle, period time.Duration) *techan.Candle {
	timePeriod := techan.NewTimePeriod(c.Time.AsTime(), period)
	candle := techan.NewCandle(timePeriod)

	candle.OpenPrice = big.NewDecimal(sdk.QuotationToFloat(c.Open))
	candle.ClosePrice = big.NewDecimal(sdk.QuotationToFloat(c.Close))
	candle.MaxPrice = big.NewDecimal(sdk.QuotationToFloat(c.High))
	candle.MinPrice = big.NewDecimal(sdk.QuotationToFloat(c.Low))
	candle.Volume = big.NewFromInt(int(c.Volume))
	return candle
}