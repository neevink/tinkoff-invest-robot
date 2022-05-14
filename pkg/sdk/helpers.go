package sdk

import (
	"fmt"
	"math/rand"
	"time"
	"tinkoff-invest-bot/investapi"
)

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
