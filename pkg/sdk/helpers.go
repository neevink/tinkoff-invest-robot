package sdk

import (
	"log"
	"math/rand"
	"time"

	"tinkoff-invest-bot/investapi"
)

func IntervalToSubscriptionInterval(s string) investapi.SubscriptionInterval {
	switch s {
	case "1_MIN":
		return investapi.SubscriptionInterval_SUBSCRIPTION_INTERVAL_ONE_MINUTE
	case "5_MIN":
		return investapi.SubscriptionInterval_SUBSCRIPTION_INTERVAL_FIVE_MINUTES
	default:
		log.Fatalf("Значение \"%s\" для интервала свечи не определено, исправьте конфигурации", s)
		return investapi.SubscriptionInterval_SUBSCRIPTION_INTERVAL_UNSPECIFIED
	}
}

func IntervalToCandleInterval(s string) investapi.CandleInterval {
	switch s {
	case "1_MIN":
		return investapi.CandleInterval_CANDLE_INTERVAL_1_MIN
	case "5_MIN":
		return investapi.CandleInterval_CANDLE_INTERVAL_5_MIN
	default:
		log.Fatalf("Значение \"%s\" для интервала свечи не определено, исправьте конфигурации", s)
		return investapi.CandleInterval_CANDLE_INTERVAL_UNSPECIFIED
	}
}

func IntervalToDuration(s string) time.Duration {
	switch s {
	case "1_MIN":
		return time.Minute
	case "5_MIN":
		return time.Minute * 5
	default:
		log.Fatalf("Значение \"%s\" для интервала свечи не определено, исправьте конфигурации", s)
		return 0
	}
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
