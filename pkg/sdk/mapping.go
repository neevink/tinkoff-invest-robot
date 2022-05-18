package sdk

import (
	"time"

	"github.com/sdcoffey/big"
	"github.com/sdcoffey/techan"

	"tinkoff-invest-bot/investapi"
)

func QuotationToFloat(q *investapi.Quotation) float64 {
	return float64(q.Units) + float64(q.Nano)/1000000000
}

func MoneyValueToFloat(q *investapi.MoneyValue) float64 {
	return float64(q.Units) + float64(q.Nano)/1000000000
}

func HistoricCandleToCandle(c *investapi.HistoricCandle, period time.Duration) *techan.Candle {
	timePeriod := techan.NewTimePeriod(c.Time.AsTime(), period)
	candle := techan.NewCandle(timePeriod)

	candle.OpenPrice = big.NewDecimal(QuotationToFloat(c.Open))
	candle.ClosePrice = big.NewDecimal(QuotationToFloat(c.Close))
	candle.MaxPrice = big.NewDecimal(QuotationToFloat(c.High))
	candle.MinPrice = big.NewDecimal(QuotationToFloat(c.Low))
	candle.Volume = big.NewFromInt(int(c.Volume))
	return candle
}

func CandleToCandle(c *investapi.Candle, period time.Duration) *techan.Candle {
	timePeriod := techan.NewTimePeriod(c.Time.AsTime(), period)
	candle := techan.NewCandle(timePeriod)

	candle.OpenPrice = big.NewDecimal(QuotationToFloat(c.Open))
	candle.ClosePrice = big.NewDecimal(QuotationToFloat(c.Close))
	candle.MaxPrice = big.NewDecimal(QuotationToFloat(c.High))
	candle.MinPrice = big.NewDecimal(QuotationToFloat(c.Low))
	candle.Volume = big.NewFromInt(int(c.Volume))
	return candle
}
