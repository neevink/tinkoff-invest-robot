package sdk

import (
	"golang.org/x/xerrors"

	api "tinkoff-invest-bot/investapi"
)

func (s *SDK) SubscribeMarketData(figi string, interval api.SubscriptionInterval, consumer *TickerPriceConsumerInterface) error {
	consumers, contains := s.marketDataConsumers[figi]
	if !contains {
		subscribeRequest := api.MarketDataRequest{
			Payload: &api.MarketDataRequest_SubscribeCandlesRequest{
				SubscribeCandlesRequest: &api.SubscribeCandlesRequest{
					SubscriptionAction: api.SubscriptionAction_SUBSCRIPTION_ACTION_SUBSCRIBE,
					Instruments: []*api.CandleInstrument{
						{
							Figi:     figi,
							Interval: interval,
						},
					},
				},
			},
		}
		if err := s.marketDataStreamClient.Send(&subscribeRequest); err != nil {
			return err
		}
		s.marketDataConsumers[figi] = make([]*TickerPriceConsumerInterface, 0)
	}

	s.marketDataConsumers[figi] = append(consumers, consumer)
	return nil
}

func (s *SDK) UnsubscribeMarketData(figi string, consumer *TickerPriceConsumerInterface) error {
	consumers, contains := s.marketDataConsumers[figi]
	if !contains {
		return xerrors.Errorf("no such consumer subscribed on figi %s", figi)
	}

	for i, c := range consumers {
		if c == consumer {
			s.marketDataConsumers[figi] = append(consumers[:i], consumers[i+1:]...)
			break
		}
	}

	if len(consumers) == 0 {
		unsubscribeRequest := api.MarketDataRequest{
			Payload: &api.MarketDataRequest_SubscribeCandlesRequest{
				SubscribeCandlesRequest: &api.SubscribeCandlesRequest{
					SubscriptionAction: api.SubscriptionAction_SUBSCRIPTION_ACTION_UNSUBSCRIBE,
					Instruments: []*api.CandleInstrument{
						{Figi: figi},
					},
				},
			},
		}
		err := s.marketDataStreamClient.Send(&unsubscribeRequest)
		if err != nil {
			return err
		}
		// add unsubscribe grpc request
		delete(s.marketDataConsumers, figi)
	}
	return nil
}
