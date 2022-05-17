package sdk

import (
	"fmt"
	"log"
	"math/rand"
	"time"
	
	"tinkoff-invest-bot/investapi"
)

func ConvertIntervalToCandleInterval(s string) investapi.CandleInterval {
	switch s {
	case "1_MIN":
		return investapi.CandleInterval_CANDLE_INTERVAL_1_MIN
	case "5_MIN":
		return investapi.CandleInterval_CANDLE_INTERVAL_5_MIN
	case "15_MIN":
		return investapi.CandleInterval_CANDLE_INTERVAL_15_MIN
	case "HOUR":
		return investapi.CandleInterval_CANDLE_INTERVAL_HOUR
	case "DAY":
		return investapi.CandleInterval_CANDLE_INTERVAL_DAY
	default:
		log.Fatalf("Значение \"%s\" для интервала свечи не определено, исправьте конфигурации", s)
		return investapi.CandleInterval_CANDLE_INTERVAL_UNSPECIFIED
	}
}

func ConvertIntervalToDuration(s string) time.Duration {
	switch s {
	case "1_MIN":
		return time.Minute
	case "5_MIN":
		return time.Minute * 5
	case "15_MIN":
		return time.Minute * 15
	case "HOUR":
		return time.Hour
	case "DAY":
		return time.Hour * 24
	default:
		log.Fatalf("Значение \"%s\" для интервала свечи не определено, исправьте конфигурации", s)
		return 0
	}
}

func ConvertMoneyValue(moneyValue *investapi.MoneyValue) float64 {
	return float64(moneyValue.Units) + float64(moneyValue.Nano)/1000000000
}

func ConvertQuotation(quotation *investapi.Quotation) float64 {
	return float64(quotation.Units) + float64(quotation.Nano)/1000000000
}

func PrintQuotation(q *investapi.Quotation) {
	fmt.Printf("%f", QuotationToFloat(q))
}

func PrintMoneyValue(q *investapi.MoneyValue) {
	fmt.Printf("%f%s", MoneyValueToFloat(q), q.Currency)
}

func GenerateOrderId() string {
	const length = 36
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

	var random = rand.New(rand.NewSource(time.Now().UnixNano()))

	orderId := make([]byte, length)
	for i := range orderId {
		orderId[i] = charset[random.Intn(len(charset))]
	}
	return string(orderId)
}
