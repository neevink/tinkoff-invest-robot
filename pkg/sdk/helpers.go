package sdk

import (
	"fmt"
	"tinkoff-invest-bot/investapi"
)

func PrintQuotation(q *investapi.Quotation) {
	fmt.Printf("%d.%d", q.Units, q.Nano)
}

func PrintMoneyValue(q *investapi.MoneyValue) {
	fmt.Printf("%d.%d%s", q.Units, q.Nano, q.Currency)
}
