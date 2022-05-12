package sdk

import (
	"fmt"
	"math/rand"
	"time"
	"tinkoff-invest-bot/investapi"
)

func PrintQuotation(q *investapi.Quotation) {
	fmt.Printf("%d.%d", q.Units, q.Nano)
}

func PrintMoneyValue(q *investapi.MoneyValue) {
	fmt.Printf("%d.%d%s", q.Units, q.Nano, q.Currency)
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
