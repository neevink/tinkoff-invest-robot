package sdk

import (
	api "tinkoff-invest-bot/investapi"
)

type TickerPriceConsumerInterface interface {
	Consume(data *api.MarketDataResponse)
}
